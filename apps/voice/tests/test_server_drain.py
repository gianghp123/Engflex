"""Drain: the admission gate flips synchronously with the 202, live sessions
get their grace, and the process is asked to exit only after the drain
completes (the exit request is monkeypatched away here)."""

import asyncio
import time

import pytest
from fastapi.testclient import TestClient
from server_fakes import FakeHandler

import server
from config import Settings
from server import create_app

SECRET = "test-secret"


@pytest.fixture
def exit_calls(monkeypatch):
    calls: list[int] = []
    monkeypatch.setattr(server, "_request_process_exit", lambda: calls.append(1))
    return calls


@pytest.fixture
def stack(exit_calls, monkeypatch):
    cfg = Settings.model_validate({})
    cfg.internal_secret = SECRET
    cfg.drain_grace_sec = 1
    cfg.max_sessions = 4

    async def fake_run_session(runner_args, *, registry):
        try:
            await asyncio.Event().wait()
        except asyncio.CancelledError:
            registry.end_live(runner_args.session_id, reason="cancelled")
            raise

    monkeypatch.setattr(server, "run_session", fake_run_session)
    app = create_app(config=cfg, handler=FakeHandler())
    with TestClient(app) as client:
        yield client, app


def _auth() -> dict:
    return {"X-Internal-Secret": SECRET}


def _start(client, conversation) -> str:
    resp = client.post(
        "/start",
        json={
            "transport": "webrtc",
            "enableDefaultIceServers": False,
            "body": {"userId": "u1", "conversationId": conversation, "maxDuration": 300},
        },
        headers=_auth(),
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["sessionId"]


def _offer(client, session_id) -> None:
    resp = client.post(
        f"/sessions/{session_id}/api/offer",
        json={"sdp": "v=0 offer", "type": "offer"},
        headers=_auth(),
    )
    assert resp.status_code == 200, resp.text


def test_drain_refuses_new_starts_and_reports_counts(stack, exit_calls):
    client, app = stack
    session_id = _start(client, "c1")
    _offer(client, session_id)

    resp = client.post("/drain", headers=_auth())
    assert resp.status_code == 202, resp.text
    assert resp.json() == {"active": 1, "pending": 0, "graceSec": 1}

    # The gate flipped synchronously with the 202, not when a task got around
    # to it: an immediate /start must already be refused.
    refused = client.post(
        "/start",
        json={
            "transport": "webrtc",
            "enableDefaultIceServers": False,
            "body": {"userId": "u1", "conversationId": "c2", "maxDuration": 300},
        },
        headers=_auth(),
    )
    assert refused.status_code == 503
    assert refused.headers["Retry-After"] == "5"

    assert client.get("/ready").status_code == 503

    # The live session is cancelled once the (tiny, test) grace expires, the
    # lease released, and only then is the process asked to exit.
    deadline = time.time() + 10
    while not exit_calls and time.time() < deadline:
        time.sleep(0.05)
    assert exit_calls, "drain never completed"
    assert app.state.registry.live_count() == 0


def test_drain_requires_auth(stack):
    client, _ = stack
    assert client.post("/drain").status_code == 401


def test_ready_is_200_while_not_draining(stack):
    client, _ = stack
    assert client.get("/ready").status_code == 200


def test_status_reports_draining_and_counters(stack):
    client, _ = stack
    assert client.get("/status").json()["status"] == "ready"
    client.post("/drain", headers=_auth())
    body = client.get("/status").json()
    assert body["status"] == "draining"
    assert body["draining"] is True
    assert body["max"] == 4
