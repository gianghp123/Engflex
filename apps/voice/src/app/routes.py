"""Non-session HTTP routes: an APIRouter mounted by whichever app serves them.

The production dispatcher (`server.py`) mounts the router behind the
internal secret; the dev runner (`bot.py`) mounts it without auth, matching
the runner's own open local surface. Route bodies and status semantics are
unchanged from the runner-app era.
"""

from fastapi import APIRouter, File, Form, HTTPException, UploadFile
from loguru import logger
from pydantic import ValidationError

from app.schemas import TranscriptCommandBody
from transcript.analyze import AnalysisRefused, AnalyzeRequest, analyze_turn

router = APIRouter()


@router.post("/analyze")
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


@router.post("/transcribe")
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


@router.post("/transcript")
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
        logger.exception("transcript command failed")
        raise HTTPException(status_code=500, detail="transcript command failed") from exc


def register_routes(app) -> None:
    """Mount the data routes on a FastAPI app.

    Idempotent: the dev runner and a test module (or several test modules
    sharing the runner's singleton app) must not duplicate the surface.

    The routes are appended eagerly rather than via ``include_router``:
    current FastAPI materializes includes as lazy ``_IncludedRouter``
    placeholders, so ``app.routes`` would hold no ``APIRoute`` entries and
    the bot wiring test (which pins ``/analyze`` as a served ``APIRoute``)
    would fail. Appending registers the same route objects with identical
    request handling.
    """
    paths = {getattr(route, "path", None) for route in app.routes}
    if {"/analyze", "/transcript", "/transcribe"} <= paths:
        return
    for route in router.routes:
        app.router.routes.append(route)
