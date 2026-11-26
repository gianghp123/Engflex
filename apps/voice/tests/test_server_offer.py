"""The offer proxy: PENDING -> LIVE, renegotiation, ICE patches, trust.

bot() itself is never run here: `server.run_session` is monkeypatched with
a parking fake so a session goes LIVE and stays LIVE for the assertions.
"""

import asyncio
import time

import pytest
from fastapi.testclient import TestClient
from server_fakes import FakeHandler

from config import Settings
from server import create_app

SECRET = "test-secret"


@pytest.fixture
def captured():
    return []


@pytest.fixture
def client(monkeypatch, captured):
    cfg = Settings.model_validate({})
    cfg.internal_secret = SECRET
    cfg.drain_grace_sec = 1  # fast teardown at TestClient exit

    async def fake_run_session(runner_args, *, registry):
        captured.append(runner_args)
        try:
            await asyncio.Event().wait()
        except asyncio.CancelledError:
            registry.end_live(runner_args.session_id, reason="cancelled")
            raise

    monkeypatch.setattr("server.run_session", fake_run_session)
    with TestClient(create_app(config=cfg, handler=FakeHandler())) as c:
        yield c


def _auth() -> dict:
    return {"X-Internal-Secret": SECRET}


def _start(client, conversation="c1") -> str:
    resp = client.post(
        "/start",
        json={
            "transport": "webrtc",
            "enableDefaultIceServers": True,
            "body": {
                "userId": "u1",
                "conversationId": conversation,
                "maxDuration": 300,
                "persona": {"name": "Amelia", "roleTitle": "Hiring director"},
            },
        },
        headers=_auth(),
    )
    assert resp.status_code == 200, resp.text
    return resp.json()["sessionId"]


def _offer_payload(**overrides) -> dict:
    payload = {"sdp": "v=0 offer", "type": "offer"}
    payload.update(overrides)
    return payload


def _wait_for(predicate, timeout: float = 2.0) -> None:
    """Bounded wait for the fire-and-forget session task to be observable.

    The offer callback starts run_session as an asyncio task; the HTTP
    response can return before the task's first step appends to `captured`.
    """
    deadline = time.time() + timeout
    while not predicate() and time.time() < deadline:
        time.sleep(0.01)
    assert predicate()


def test_unknown_session_is_404(client):
    resp = client.post("/sessions/nope/api/offer", json=_offer_payload(), headers=_auth())
    assert resp.status_code == 404


def test_pending_offer_goes_live_and_starts_the_pipeline(client, captured):
    session_id = _start(client)
    resp = client.post(
        f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth()
    )
    assert resp.status_code == 200, resp.text
    assert resp.json() == {"sdp": "v=0 answer", "type": "answer", "pc_id": "pc-1"}
    registry = client.app.state.registry
    assert registry.get_live(session_id) is not None
    assert registry.pending_count() == 0
    _wait_for(lambda: len(captured) == 1)
    assert captured[0].session_id == session_id


def test_request_data_cannot_override_the_start_body(client, captured):
    # Trust boundary: persona/scenario/learner come only from the /start body
    # Go wrote. Whatever the offer carries is discarded.
    session_id = _start(client)
    payload = _offer_payload(request_data={"persona": {"name": "Impostor"}})
    resp = client.post(f"/sessions/{session_id}/api/offer", json=payload, headers=_auth())
    assert resp.status_code == 200, resp.text
    _wait_for(lambda: len(captured) == 1)
    assert captured[0].body["persona"]["name"] == "Amelia"
    handler = client.app.state.handler
    assert handler.web_requests[0].request_data is None


def test_duplicate_offer_on_a_live_session_is_409(client):
    session_id = _start(client)
    first = client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    assert first.status_code == 200
    resp = client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    assert resp.status_code == 409


