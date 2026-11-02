"""The transcript command: one endpoint, three actions, one reviewed turn.

Everything here acts on a live session and is reached from an HTTP route that Go
proxies. There is one operation: **rewriting** the turn under review. The field
holds one sentence however the learner filled it — typed, or transcribed from
an isolated re-speak that never entered the pipeline — so both mean the same
thing: the turn's text is replaced, everything after it is dropped, and
inference re-runs so the tutor answers the new words.

The reviewed turn is tracked as a position on the collector, not as pipeline
state. `review` marks it, `send` rewrites it, `dismiss` forgets it.
"""

from typing import Any

from fastapi import HTTPException
from loguru import logger
from pipecat.frames.frames import (
    InterruptionFrame,
    LLMMessagesTransformFrame,
)

from transcript.correction import replace_last_user_message
from transcript.sessions import SessionControl, get

# A correction is a sentence or two, not an essay; the cap stops a paste from
# becoming a context bomb.
MAX_CORRECTION_CHARS = 2000

ACTIONS = ("review", "send", "dismiss")


def require_session(conversation_id: str) -> SessionControl:
    control = get(conversation_id)
    if control is None:
        raise HTTPException(status_code=404, detail="no live session")
    return control


# --- actions -----------------------------------------------------------------


async def _review(control: SessionControl) -> dict[str, Any]:
    """Hand back the latest learner turn and mark it as the one under review."""
    record = control.collector.latest_user_record()
    if record is None:
        raise HTTPException(status_code=409, detail="no user turn to review")
    control.collector.mark_reviewed(record.position)
    logger.info(f"transcript_reviewed position={record.position}")
    return {"state": "reviewing", "text": record.text}


async def _dismiss(control: SessionControl) -> dict[str, Any]:
    """Forget the review. Records nothing — the utterance was never kept."""
    control.collector.clear_reviewed()
    return {"state": "idle"}


async def _send(control: SessionControl, text: str) -> dict[str, Any]:
    cleaned = (text or "").strip()
    if not cleaned:
        raise HTTPException(status_code=422, detail="empty text")
    if len(cleaned) > MAX_CORRECTION_CHARS:
        raise HTTPException(status_code=422, detail="text too long")

    # There is one operation, not two. The field holds one sentence however it
    # was filled, so a transcribed re-speak and a typed edit both mean "this is
    # what I said", and both replace the turn under review.
    position = control.collector.reviewed_position
    if position is None:
        raise HTTPException(status_code=409, detail="no correction in progress")
    record = next(
        (r for r in control.collector.records() if r.position == position and r.role == "user"),
        None,
    )
    if record is None:
        raise HTTPException(status_code=409, detail="the reviewed turn is gone")
    if cleaned == record.text:
        # Re-sending what we already have must not spend an LLM run and a
        # spoken reply, so refuse instead of regenerating the same answer.
        raise HTTPException(status_code=422, detail="text unchanged")
    if control.collector.latest_user_record() is not record:
        # Something committed after the review. Refuse: the learner is looking
        # at one turn and we would rewrite another.
        raise HTTPException(status_code=409, detail="a newer turn was committed")
    return await _rewrite(control, cleaned)


# --- the operation ------------------------------------------------------------


async def _rewrite(control: SessionControl, cleaned: str) -> dict[str, Any]:
    """Rewrite the reviewed turn in place and regenerate the tutor's reply."""
    from pipecat.frames.frames import EndWorkerFrame

    from transcript.capture import SnapshotLimitExceeded
    try:
        position = control.collector.begin_correction(cleaned)
    except SnapshotLimitExceeded as exc:
        await control.worker.queue_frame(EndWorkerFrame())
        raise HTTPException(status_code=409 if control.collector.finalized else 413, detail=str(exc)) from exc
    if position is None:  # pragma: no cover - send resolved the record above
        raise HTTPException(status_code=409, detail="no user turn to correct")

    # The in-flight reply answered text that no longer exists. A no-op when the
    # tutor already finished.
    await control.worker.queue_frame(InterruptionFrame())
    # A transform, not a rebuilt list: pipecat calls it with the LIVE messages
    # when the frame is processed, so the rewrite cannot race a turn that
    # landed after we read the collector, and anything we do not touch (an
    # LLM-specific message, a future type) passes through untouched.
    await control.worker.queue_frame(
        LLMMessagesTransformFrame(transform=_correction_transform(cleaned), run_llm=True)
    )
    control.collector.clear_reviewed()
    logger.info(f"transcript_corrected position={position} chars={len(cleaned)}")
    return {"state": "idle", "text": cleaned}


def _correction_transform(corrected_text: str) -> Any:
    """The transform the aggregator applies to its own message list."""

    def transform(messages: list[Any]) -> list[Any]:
        rebuilt = replace_last_user_message(messages, corrected_text)
        # Nothing to correct: leave the history alone rather than truncating it.
        return messages if rebuilt is None else rebuilt

    return transform


# --- dispatch ----------------------------------------------------------------


async def run_action(conversation_id: str, action: str, text: str | None) -> dict[str, Any]:
    """Apply one command. Each action refuses with its own status: 404 no live
    session, 409 illegal for the current state, 422 unusable text."""
    if action not in ACTIONS:
        raise HTTPException(status_code=422, detail="unknown action")
    control = require_session(conversation_id)
    match action:
        case "review":
            return await _review(control)
        case "dismiss":
            return await _dismiss(control)
        case "send":
            return await _send(control, text or "")
    raise HTTPException(status_code=422, detail="unknown action")  # pragma: no cover
