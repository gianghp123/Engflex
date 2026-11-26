"""Dispatcher session state: leases, TTLs, transitions, drain.

One process hosts every pipeline (the design spec's §4.1 argues the
deliberate deviation from Pipecat's per-session-process guidance), so this
module is the single source of truth for how many sessions exist and what
state each is in. Every PENDING and LIVE entry holds exactly one lease
against MAX_SESSIONS; a lease is released exactly once — pending expiry,
live end, or drain cancel — and the sweeper reclaims anything a missed path
leaves behind.

The transcript registry (`transcript.sessions._SESSIONS`) stays separate and
keyed by conversation id: `bot()` owns it, the dispatcher never writes it.
"""

import asyncio
import time
import uuid
from dataclasses import dataclass, field
from typing import Any

from loguru import logger


class SessionActive(Exception):
    """A session for the conversation is already LIVE (409)."""


class CapacityFull(Exception):
    """MAX_SESSIONS or MAX_PENDING is exhausted (429)."""


class Draining(Exception):
    """The worker refuses new sessions while draining (503)."""


@dataclass
class PendingSession:
    """A /start lease waiting for its SDP offer."""

    session_id: str
    conversation_id: str
    body: dict[str, Any]
    ice_config: dict[str, Any] | None
    created_at: float
    expires_at: float


@dataclass
class LiveSession:
    """A running pipeline: the dispatcher tracks it, bot() owns its state."""

    session_id: str
    conversation_id: str
    pc_id: str | None = None
    connection: Any | None = None
    task: asyncio.Task | None = None
    started_at: float = field(default_factory=time.time)