def test_restart_pc_renegotiates_a_live_session(client, captured):
    session_id = _start(client)
    assert (
        client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
        .status_code
        == 200
    )
    resp = client.post(
        f"/sessions/{session_id}/api/offer",
        json=_offer_payload(pc_id="pc-1", restart_pc=True),
        headers=_auth(),
    )
    assert resp.status_code == 200, resp.text
    # Still exactly one pipeline: the renegotiation answered without a
    # second bot run.
    _wait_for(lambda: len(captured) == 1)
    assert len(captured) == 1


def test_expired_pending_offer_is_404(client):
    session_id = _start(client)
    client.app.state.registry._pending[session_id].expires_at = time.time() - 1
    resp = client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    assert resp.status_code == 404


def test_offer_expiring_during_negotiation_closes_the_connection(client, captured):
    # The TTL can win the race with SDP negotiation: the callback must
    # refuse to go live and close the connection it was handed.
    session_id = _start(client)
    handler = client.app.state.handler

    def expire_now():
        client.app.state.registry._pending[session_id].expires_at = time.time() - 1

    handler.before_callback = expire_now
    resp = client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    assert resp.status_code == 200  # the handler still produced an answer
    assert client.app.state.registry.get_live(session_id) is None
    assert captured == []
    assert handler.last_connection is not None
    assert handler.last_connection.disconnected is True


def test_malformed_offer_body_is_400(client):
    session_id = _start(client)
    resp = client.post(
        f"/sessions/{session_id}/api/offer", json={"type": "offer"}, headers=_auth()
    )
    assert resp.status_code == 400
    resp = client.post(
        f"/sessions/{session_id}/api/offer", content=b"not json", headers=_auth()
    )
    assert resp.status_code == 400


def test_null_sdp_offer_is_400_and_keeps_the_lease(client):
    # A null sdp/type fails validation before negotiation: the pending lease
    # must survive, so a subsequent valid offer for the same session succeeds.
    session_id = _start(client)
    resp = client.post(
        f"/sessions/{session_id}/api/offer",
        json={"sdp": None, "type": "offer"},
        headers=_auth(),
    )
    assert resp.status_code == 400
    assert client.app.state.registry.get_pending(session_id) is not None
    resp = client.post(
        f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth()
    )
    assert resp.status_code == 200, resp.text


def test_failed_offer_releases_the_lease(client):
    # SDP negotiation failed: the pending lease must be released (spec §9),
    # so the same conversation can /start again and get a fresh session.
    session_id = _start(client)

    async def exploding(request, webrtc_connection_callback):
        raise RuntimeError("sdp negotiation failed")

    client.app.state.handler.handle_web_request = exploding
    with pytest.raises(RuntimeError, match="sdp negotiation failed"):
        client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    assert client.app.state.registry.get_pending(session_id) is None
    fresh = _start(client)
    assert fresh != session_id
    assert client.app.state.registry.pending_count() == 1


# --- ICE patches -------------------------------------------------------------


def test_patch_for_unknown_session_is_404(client):
    resp = client.patch(
        "/sessions/nope/api/offer", json={"pc_id": "pc-1", "candidates": []}, headers=_auth()
    )
    assert resp.status_code == 404


def test_patch_before_the_offer_is_409(client):
    session_id = _start(client)
    resp = client.patch(
        f"/sessions/{session_id}/api/offer",
        json={"pc_id": "pc-1", "candidates": []},
        headers=_auth(),
    )
    assert resp.status_code == 409


def test_patch_after_the_offer_routs_the_candidates(client):
    session_id = _start(client)
    client.post(f"/sessions/{session_id}/api/offer", json=_offer_payload(), headers=_auth())
    candidates = [{"candidate": "candidate:1 udp", "sdpMid": "0", "sdpMLineIndex": 0}]
    resp = client.patch(
        f"/sessions/{session_id}/api/offer",
        json={"pc_id": "pc-1", "candidates": candidates},
        headers=_auth(),
    )
    assert resp.status_code == 200, resp.text
    assert resp.json() == {"status": "success"}
    handler = client.app.state.handler
    assert handler.patch_requests[0].pc_id == "pc-1"
    assert handler.patch_requests[0].candidates[0].candidate == "candidate:1 udp"
