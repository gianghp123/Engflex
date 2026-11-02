import asyncio
from unittest.mock import AsyncMock

import httpx
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.internal import InternalEngineMiddleware
from transcript import sessions
from transcript.capture import TurnCollector
from transcript.sessions import SessionControl


def engine_app():
    app = FastAPI()
    state = {"starts": 0}
    @app.post("/start")
    async def start():
        state["starts"] += 1
        await asyncio.sleep(0)
        return {"sessionId": "official-session"}
    app.add_middleware(InternalEngineMiddleware)
    return app, state


def test_internal_auth_and_hidden_runner_routes():
    app, state = engine_app()
    client = TestClient(app)
    assert client.get("/health").status_code == 200
    assert client.post("/start", json={}).status_code == 401
    assert client.post("/start", json={}, headers=[(b"X-Internal-Secret", b"\xff")]).status_code == 401
    assert client.get("/", headers={"X-Internal-Secret": "test"}).status_code == 404
    assert state["starts"] == 0


def test_concurrent_start_replays_and_rejects_changed_input():
    app, state = engine_app()
    async def run():
        async with httpx.AsyncClient(transport=httpx.ASGITransport(app=app), base_url="http://engine", headers={"X-Internal-Secret": "test"}) as client:
            payload = {"transport": "webrtc", "body": {"conversationId": "c1"}}
            responses = await asyncio.gather(*(client.post("/start", json=payload) for _ in range(8)))
            assert all(r.status_code == 200 and r.json()["sessionId"] == "official-session" for r in responses)
            changed = await client.post("/start", json={**payload, "transport": "daily"})
            assert changed.status_code == 409
            assert state["starts"] == 1
    asyncio.run(run())


def test_drain_ends_active_workers_and_closes_http_client(monkeypatch):
    from app import internal
    app, _ = engine_app()
    middleware = InternalEngineMiddleware(app)
    closed = AsyncMock()
    monkeypatch.setattr(internal, "close_callbacks", closed)
    async def run():
        worker = AsyncMock()
        async def end(*args): sessions.unregister("drain-test")
        worker.queue_frame.side_effect = end
        sessions.register("drain-test", SessionControl("drain-test", worker, TurnCollector()))
        await middleware.drain()
        worker.queue_frame.assert_awaited_once()
        closed.assert_awaited_once()
    asyncio.run(run())
