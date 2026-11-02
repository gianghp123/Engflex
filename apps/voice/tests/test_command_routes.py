"""Route-level tests for the transcript command endpoint.

Every command acts on one live session, so each test publishes a real
`SessionControl`: a real `TurnCollector` and a cast for the worker, whose
behaviour is asserted through what it was handed.
"""

from typing import Any, cast

import pytest
from fastapi.testclient import TestClient
from pipecat.frames.frames import (
    LLMMessagesTransformFrame,
)

# Named for what it is: the runner's FastAPI singleton, not our app package.
from pipecat.runner.run import app as runner_app

import app.routes  # noqa: F401 -- side effect: registers the transcript route
from transcript import sessions
from transcript.capture import TurnCollector
from transcript.sessions import SessionControl

CONV = "conv-1"

MESSAGES = [
    {"role": "user", "content": "Hello!"},
    {"role": "user", "content": "i go yesterday"},
    {"role": "assistant", "content": "so you went"},
]
RECORDS = [("user", "i go yesterday"), ("ai", "so you went")]


class _FakeWorker:
    def __init__(self):
        self.frames: list[object] = []

    async def queue_frame(self, frame) -> None:
        self.frames.append(frame)


class _Harness:
    def __init__(self, control: SessionControl, worker: _FakeWorker):
        self.control = control
        self.worker = worker

    @property
    def collector(self) -> TurnCollector:
        return self.control.collector

    def texts(self) -> list[str]:
        return [r.text for r in self.collector.records()]

    def positions(self) -> list[int]:
        return [r.position for r in self.collector.records()]


def _install(*, records=(), session=True) -> _Harness:
    worker = _FakeWorker()
    collector = TurnCollector()
    for role, text in records:
        if role == "user":
            collector.add_user_text(text)
        else:
            collector.add_bot_text(text)
    control = SessionControl(
        # Any, not the concrete class: this stands in for the pipeline worker,
        # and the assertions read the fake directly.
        conversation_id=CONV,
        worker=cast(Any, worker),
        collector=collector,
    )
    if session:
        sessions.register(CONV, control)
    return _Harness(control, worker)


def _post(action: str, **body):
    return TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post(
        "/transcript", json={"conversationId": CONV, "action": action, **body}
    )


@pytest.fixture(autouse=True)
def _clean_registry():
    yield
    sessions.unregister(CONV)


@pytest.fixture(autouse=True)
def _capture_turn_posts(monkeypatch):
    """Observe copied snapshots at the mutation boundary, including corrections."""
    posted = []
    original = TurnCollector._changed
    def changed(self):
        original(self)
        posted.append(list(self.snapshot()[1]))
    monkeypatch.setattr(TurnCollector, "_changed", changed)
    return posted


# --- review ------------------------------------------------------------------


def test_review_marks_the_turn_and_hands_back_its_text():
    h = _install(records=RECORDS)
    resp = _post("review")
    assert resp.status_code == 200
    assert resp.json() == {"state": "reviewing", "text": "i go yesterday"}
    assert h.control.collector.reviewed_position == 1


def test_review_needs_a_turn_to_review():
    _install(records=[])
    assert _post("review").status_code == 409


def test_review_on_an_unknown_session_is_404():
    _install(records=RECORDS, session=False)
    resp = _post("review")
    assert resp.status_code == 404
    # Distinguish our refusal from FastAPI's router 404 for a missing route.
    assert resp.json()["detail"] == "no live session"


def test_every_action_needs_a_live_session():
    for action in ("review", "send", "dismiss"):
        _install(records=RECORDS, session=False)
        assert _post(action, text="x").status_code == 404, action


def test_send_is_refused_without_a_review():
    h = _install(records=RECORDS)
    assert _post("send", text="i went yesterday").status_code == 409
    assert h.worker.frames == []
    assert h.collector.reviewed_position is None


def test_dismiss_is_idempotent():
    # Closing something already closed is harmless, and the client benefits
    # from being able to dismiss defensively when its state is stale.
    h = _install(records=RECORDS)
    resp = _post("dismiss")
    assert resp.status_code == 200
    assert resp.json()["state"] == "idle"
    assert h.collector.reviewed_position is None


# --- send: the rewrite -------------------------------------------------------


