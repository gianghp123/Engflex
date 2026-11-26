"""ICE resolution: the dev runner's PIPECAT_ICE_SERVERS rules, verbatim.

A malformed configuration must fail at startup — never at the first session.
"""

import pytest
from fastapi.testclient import TestClient
from server_fakes import FakeHandler

from config import Settings
from server import create_app, resolve_ice_servers

SECRET = "test-secret"


def _config(ice: str = "") -> Settings:
    cfg = Settings.model_validate({})
    cfg.internal_secret = SECRET
    cfg.pipecat_ice_servers = ice
    return cfg


def test_empty_config_yields_no_servers():
    assert resolve_ice_servers("") == []


def test_json_array_with_credentials():
    raw = '[{"urls": ["turn:t.example.com:3478"], "username": "u", "credential": "p"}]'
    assert resolve_ice_servers(raw) == [
        {"urls": ["turn:t.example.com:3478"], "username": "u", "credential": "p"}
    ]


def test_single_json_object():
    raw = '{"urls": "stun:stun.example.com:19302"}'
    assert resolve_ice_servers(raw) == [{"urls": "stun:stun.example.com:19302"}]


def test_comma_separated_bare_urls():
    raw = "stun:a.example:19302, stun:b.example:19302"
    assert resolve_ice_servers(raw) == [
        {"urls": "stun:a.example:19302"},
        {"urls": "stun:b.example:19302"},
    ]


def test_unknown_keys_and_missing_urls_are_refused():
    with pytest.raises(ValueError):
        resolve_ice_servers('[{"urls": "stun:x", "password": "no"}]')
    with pytest.raises(ValueError):
        resolve_ice_servers('[{"username": "u"}]')


def test_malformed_json_is_refused():
    with pytest.raises(ValueError):
        resolve_ice_servers("[{not json")


def test_malformed_env_fails_app_startup():
    app = create_app(config=_config("[{not json"), handler=FakeHandler())
    with pytest.raises((ValueError, RuntimeError)), TestClient(app):
        pass


def test_start_ice_config_prefers_configured_servers():
    # (Task 6 adds the /start route; until then this test fails with 404 —
    # it is the failing-first probe for that task. It lives here because the
    # behavior under test is ICE resolution.)
    cfg = _config('[{"urls": ["turn:t.example.com:3478"], "username": "u", "credential": "p"}]')
    with TestClient(create_app(config=cfg, handler=FakeHandler())) as client:
        resp = client.post(
            "/start",
            json={
                "transport": "webrtc",
                "enableDefaultIceServers": True,
                "body": {"userId": "u1", "conversationId": "c1", "maxDuration": 300},
            },
            headers={"X-Internal-Secret": SECRET},
        )
        assert resp.status_code == 200, resp.text
        assert resp.json()["iceConfig"] == {
            "iceServers": [
                {"urls": ["turn:t.example.com:3478"], "username": "u", "credential": "p"}
            ]
        }
