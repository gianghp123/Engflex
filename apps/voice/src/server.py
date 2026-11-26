"""Production dispatcher: the FastAPI entrypoint that replaces the dev runner.

Serves exactly the session API Go already speaks — POST /start, the offer
proxy, the data routes — behind the internal secret, plus the
health/ready/status/drain surface an operator needs. One process hosts up to
MAX_SESSIONS pipelines; the deliberate deviation from Pipecat's
per-session-process guidance (SmallWebRTC binds the peer connection to the
process that handled the offer) is argued in the design spec, §4.1.

Entrypoint: `uv run src/server.py` (VOICE_HOST / VOICE_PORT).
"""

import asyncio
import hmac
import json
import os
import signal
import time
from contextlib import asynccontextmanager
from typing import Any

import uvicorn
from fastapi import Depends, FastAPI, Header, HTTPException, Request
from fastapi.responses import JSONResponse
from fastapi.routing import APIRouter
from loguru import logger
from pipecat.runner.types import SmallWebRTCRunnerArguments
from pipecat.transports.smallwebrtc.connection import IceServer
from pipecat.transports.smallwebrtc.request_handler import (
    IceCandidate,
    SmallWebRTCPatchRequest,
    SmallWebRTCRequest,
    SmallWebRTCRequestHandler,
)

from app.routes import router as data_routes
from bot import run_session
from config import Settings, settings
from session import CapacityFull, Draining, SessionActive, SessionRegistry

SWEEP_INTERVAL_SEC = 5
DEFAULT_STUN = [{"urls": ["stun:stun.l.google.com:19302"]}]


def resolve_ice_servers(raw: str) -> list[dict[str, Any]]:
    """Parse PIPECAT_ICE_SERVERS with the dev runner's rules, verbatim.

    A JSON array of {urls, username?, credential?}, a single JSON object, or
    a comma-separated list of bare URLs. Plain dictionaries, not IceServer
    objects: the same values cross the wire to the browser. Malformed input
    raises ValueError — the app refuses to start rather than failing at the
    first session.
    """
    text = (raw or "").strip()
    if not text:
        return []
    if text.startswith("["):
        try:
            entries: list[Any] = json.loads(text)
        except json.JSONDecodeError as exc:
            raise ValueError(f"could not parse {text!r} as a JSON array: {exc}") from exc
        if not isinstance(entries, list):
            raise ValueError(f"expected a JSON array of ICE servers, got {text!r}")
    elif text.startswith("{"):
        entries = [text]
    else:
        entries = [item.strip() for item in text.split(",") if item.strip()]

    allowed_keys = {"urls", "username", "credential"}
    servers: list[dict[str, Any]] = []
    for entry in entries:
        if isinstance(entry, str):
            stripped = entry.strip()
            if not stripped:
                continue
            if not stripped.startswith("{"):
                servers.append({"urls": stripped})
                continue
            try:
                entry = json.loads(stripped)
            except json.JSONDecodeError as exc:
                raise ValueError(f"could not parse {stripped!r} as a JSON object: {exc}") from exc
        if not isinstance(entry, dict):
            raise ValueError(f"expected a URL or a JSON object, got {entry!r}")  # noqa: TRY004
        unknown = set(entry) - allowed_keys
        if unknown:
            raise ValueError(
                f"unsupported key(s) {sorted(unknown)} in {entry!r}; "
                f"supported keys are {sorted(allowed_keys)}"
            )
        if not entry.get("urls"):
            raise ValueError(f"missing 'urls' in {entry!r}")
        servers.append({key: entry[key] for key in entry if entry[key] is not None})
    return servers


def _request_process_exit() -> None:
    """Ask uvicorn to shut down once a drain has finished. A test seam: the
    drain tests monkeypatch this so the suite's own process is never
    signalled."""
    os.kill(os.getpid(), signal.SIGINT)


async def _sweep_loop(registry: SessionRegistry) -> None:
    while True:
        await asyncio.sleep(SWEEP_INTERVAL_SEC)
        try:
            await registry.sweep_once()
        except Exception:  # noqa: BLE001 - one bad pass must not kill the sweeper
            logger.exception("sweep pass failed")