def test_send_rewrites_the_reviewed_turn_and_regenerates():
    h = _install(records=RECORDS)
    assert _post("review").status_code == 200
    resp = _post("send", text="i went yesterday")
    assert resp.status_code == 200
    assert resp.json()["state"] == "idle"
    kinds = [type(f).__name__ for f in h.worker.frames]
    assert kinds == ["InterruptionFrame", "LLMMessagesTransformFrame"]
    transform_frame = cast(LLMMessagesTransformFrame, h.worker.frames[1])
    assert transform_frame.run_llm is True
    # The frame carries the edit as a transform over the aggregator's LIVE list,
    # so apply it to the context as pipecat would: the stale reply and the note
    # are both gone, and nothing before the target is touched.
    assert transform_frame.transform(cast("list[Any]", MESSAGES)) == [
        {"role": "user", "content": "Hello!"},
        {"role": "user", "content": "i went yesterday"},
    ]
    assert h.texts() == ["i went yesterday"]
    assert h.collector.reviewed_position is None


def test_send_refuses_when_a_newer_turn_was_committed():
    # Silently rewriting "whatever is newest" is the one failure this design
    # exists to prevent, so a turn that landed after the review refuses.
    h = _install(records=RECORDS)
    _post("review")
    h.collector.add_user_text("something newer")
    assert _post("send", text="i went yesterday").status_code == 409
    # Refused, so nothing was rewritten and no reply was generated.
    assert h.texts() == ["i go yesterday", "so you went", "something newer"]
    assert h.worker.frames == []


def test_send_refuses_unchanged_text():
    h = _install(records=RECORDS)
    _post("review")
    before = len(h.worker.frames)
    assert _post("send", text="i go yesterday").status_code == 422
    assert len(h.worker.frames) == before
    assert h.collector.reviewed_position == 1, "a refusal keeps the review open"


def test_send_refuses_empty_and_oversized_text():
    h = _install(records=RECORDS)
    _post("review")
    assert _post("send", text="   ").status_code == 422
    assert _post("send", text="x" * 2001).status_code == 422
    assert h.collector.reviewed_position == 1


def test_send_reposts_the_corrected_turn_for_go(_capture_turn_posts):
    # The turn was already posted when it completed, and Go upserts on
    # (conversation_id, position), so without the re-post the database would
    # keep the misheard text.
    _install(records=RECORDS)
    _post("review")
    assert _post("send", text="i went yesterday").status_code == 200
    (record,) = _capture_turn_posts[-1]
    assert record.position == 1
    assert record.text == "i went yesterday"


def test_correction_notifies_the_session_publisher():
    h = _install(records=RECORDS)
    notified = []
    h.collector.on_change = lambda: notified.append(h.collector.snapshot())
    _post("review")
    assert _post("send", text="i went yesterday").status_code == 200
    assert notified[-1][1][0].text == "i went yesterday"


def test_the_regenerated_reply_replaces_the_stale_one_on_the_reserved_slot():
    # The whole reason the reply slot is reserved: the tutor's answer to the
    # mishearing is overwritten, not left dangling before a second answer.
    h = _install(records=RECORDS)
    _post("review")
    _post("send", text="i went yesterday")
    h.collector.add_bot_text("you went to the market")
    assert h.texts() == ["i went yesterday", "you went to the market"]
    assert h.positions() == [1, 2]


def test_a_spoken_attempt_identical_to_the_turn_is_refused():
    # Transcribing the same words is not a correction. Refusing costs the
    # learner nothing: the text they want is already in the field.
    h = _install(records=RECORDS)
    _post("review")
    resp = _post("send", text="i go yesterday")
    assert resp.status_code == 422
    assert resp.json()["detail"] == "text unchanged"
    # Refused means refused: no turn rewritten, and the review stays open so
    # the learner can try again rather than being stranded.
    assert h.texts() == ["i go yesterday", "so you went"]
    assert h.collector.reviewed_position == 1


# --- dismiss -----------------------------------------------------------------


def test_dismiss_closes_the_review_without_recording_anything():
    h = _install(records=RECORDS)
    _post("review")
    resp = _post("dismiss")
    assert resp.status_code == 200
    assert resp.json()["state"] == "idle"
    assert h.texts() == ["i go yesterday", "so you went"]
    assert h.collector.reviewed_position is None


def test_dismiss_from_reviewing_records_nothing():
    h = _install(records=RECORDS)
    _post("review")
    assert _post("dismiss").json()["state"] == "idle"
    assert h.texts() == ["i go yesterday", "so you went"]
