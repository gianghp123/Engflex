"""Route-level tests for POST /transcribe."""

from typing import Any, cast

from fastapi.testclient import TestClient
from pipecat.runner.run import app as runner_app

import app.routes  # noqa: F401 -- side effect: registers the routes
from transcript import sessions
from transcript.capture import TurnCollector
from transcript.sessions import SessionControl

CONV = "conv-1"


def _register() -> None:
    sessions.register(
        CONV,
        SessionControl(
            conversation_id=CONV,
            worker=cast(Any, object()),
            collector=TurnCollector(),
        ),
    )


def _post_transcribe(audio: bytes, mime: str):
    _register()
    client = TestClient(runner_app, headers={"X-Internal-Secret": "test"})
    return client.post(
        "/transcribe",
        data={"conversationId": CONV},
        files={"audio": ("retake.webm", audio, mime)},
    )


class _Alt:
    transcript = "i went yesterday"


def _heard(transcript: str):
    alt = _Alt()
    alt.transcript = transcript

    class _Channel:
        def __init__(self):
            self.alternatives = [alt]

    class _Results:
        def __init__(self):
            self.channels = [_Channel()]

    class _Heard:
        def __init__(self):
            self.results = _Results()

    return _Heard()


def test_unknown_session_is_404():
    client = TestClient(runner_app, headers={"X-Internal-Secret": "test"})
    resp = client.post(
        "/transcribe",
        data={"conversationId": "nope"},
        files={"audio": ("retake.webm", b"bytes", "audio/webm")},
    )
    assert resp.status_code == 404


def test_empty_transcript_is_422(monkeypatch):
    from transcript import prerecorded

    async def silence(audio, mime):
        return _heard("  ")

    monkeypatch.setattr(prerecorded, "_post", silence)
    assert _post_transcribe(b"bytes", "audio/webm").status_code == 422


def test_deepgram_failure_is_502(monkeypatch):
    from transcript import prerecorded

    async def down(audio, mime):
        raise prerecorded.TranscribeUnavailableError("down")

    monkeypatch.setattr(prerecorded, "_post", down)
    assert _post_transcribe(b"bytes", "audio/webm").status_code == 502


def test_happy_path_returns_the_sentence(monkeypatch):
    from transcript import prerecorded

    async def heard(audio, mime):
        return _heard("i went yesterday")

    monkeypatch.setattr(prerecorded, "_post", heard)
    resp = _post_transcribe(b"bytes", "audio/webm")
    assert resp.status_code == 200
    assert resp.json() == {"text": "i went yesterday"}
