"""The auth boundary: every data route demands the internal secret; the
health endpoints an orchestrator probes stay open.

Session routes (/start, offer, /drain) join DATA_ROUTES as Tasks 6-8 add
them; each of those tasks extends this list first and watches it fail.
"""

from fastapi.testclient import TestClient
from server_fakes import FakeHandler

from config import Settings
from server import create_app

SECRET = "test-secret"

DATA_ROUTES = [
    ("post", "/start"),
    ("post", "/sessions/any/api/offer"),
    ("patch", "/sessions/any/api/offer"),
    ("post", "/transcript"),
    ("post", "/transcribe"),
    ("post", "/analyze"),
    ("post", "/drain"),
]


def _client() -> TestClient:
    # conftest pins INTERNAL_SECRET=test; the app must be built with the
    # secret THIS file asserts against, not whatever the ambient env holds.
    cfg = Settings.model_validate({})
    cfg.internal_secret = SECRET
    return TestClient(create_app(config=cfg, handler=FakeHandler()))


def test_data_routes_refuse_a_missing_secret():
    with _client() as client:
        for method, path in DATA_ROUTES:
            resp = getattr(client, method)(path)
            assert resp.status_code == 401, (path, resp.status_code)
            assert resp.json()["detail"] == "unauthorized"


def test_data_routes_refuse_a_wrong_secret():
    with _client() as client:
        for method, path in DATA_ROUTES:
            resp = getattr(client, method)(path, headers={"X-Internal-Secret": "nope"})
            assert resp.status_code == 401, (path, resp.status_code)


def test_health_endpoints_stay_open():
    with _client() as client:
        assert client.get("/health").status_code == 200
        assert client.get("/ready").status_code == 200
        assert client.get("/status").status_code == 200


def test_non_ascii_secret_is_401_not_500():
    # hmac.compare_digest raises TypeError on non-ASCII str; an auth
    # failure must stay a 401, never a 500. Raw latin-1 bytes reach the
    # server as a non-ASCII str (httpx refuses non-ASCII str headers
    # client-side, so str headers can never exercise this path).
    with _client() as client:
        resp = client.post(
            "/start", headers=[(b"x-internal-secret", "ünicode".encode("latin-1"))]
        )
        assert resp.status_code == 401, (resp.status_code, resp.text)
        assert resp.json()["detail"] == "unauthorized"


def test_the_right_secret_reaches_the_route():
    # 401 must not mask other refusals: with the secret present, /analyze's
    # own validation answers instead.
    with _client() as client:
        resp = client.post(
            "/analyze",
            json={"conversationId": "c1", "turnId": "t1"},
            headers={"X-Internal-Secret": SECRET},
        )
        assert resp.status_code == 422