async def _drain_then_exit(app: FastAPI) -> None:
    await app.state.registry.wait_drain()
    _request_process_exit()


def _parse_offer_body(data: dict[str, Any]) -> SmallWebRTCRequest:
    """Build the handler request from the offer JSON; client request_data is
    deliberately dropped — the stored /start body is the only body."""
    if not isinstance(data.get("sdp"), str) or not data["sdp"]:
        raise HTTPException(status_code=400, detail="Invalid WebRTC request")
    if not isinstance(data.get("type"), str) or not data["type"]:
        raise HTTPException(status_code=400, detail="Invalid WebRTC request")
    try:
        return SmallWebRTCRequest(
            sdp=data["sdp"],
            type=data["type"],
            pc_id=data.get("pc_id"),
            restart_pc=data.get("restart_pc"),
        )
    except (KeyError, TypeError) as exc:
        raise HTTPException(status_code=400, detail="Invalid WebRTC request") from exc


async def _close_connection(connection: Any) -> None:
    """Best-effort close; cleanup must never mask the failure it answers."""
    try:
        await connection.disconnect()
    except Exception:  # noqa: BLE001
        logger.warning("connection close failed")


def _parse_ice_candidate(data: Any) -> IceCandidate:
    """Build one candidate from the browser's JSON.

    `RTCIceCandidate.toJSON()` speaks camelCase (`sdpMid`, `sdpMLineIndex`)
    while the installed `IceCandidate` dataclass takes snake_case kwargs, so
    both spellings are accepted here. Anything else is a 400.
    """
    if not isinstance(data, dict):
        raise TypeError(f"expected a candidate object, got {data!r}")
    try:
        candidate = data["candidate"]
        sdp_mid = data["sdp_mid"] if "sdp_mid" in data else data["sdpMid"]
        sdp_mline_index = (
            data["sdp_mline_index"] if "sdp_mline_index" in data else data["sdpMLineIndex"]
        )
    except KeyError as exc:
        raise ValueError(f"missing candidate field: {exc}") from exc
    return IceCandidate(candidate=candidate, sdp_mid=sdp_mid, sdp_mline_index=sdp_mline_index)


