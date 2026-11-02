"""Route-level tests for POST /analyze.

The handler is thin by design (parse -> analyze_turn -> shape), so these
pin the contract Go depends on: request fields accepted, response envelope,
and status codes for refusal vs. unusable reply vs. transport failure.
"""

import pytest
from fastapi.testclient import TestClient

# Named for what it is: the runner's FastAPI singleton, not our app package.
from pipecat.runner.run import app as runner_app
from pydantic import ValidationError

import app.routes  # noqa: F401 -- side effect: registers the routes
import transcript.analyze as analysis
from app import routes
from transcript.analyze import TurnFeedback


def _payload(**overrides):
    body = {
        "conversationId": "c1",
        "turnId": "t1",
        "text": "I go yesterday.",
        "context": [],
    }
    body.update(overrides)
    return body


def _feedback(**overrides: object) -> TurnFeedback:
    data: dict[str, object] = {
        "corrected": "I went yesterday.",
        "spans": [],
        "relevance": {"status": "relevant", "reason": None},
        "alternatives": {"language": None, "contextual": None},
        "tip": "past tense",
    }
    data.update(overrides)
    return TurnFeedback.model_validate(data)


def test_analyze_returns_feedback_envelope(monkeypatch):
    async def fake_analyze(req):
        assert req.turnId == "t1"
        assert req.text == "I go yesterday."
        return _feedback()

    monkeypatch.setattr(routes, "analyze_turn", fake_analyze)
    resp = TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post("/analyze", json=_payload())
    assert resp.status_code == 200, resp.text
    assert resp.json()["feedback"]["tip"] == "past tense"


def test_analyze_refusal_is_422(monkeypatch):
    async def fake_analyze(req):
        raise analysis.AnalysisRefused("declined")

    monkeypatch.setattr(routes, "analyze_turn", fake_analyze)
    resp = TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post("/analyze", json=_payload())
    assert resp.status_code == 422


def test_analyze_unusable_reply_is_422(monkeypatch):
    async def fake_analyze(req):
        raise ValueError("no JSON object")

    monkeypatch.setattr(routes, "analyze_turn", fake_analyze)
    resp = TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post("/analyze", json=_payload())
    assert resp.status_code == 422


def test_analyze_transport_failure_is_500(monkeypatch):
    async def fake_analyze(req):
        raise RuntimeError("connection reset")

    monkeypatch.setattr(routes, "analyze_turn", fake_analyze)
    resp = TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post("/analyze", json=_payload())
    assert resp.status_code == 500


def test_analyze_rejects_missing_text():
    resp = TestClient(runner_app, headers={"X-Internal-Secret": "test"}).post("/analyze", json={"conversationId": "c1", "turnId": "t1"})
    assert resp.status_code == 422


def test_validation_error_shape_is_rejected():
    with pytest.raises(ValidationError):
        # `spans`, `relevance` and `alternatives` are required by the schema,
        # which is the point of this test — model_validate keeps the type
        # checker happy.
        TurnFeedback.model_validate({"corrected": "a", "tip": "t"})