class SessionRegistry:
    """PENDING/LIVE maps and the lease invariant for one worker."""

    def __init__(
        self,
        *,
        max_sessions: int,
        max_pending: int,
        pending_ttl_sec: int,
        drain_grace_sec: int,
        max_duration_sec: int,
    ) -> None:
        self.max_sessions = max_sessions
        self.max_pending = max_pending
        self.pending_ttl_sec = pending_ttl_sec
        self.drain_grace_sec = drain_grace_sec
        self.max_duration_sec = max_duration_sec
        self.draining = False
        self.started_total = 0
        self.completed_total = 0
        self.rejected_total = 0
        self._lock = asyncio.Lock()
        self._pending: dict[str, PendingSession] = {}
        self._live: dict[str, LiveSession] = {}
        self._by_conversation: dict[str, str] = {}

    # --- reads --------------------------------------------------------------

    def pending_count(self) -> int:
        return len(self._pending)

    def live_count(self) -> int:
        return len(self._live)

    def get_pending(self, session_id: str) -> PendingSession | None:
        """The pending entry, or None when missing or past its TTL.

        A read past the TTL reclaims the entry and its lease on the spot —
        a read is as good as a sweep.
        """
        entry = self._pending.get(session_id)
        if entry is None:
            return None
        if time.time() > entry.expires_at:
            self._expire_pending(session_id)
            return None
        return entry

    def get_live(self, session_id: str) -> LiveSession | None:
        return self._live.get(session_id)

    # --- transitions ---------------------------------------------------------

    async def start(
        self, conversation_id: str, body: dict[str, Any], ice_config: dict[str, Any] | None
    ) -> PendingSession:
        """Admit one session under the lease invariant.

        Idempotent per conversation: a repeat /start returns the existing
        pending session (still one lease); a LIVE conversation refuses with
        SessionActive; capacity and the pending flood guard refuse with
        CapacityFull; a draining worker refuses with Draining.
        """
        async with self._lock:
            if self.draining:
                raise Draining()
            existing_id = self._by_conversation.get(conversation_id)
            if existing_id is not None and existing_id in self._live:
                raise SessionActive(conversation_id)
            if existing_id is not None and existing_id in self._pending:
                entry = self._pending[existing_id]
                if time.time() <= entry.expires_at:
                    return entry
                self._expire_pending(existing_id)  # stale: a fresh lease below
            if len(self._pending) + len(self._live) >= self.max_sessions:
                self.rejected_total += 1
                raise CapacityFull("MAX_SESSIONS")
            if len(self._pending) >= self.max_pending:
                self.rejected_total += 1
                raise CapacityFull("MAX_PENDING")
            entry = PendingSession(
                session_id=str(uuid.uuid4()),
                conversation_id=conversation_id,
                body=body,
                ice_config=ice_config,
                created_at=time.time(),
                expires_at=time.time() + self.pending_ttl_sec,
            )
            self._pending[entry.session_id] = entry
            self._by_conversation[conversation_id] = entry.session_id
            self.started_total += 1
        logger.info(
            "session_start",
            session_id=entry.session_id,
            conversation_id=conversation_id,
            active=len(self._live),
            pending=len(self._pending),
        )
        return entry

    async def mark_live(
        self, session_id: str, pc_id: str, connection: Any, task: asyncio.Task
    ) -> LiveSession | None:
        """PENDING -> LIVE. None when the pending entry is gone (the TTL won
        the race) — the caller must then close the connection instead of
        running the pipeline."""
        async with self._lock:
            pending = self._pending.pop(session_id, None)
            if pending is None:
                return None
            if time.time() > pending.expires_at:
                # The TTL won the race with negotiation: drop the
                # conversation mapping and refuse to go live.
                self._by_conversation.pop(pending.conversation_id, None)
                return None
            live = LiveSession(
                session_id=session_id,
                conversation_id=pending.conversation_id,
                pc_id=pc_id,
                connection=connection,
                task=task,
                started_at=time.time(),
            )
            self._live[session_id] = live
        logger.info(
            "session_live",
            session_id=session_id,
            duration_ms=int((time.time() - pending.created_at) * 1000),
        )
        return live

    def end_live(self, session_id: str, reason: str) -> LiveSession | None:
        """Release a LIVE entry's lease — the one guaranteed path is
        run_session's finally. Idempotent: a second caller gets None."""
        live = self._live.pop(session_id, None)
        if live is None:
            return None
        self._by_conversation.pop(live.conversation_id, None)
        self.completed_total += 1
        logger.info(
            "session_end",
            session_id=session_id,
            reason=reason,
            duration_sec=int(time.time() - live.started_at),
        )
        return live

    def _expire_pending(self, session_id: str) -> None:
        entry = self._pending.pop(session_id, None)
        if entry is not None:
            self._by_conversation.pop(entry.conversation_id, None)
            logger.info("session_expired", session_id=session_id)

    def drop_pending(self, session_id: str) -> None:
        """Release a pending lease on offer failure (spec §9: the client can
        /start again instead of waiting out the TTL). No-op for unknown or
        live sessions."""
        entry = self._pending.pop(session_id, None)
        if entry is not None:
            self._by_conversation.pop(entry.conversation_id, None)
            logger.warning("session dropped after offer failure", session_id=session_id)

    # --- sweeper -----------------------------------------------------------

    async def sweep_once(self) -> None:
        """One sweeper pass: expire overdue pending, reap live entries whose
        task already finished, cancel sessions past the hard duration cap."""
        now = time.time()
        for session_id, entry in list(self._pending.items()):
            if now > entry.expires_at:
                self._expire_pending(session_id)
        for session_id, live in list(self._live.items()):
            if live.task is not None and live.task.done():
                self.end_live(session_id, reason="swept-done")
                continue
            if now - live.started_at > self.max_duration_sec + 30:
                if live.task is not None:
                    live.task.cancel()
                    try:
                        await live.task
                    except (asyncio.CancelledError, Exception):  # noqa: BLE001, S110
                        pass
                self.end_live(session_id, reason="overrun")
                logger.warning("session overrun cancelled", session_id=session_id)

    # --- drain ---------------------------------------------------------------

    def begin_drain(self) -> None:
        """Flip the admission gate and cancel every PENDING session. Live
        sessions keep running; wait_drain bounds the rest."""
        self.draining = True
        for session_id in list(self._pending):
            self._expire_pending(session_id)
        logger.info("drain_start", active=len(self._live), pending=0)

    async def wait_drain(self) -> dict[str, int]:
        """Wait for live sessions to end naturally, then force the rest."""
        deadline = time.time() + self.drain_grace_sec
        while self._live and time.time() < deadline:
            await asyncio.sleep(0.5)
        for live in list(self._live.values()):
            if live.task is not None:
                live.task.cancel()
        for session_id, live in list(self._live.items()):
            if live.task is not None:
                try:
                    await live.task
                except (asyncio.CancelledError, Exception):  # noqa: BLE001, S110
                    pass
            # run_session's finally normally did this; the second pop is a
            # no-op when it did, and the net when it could not.
            self.end_live(session_id, reason="drained")
        logger.info("drain_done", active=len(self._live))
        return {"active": len(self._live)}

    async def drain(self) -> dict[str, int]:
        """The full SIGTERM path: begin, then wait."""
        self.begin_drain()
        return await self.wait_drain()

    # --- status ---------------------------------------------------------------

    def status(self) -> dict[str, Any]:
        return {
            "active": len(self._live),
            "pending": len(self._pending),
            "max": self.max_sessions,
            "pendingMax": self.max_pending,
            "draining": self.draining,
            "startedTotal": self.started_total,
            "completedTotal": self.completed_total,
            "rejectedTotal": self.rejected_total,
        }