def create_app(config: Settings | None = None, handler: Any | None = None) -> FastAPI:
    """Build the dispatcher app.

    `config` defaults to the process settings; tests pass their own instance.
    `handler` defaults to a real SmallWebRTCRequestHandler; tests pass a fake
    that pins its call contract without negotiating a peer connection.
    """
    cfg = config or settings

    @asynccontextmanager
    async def lifespan(app: FastAPI):
        # Fail fast: a malformed ICE list is a boot error, not a runtime one.
        configured_ice = resolve_ice_servers(cfg.pipecat_ice_servers)
        webrtc_handler = handler or SmallWebRTCRequestHandler(
            ice_servers=[IceServer(**s) for s in configured_ice] or None
        )
        app.state.handler = webrtc_handler
        app.state.configured_ice = configured_ice
        app.state.registry = SessionRegistry(
            max_sessions=cfg.max_sessions,
            # Unreachable after Settings validation (MAX_PENDING >= 1); kept
            # for typing, since max_pending is Optional.
            max_pending=cfg.max_pending or 1,
            pending_ttl_sec=cfg.pending_ttl_sec,
            drain_grace_sec=cfg.drain_grace_sec,
            max_duration_sec=cfg.max_duration_sec,
        )
        app.state.started_at = time.time()
        app.state.sweeper = asyncio.create_task(_sweep_loop(app.state.registry))
        logger.info(
            f"dispatcher up max_sessions={cfg.max_sessions} "
            f"max_pending={app.state.registry.max_pending} "
            f"ice_servers={len(configured_ice)}"
        )
        try:
            yield
        finally:
            app.state.sweeper.cancel()
            try:
                await app.state.sweeper
            except asyncio.CancelledError:
                pass
            # SIGTERM lands here after uvicorn stopped accepting: block the
            # exit until live sessions drained or the grace expired.
            await app.state.registry.drain()
            await webrtc_handler.close()
            logger.info("dispatcher down")

    app = FastAPI(title="engflex-voice", lifespan=lifespan)

    async def require_secret(
        x_internal_secret: str | None = Header(default=None, alias="X-Internal-Secret"),
    ) -> None:
        if x_internal_secret is None:
            raise HTTPException(status_code=401, detail="unauthorized")
        try:
            ok = hmac.compare_digest(x_internal_secret, cfg.internal_secret)
        except (TypeError, ValueError):
            ok = False
        if not ok:
            raise HTTPException(status_code=401, detail="unauthorized")

    # --- session routes (every one behind the internal secret) --------------
    internal = APIRouter(dependencies=[Depends(require_secret)])

    @internal.post("/start")
    async def start(request: Request):
        """Admit one session: validate, lease, park it as PENDING.

        Mirrors the dev runner's /start contract (sessionId + optional
        iceConfig) with admission bolted on: 429 at capacity, 409 when the
        conversation is already live, 503 while draining.
        """
        registry: SessionRegistry = request.app.state.registry
        try:
            data = await request.json()
        except Exception as exc:
            raise HTTPException(status_code=400, detail="invalid request body") from exc
        if not isinstance(data, dict):
            raise HTTPException(status_code=400, detail="invalid request body")
        transport = data.get("transport") or "webrtc"
        if transport != "webrtc":
            raise HTTPException(
                status_code=400, detail="only the webrtc transport is supported"
            )
        body = data.get("body")
        if not isinstance(body, dict):
            raise HTTPException(status_code=400, detail="missing body")
        conversation_id = body.get("conversationId")
        if not isinstance(conversation_id, str) or not conversation_id:
            raise HTTPException(status_code=400, detail="missing conversationId")
        max_duration = body.get("maxDuration", 300)
        if (
            isinstance(max_duration, bool)
            or not isinstance(max_duration, int)
            or max_duration <= 0
        ):
            raise HTTPException(
                status_code=400, detail="maxDuration must be a positive integer"
            )
        # Clamped, never rejected: the cap is a worker limit, not caller
        # validation.
        body = {**body, "maxDuration": min(max_duration, cfg.max_duration_sec)}
        # Runner parity: configured servers win; the Google STUN fallback is
        # returned to the browser only, never added to the handler's list.
        if request.app.state.configured_ice:
            ice_config: dict[str, Any] | None = {
                "iceServers": request.app.state.configured_ice
            }
        elif data.get("enableDefaultIceServers"):
            ice_config = {"iceServers": DEFAULT_STUN}
        else:
            ice_config = None
        try:
            pending = await registry.start(conversation_id, body, ice_config)
        except SessionActive as exc:
            raise HTTPException(status_code=409, detail="conversation already live") from exc
        except CapacityFull as exc:
            logger.info(
                "session_rejected",
                reason=str(exc),
                active=registry.live_count(),
                pending=registry.pending_count(),
            )
            raise HTTPException(
                status_code=429,
                detail="capacity exhausted",
                headers={"Retry-After": "5"},
            ) from exc
        except Draining as exc:
            raise HTTPException(
                status_code=503, detail="draining", headers={"Retry-After": "5"}
            ) from exc
        result: dict[str, Any] = {"sessionId": pending.session_id}
        if pending.ice_config is not None:
            result["iceConfig"] = pending.ice_config
        return result

    @internal.post("/sessions/{session_id}/api/offer")
    async def offer(session_id: str, request: Request):
        """Submit the browser's SDP offer; the pipeline starts here.

        The stored /start body is authoritative: any request_data the offer
        carries is discarded (trust boundary — the browser must not be able
        to change persona/scenario/learner).
        """
        registry = request.app.state.registry
        live = registry.get_live(session_id)
        pending = registry.get_pending(session_id)
        if live is None and pending is None:
            raise HTTPException(status_code=404, detail="unknown or expired session")
        try:
            data = await request.json()
        except Exception as exc:
            raise HTTPException(status_code=400, detail="Invalid WebRTC request") from exc
        if not isinstance(data, dict):
            raise HTTPException(status_code=400, detail="Invalid WebRTC request")
        if live is not None and not data.get("restart_pc"):
            raise HTTPException(status_code=409, detail="session already live")
        webrtc_request = _parse_offer_body(data)

        async def connection_callback(connection: Any) -> None:
            entry = registry.get_pending(session_id)
            if entry is None:
                # The pending lease expired during negotiation: nobody holds
                # capacity for this connection, so refuse to go live.
                logger.warning("session expired during offer", session_id=session_id)
                await _close_connection(connection)
                return
            runner_args = SmallWebRTCRunnerArguments(
                webrtc_connection=connection, body=entry.body, session_id=session_id
            )
            task = asyncio.create_task(run_session(runner_args, registry=registry))
            if await registry.mark_live(session_id, connection.pc_id, connection, task) is None:
                # mark_live lost the TTL race: undo the just-started pipeline.
                task.cancel()
                await _close_connection(connection)

        try:
            answer = await request.app.state.handler.handle_web_request(
                request=webrtc_request, webrtc_connection_callback=connection_callback
            )
        except Exception:
            # Offer failure releases the lease (spec §9): the client can
            # /start again instead of waiting out the TTL. The 500 this
            # becomes is spec §6.2's "handler/pipeline start failure".
            registry.drop_pending(session_id)
            raise
        logger.info("session_offer", session_id=session_id, ok=True)
        return answer

    @internal.patch("/sessions/{session_id}/api/offer")
    async def patch_ice(session_id: str, request: Request):
        """Route ICE candidates to the session's peer connection."""
        registry = request.app.state.registry
        live = registry.get_live(session_id)
        pending = registry.get_pending(session_id)
        if live is None and pending is None:
            raise HTTPException(status_code=404, detail="unknown or expired session")
        if live is None or live.pc_id is None:
            raise HTTPException(status_code=409, detail="no peer connection yet")
        try:
            data = await request.json()
            patch_request = SmallWebRTCPatchRequest(
                pc_id=data["pc_id"],
                candidates=[_parse_ice_candidate(c) for c in data.get("candidates", [])],
            )
        except (AttributeError, KeyError, TypeError, ValueError) as exc:
            raise HTTPException(status_code=400, detail="Invalid WebRTC request") from exc
        # The handler 404s an unknown pc_id; the session checks above keep
        # that from ever masking a missing session.
        await request.app.state.handler.handle_patch_request(patch_request)
        return {"status": "success"}

    @internal.post("/drain")
    async def drain(request: Request):
        """Begin a graceful drain: no new sessions, live ones get the grace.

        The gate flips synchronously with this response, so an immediate
        /start is already refused; the background half (waiting for live
        sessions, then exiting) runs after the 202.
        """
        registry: SessionRegistry = request.app.state.registry
        if not registry.draining:
            registry.begin_drain()
            asyncio.create_task(_drain_then_exit(request.app))
        return JSONResponse(
            status_code=202,
            content={
                "active": registry.live_count(),
                "pending": registry.pending_count(),
                "graceSec": registry.drain_grace_sec,
            },
        )

    app.include_router(internal)

    # --- existing data routes, now behind the secret -------------------------
    app.include_router(data_routes, dependencies=[Depends(require_secret)])

    # --- operator surface (no auth: healthcheck probes) ---------------------

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.get("/ready")
    async def ready(request: Request):
        if request.app.state.registry.draining:
            raise HTTPException(status_code=503, detail="draining", headers={"Retry-After": "5"})
        return {"status": "ready"}

    @app.get("/status")
    async def status(request: Request):
        registry = request.app.state.registry
        body: dict[str, Any] = {
            "status": "draining" if registry.draining else "ready",
            "version": cfg.voice_version,
            "transports": ["webrtc"],
            "uptimeSec": int(time.time() - request.app.state.started_at),
        }
        body.update(registry.status())
        return body

    return app


app = create_app()


if __name__ == "__main__":
    uvicorn.run(app, host=settings.voice_host, port=settings.voice_port)
