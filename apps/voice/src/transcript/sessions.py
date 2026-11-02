"""Registry of live voice sessions, keyed by conversation id.

An HTTP route cannot reach a running pipeline: `bot()` holds the
`PipelineWorker` as a local, and pipecat's own `active_sessions` is a local
inside `main()` (`runner/run.py`). So `bot()` publishes a handle here and drops
it in a `finally`, and the transcript routes resolve it per request. A missing
entry means the session ended (or never existed) and the route answers 404.
"""

from dataclasses import dataclass

from pipecat.pipeline.worker import PipelineWorker

from transcript.capture import TurnCollector


@dataclass
class SessionControl:
    """Everything a transcript command needs to act on one live session.

    The aggregators are deliberately absent: the correction is a *transform* the
    aggregator applies to its own live message list, so a command never has to
    read or rebuild the context itself.
    """

    conversation_id: str
    worker: PipelineWorker
    collector: TurnCollector


_SESSIONS: dict[str, SessionControl] = {}


def register(conversation_id: str, control: SessionControl) -> None:
    """Publish a live session, replacing any stale entry for the same id."""
    _SESSIONS[conversation_id] = control


def get(conversation_id: str) -> SessionControl | None:
    return _SESSIONS.get(conversation_id)


def unregister(conversation_id: str) -> None:
    _SESSIONS.pop(conversation_id, None)


def active() -> list[SessionControl]:
    return list(_SESSIONS.values())
