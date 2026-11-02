"""Non-session HTTP routes mounted on the pipecat runner's FastAPI app.

The dev runner only serves session endpoints (`/start`, per-session WebRTC)
out of the box. Anything session-independent — like one-shot turn analysis —
is registered here via the runner's shared `app` singleton, which must be
imported before `pipecat.runner.run.main()` runs (see `bot.py`).
"""

from fastapi import File, Form, HTTPException, UploadFile
from loguru import logger

# Named for what it is: the runner's FastAPI singleton, not our app package.
from pipecat.runner.run import app as runner_app
from pydantic import ValidationError

from app.internal import InternalEngineMiddleware
from app.schemas import TranscriptCommandBody

runner_app.add_middleware(InternalEngineMiddleware)
from transcript.analyze import AnalysisRefused, AnalyzeRequest, analyze_turn


@runner_app.post("/analyze")
async def analyze(req: AnalyzeRequest):
    """Analyze one learner turn. Any failure surfaces as an HTTP error; Go
    maps non-2xx to a per-turn error and never stores partial feedback."""
    try:
        feedback = await analyze_turn(req)
    except AnalysisRefused as exc:
        logger.warning("analysis refused", reason=str(exc)[:200])
        raise HTTPException(status_code=422, detail="analysis refused") from exc
    except (ValueError, ValidationError) as exc:
        logger.warning("analysis reply unusable", error=str(exc)[:200])
        raise HTTPException(status_code=422, detail="analysis reply unusable") from exc
    except Exception as exc:
        logger.exception("analysis failed")
        raise HTTPException(status_code=500, detail="analysis failed") from exc
    return {"feedback": feedback.model_dump()}


# --- Transcript commands -----------------------------------------------------
# All three are learner-initiated and session-scoped. Each refuses with its own
# status (404 no live session, 409 nothing to act on, 422 unusable input) and
# lets anything unexpected become a 500, mirroring /analyze above.


@runner_app.post("/transcribe")
async def transcribe(
    conversationId: str = Form(...),
    audio: UploadFile = File(...),  # noqa: B008 - standard FastAPI idiom
):
    """Transcribe one modal re-speak. Stateless: no window, no frames.

    Refusals mirror /transcript's vocabulary: 404 no live session, 413 over
    the cap, 422 the audio held nothing usable, 502 Deepgram failed.
    """
    from transcript import prerecorded
    from transcript.command import require_session

    require_session(conversationId)
    raw = await audio.read()
    if len(raw) > prerecorded.MAX_AUDIO_BYTES:
        raise HTTPException(status_code=413, detail="audio too large")
    try:
        text = await prerecorded.transcribe_bytes(raw, audio.content_type or "")
    except prerecorded.EmptyTranscriptError as exc:
        raise HTTPException(status_code=422, detail=str(exc)) from exc
    except prerecorded.TranscribeRefusedError as exc:
        raise HTTPException(status_code=422, detail=str(exc)) from exc
    except prerecorded.TranscribeUnavailableError as exc:
        raise HTTPException(status_code=502, detail="transcription unavailable") from exc
    return {"text": text}


@runner_app.post("/transcript")
async def transcript(req: TranscriptCommandBody):
    """One learner-initiated transcript command.

    The three actions (review / send / dismiss) share the reviewed turn, so
    they are one endpoint. Each refuses with its own status: 404 no live
    session, 409 illegal for the current state, 422 unusable text.
    """
    from transcript.command import run_action

    try:
        return await run_action(req.conversationId, req.action, req.text)
    except HTTPException:
        raise
    except Exception as exc:
        logger.error("transcript command failed", error_type=type(exc).__name__,
                     conversation_id=req.conversationId)
        raise HTTPException(status_code=500, detail="transcript command failed") from exc
