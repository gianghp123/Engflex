"""Capacity and admission over HTTP: MAX_SESSIONS, MAX_PENDING, idempotency,
validation, clamping, pending expiry."""

import time

from fastapi.testclient import TestClient
from server_fakes import FakeHandler

from config import Settings
from server import create_app

SECRET = "test-secret"


def _config(**overrides) -> Settings:
    cfg = Settings.model_validate({})
    cfg.internal_secret = SECRET  # conftest pins INTERNAL_SECRET=test; use ours
    for key, value in overrides.items():
        setattr(cfg, key, value)
    return cfg


def _client(cfg: Settings) -> TestClient:
    return TestClient(create_app(config=cfg, handler=FakeHandler()))


def _auth() -> dict:
    return {"X-Internal-Secret": SECRET}


def _start_body(conversation: str, **body_overrides) -> dict:
    body: dict = {"userId": "u1", "conversationId": conversation, "maxDuration": 300}
    body.update(body_overrides)
    return {"transport": "webrtc", "enableDefaultIceServers": True, "body": body}


def _start(client: TestClient, conversation: str):
    return client.post("/start", json=_start_body(conversation), headers=_auth())


def test_capacity_rejects_the_next_start_with_429():
    with _client(_config(max_sessions=2, max_pending=2)) as client:
        assert _start(client, "c1").status_code == 200
        assert _start(client, "c2").status_code == 200
        resp = _start(client, "c3")
        assert resp.status_code == 429
        assert resp.headers["Retry-After"] == "5"
        assert resp.json()["detail"] == "capacity exhausted"
        registry = client.app.state.registry
        assert registry.rejected_total == 1
        assert registry.live_count() + registry.pending_count() == 2


def test_max_pending_blocks_a_start_flood_before_sessions():
    # Four session slots but one pending slot: a second start that never
    # produced an offer is the flood this guard exists to stop.
    with _client(_config(max_sessions=4, max_pending=1)) as client:
        assert _start(client, "c1").status_code == 200
        assert _start(client, "c2").status_code == 429


def test_repeated_start_for_one_conversation_is_idempotent():
    with _client(_config()) as client:
        first = _start(client, "c1")
        second = _start(client, "c1")
        assert first.status_code == second.status_code == 200
        assert first.json()["sessionId"] == second.json()["sessionId"]
        registry = client.app.state.registry
        assert registry.pending_count() == 1
        assert registry.started_total == 1


def test_validation_rejections():
    cases = [
        ({"transport": "daily"}, 400),
        ({"transport": "webrtc"}, 400),  # missing body
        ({"transport": "webrtc", "body": {"userId": "u", "maxDuration": 300}}, 400),
        ({"transport": "webrtc", "body": {"conversationId": "c1", "maxDuration": 0}}, 400),
        ({"transport": "webrtc", "body": {"conversationId": "c1", "maxDuration": -5}}, 400),
        ({"transport": "webrtc", "body": {"conversationId": "c1", "maxDuration": "x"}}, 400),
    ]
    with _client(_config()) as client:
        for payload, expected in cases:
            resp = client.post("/start", json=payload, headers=_auth())
            assert resp.status_code == expected, (payload, resp.text)


def test_missing_transport_defaults_to_webrtc():
    # Runner parity: Go always names it, but the default matches the runner.
    payload = {"body": {"userId": "u1", "conversationId": "c1", "maxDuration": 300}}
    with _client(_config()) as client:
        assert client.post("/start", json=payload, headers=_auth()).status_code == 200


def test_max_duration_is_clamped_not_rejected():
    cfg = _config(max_duration_sec=300)
    with _client(cfg) as client:
        payload = _start_body("c1", maxDuration=3600)
        resp = client.post("/start", json=payload, headers=_auth())
        assert resp.status_code == 200, resp.text
        stored = client.app.state.registry.get_pending(resp.json()["sessionId"])
        assert stored is not None
        assert stored.body["maxDuration"] == 300


def test_default_stun_is_returned_when_nothing_is_configured():
    with _client(_config()) as client:
        resp = _start(client, "c1")
        assert resp.status_code == 200, resp.text
        assert resp.json()["iceConfig"] == {
            "iceServers": [{"urls": ["stun:stun.l.google.com:19302"]}]
        }


def test_ice_config_is_omitted_when_the_flag_is_off():
    with _client(_config()) as client:
        payload = _start_body("c1")
        payload["enableDefaultIceServers"] = False
        resp = client.post("/start", json=payload, headers=_auth())
        assert resp.status_code == 200
        assert "iceConfig" not in resp.json()


def test_pending_expiry_releases_the_lease():
    with _client(_config(pending_ttl_sec=1)) as client:
        resp = _start(client, "c1")
        session_id = resp.json()["sessionId"]
        registry = client.app.state.registry
        assert registry.get_pending(session_id) is not None
        time.sleep(1.1)
        # A read past the TTL reclaims the entry and its lease.
        assert registry.get_pending(session_id) is None
        assert registry.pending_count() == 0
        # The freed slot admits a new conversation.
        assert _start(client, "c2").status_code == 200
