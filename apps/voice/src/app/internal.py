"""Protect runner internals and wrap its official /start route idempotently."""
import asyncio
import hashlib
import hmac
import json
import time

from loguru import logger
from starlette.requests import Request
from starlette.responses import JSONResponse, Response

from clients.go_callbacks import close_callbacks
from config import settings


class InternalEngineMiddleware:
    def __init__(self, app):
        self.app = app
        self.draining = False
        self.starts: dict[str, tuple[str, bytes, int, float]] = {}
        self.locks: dict[str, asyncio.Lock] = {}

    async def __call__(self, scope, receive, send):
        if scope["type"] == "lifespan":
            async def lifecycle_receive():
                message = await receive()
                if message["type"] == "lifespan.startup":
                    from observability import configure_logging
                    configure_logging()
                if message["type"] == "lifespan.shutdown":
                    self.draining = True
                    await self.drain()
                return message
            return await self.app(scope, lifecycle_receive, send)
        if scope["type"] != "http":
            if scope["type"] == "websocket":
                await send({"type": "websocket.close", "code": 1008})
                return
            return await self.app(scope, receive, send)
        request = Request(scope, receive)
        path = scope["path"]
        if path in {"/health", "/readyz"}:
            status = 503 if self.draining and path == "/readyz" else 200
            return await JSONResponse({"status": "draining" if status == 503 else "ok"}, status)(scope, receive, send)
        supplied = request.headers.get("X-Internal-Secret", "")
        if not supplied or not hmac.compare_digest(supplied.encode(), settings.internal_secret.encode()):
            return await JSONResponse({"detail": "unauthorized"}, 401)(scope, receive, send)
        allowed = path in {"/start", "/analyze", "/transcribe", "/transcript"} or (
            path.startswith("/sessions/") and path.endswith("/api/offer"))
        if not allowed:
            return await JSONResponse({"detail": "not found"}, 404)(scope, receive, send)
        limit = 21 * 1024 * 1024 if path == "/transcribe" else 6 * 1024 * 1024
        body = bytearray()
        while True:
            part = await receive()
            if part["type"] == "http.disconnect":
                return
            body.extend(part.get("body", b""))
            if len(body) > limit:
                return await JSONResponse({"detail": "body too large"}, 413)(scope, receive, send)
            if not part.get("more_body", False):
                break
        async def replay():
            return {"type": "http.request", "body": bytes(body), "more_body": False}
        if path != "/start":
            return await self.app(scope, replay, send)
        if self.draining:
            return await JSONResponse({"detail": "draining"}, 503)(scope, replay, send)
        try:
            payload = json.loads(body)
            conversation_id = payload["body"]["conversationId"]
            if not isinstance(conversation_id, str) or not conversation_id:
                raise ValueError("conversation id required")
            fingerprint = hashlib.sha256(json.dumps(payload, sort_keys=True).encode()).hexdigest()
        except (ValueError, KeyError, TypeError):
            return await JSONResponse({"detail": "invalid start input"}, 422)(scope, replay, send)
        # A bounded process-local cache is sufficient for the single replica.
        # Retain completed starts until process restart so retries never mint
        # a replacement for an expired runner session.
        if len(self.locks) >= 10000 and conversation_id not in self.locks:
            return await JSONResponse({"detail": "start capacity reached"}, 503)(scope, replay, send)
        lock = self.locks.setdefault(conversation_id, asyncio.Lock())
        async with lock:
            cached = self.starts.get(conversation_id)
            if cached:
                previous, content, status, _ = cached
                response = Response(content, status, media_type="application/json") if previous == fingerprint else JSONResponse({"detail": "start input differs"}, 409)
                return await response(scope, replay, send)
            messages = []
            async def capture(message):
                messages.append(message)
            await self.app(scope, replay, capture)
            status = next(m["status"] for m in messages if m["type"] == "http.response.start")
            content = b"".join(m.get("body", b"") for m in messages if m["type"] == "http.response.body")
            if 200 <= status < 300:
                self.starts[conversation_id] = (fingerprint, content, status, time.monotonic())
            for message in messages:
                await send(message)

    async def drain(self):
        from pipecat.frames.frames import EndTaskFrame
        from pipecat.processors.frame_processor import FrameDirection

        from transcript.sessions import active
        controls = active()
        for control in controls:
            await control.worker.queue_frame(EndTaskFrame(), FrameDirection.UPSTREAM)
        try:
            async with asyncio.timeout(90):
                while active():
                    await asyncio.sleep(0.05)
        except TimeoutError:
            logger.error("voice drain expired; callbacks may be unacknowledged",
                         conversations=[c.conversation_id for c in active()])
        await close_callbacks()
