"""The dispatcher's lease invariant, tested directly against the registry.

Each scenario runs inside a single asyncio.run: the registry's lock binds to
the loop it first awaits on, so every await for one registry instance must
happen inside one loop (and each test gets a fresh registry/lock anyway).
"""

import asyncio
import time

import pytest

from session import CapacityFull, Draining, SessionActive, SessionRegistry


def make_registry(**overrides) -> SessionRegistry:
    params: dict = {
        "max_sessions": 2,
        "max_pending": 2,
        "pending_ttl_sec": 120,
        "drain_grace_sec": 1,
        "max_duration_sec": 300,
    }
    params.update(overrides)
    return SessionRegistry(**params)


async def _park() -> None:
    await asyncio.Event().wait()


async def _live_session(registry: SessionRegistry, conversation: str):
    pending = await registry.start(conversation, {}, None)
    task = asyncio.create_task(_park())
    live = await registry.mark_live(pending.session_id, "pc-1", object(), task)
    assert live is not None
    return pending.session_id, task


async def _cancel(task: asyncio.Task) -> None:
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass


def test_start_admits_up_to_capacity_then_refuses():
    async def scenario():
        registry = make_registry()
        first = await registry.start("c1", {}, None)
        second = await registry.start("c2", {}, None)
        assert first.session_id != second.session_id
        with pytest.raises(CapacityFull):
            await registry.start("c3", {}, None)
        assert registry.rejected_total == 1

    asyncio.run(scenario())


def test_pending_flood_guard_is_independent_of_capacity():
    async def scenario():
        registry = make_registry(max_sessions=4, max_pending=1)
        await registry.start("c1", {}, None)
        with pytest.raises(CapacityFull):
            await registry.start("c2", {}, None)

    asyncio.run(scenario())


def test_repeated_start_for_one_conversation_is_idempotent():
    async def scenario():
        registry = make_registry()
        first = await registry.start("c1", {}, None)
        again = await registry.start("c1", {}, None)
        assert again.session_id == first.session_id
        assert registry.started_total == 1
        assert registry.pending_count() == 1

    asyncio.run(scenario())


def test_a_live_conversation_cannot_start_again():
    async def scenario():
        registry = make_registry()
        await _live_session(registry, "c1")
        with pytest.raises(SessionActive):
            await registry.start("c1", {}, None)

    asyncio.run(scenario())


def test_end_live_releases_the_lease():
    async def scenario():
        registry = make_registry()
        session_id, task = await _live_session(registry, "c1")
        assert registry.end_live(session_id, reason="finished") is not None
        assert registry.end_live(session_id, reason="again") is None  # idempotent
        assert registry.live_count() == 0
        # The freed slot admits a new conversation.
        await registry.start("c2", {}, None)
        await _cancel(task)

    asyncio.run(scenario())


def test_pending_expiry_releases_the_lease():
    async def scenario():
        registry = make_registry(pending_ttl_sec=1)
        pending = await registry.start("c1", {}, None)
        await asyncio.sleep(1.1)
        # A read past the TTL reclaims the entry and its lease.
        assert registry.get_pending(pending.session_id) is None
        assert registry.pending_count() == 0
        await registry.start("c2", {}, None)
        await registry.start("c3", {}, None)  # c1's freed slot
        with pytest.raises(CapacityFull):
            await registry.start("c4", {}, None)

    asyncio.run(scenario())


def test_drop_pending_releases_the_lease():
    async def scenario():
        registry = make_registry()
        pending = await registry.start("c1", {}, None)
        registry.drop_pending(pending.session_id)
        assert registry.get_pending(pending.session_id) is None
        assert registry.pending_count() == 0
        # No-op for unknown ids (a live session, say).
        registry.drop_pending("nope")
        # The conversation can start again and gets a fresh session.
        fresh = await registry.start("c1", {}, None)
        assert fresh.session_id != pending.session_id

    asyncio.run(scenario())


def test_sweep_expires_overdue_pending_and_cancels_overruns():
    async def scenario():
        registry = make_registry(max_sessions=4, max_pending=4)
        fresh = await registry.start("fresh", {}, None)
        stale = await registry.start("stale", {}, None)
        registry._pending[stale.session_id].expires_at = time.time() - 1
        live_id, _task = await _live_session(registry, "c-live")
        registry._live[live_id].started_at = time.time() - 400  # past cap + 30

        await registry.sweep_once()

        assert registry.get_pending(fresh.session_id) is not None
        assert registry.get_pending(stale.session_id) is None
        assert registry.live_count() == 0  # overrun cancelled and reaped

    asyncio.run(scenario())


def test_sweep_reaps_finished_tasks_the_normal_path_missed():
    async def scenario():
        registry = make_registry()
        _session_id, task = await _live_session(registry, "c1")
        await _cancel(task)
        # The finally that would have removed it never ran (this is the
        # missed-path net): the sweep reaps the orphaned entry.
        assert registry.live_count() == 1
        await registry.sweep_once()
        assert registry.live_count() == 0

    asyncio.run(scenario())


def test_drain_cancels_pending_waits_for_live_then_refuses_starts():
    async def scenario():
        registry = make_registry(max_sessions=4, max_pending=4)
        pending = await registry.start("c-pending", {}, None)
        _, _task = await _live_session(registry, "c-live")

        result = await registry.drain()

        assert registry.draining is True
        assert registry.get_pending(pending.session_id) is None
        assert registry.live_count() == 0
        assert result == {"active": 0}
        with pytest.raises(Draining):
            await registry.start("c-new", {}, None)

    asyncio.run(scenario())


def test_status_reports_the_operational_counters():
    async def scenario():
        registry = make_registry(max_sessions=4, max_pending=3)
        await registry.start("c1", {}, None)
        _, task = await _live_session(registry, "c2")
        body = registry.status()
        assert body == {
            "active": 1,
            "pending": 1,
            "max": 4,
            "pendingMax": 3,
            "draining": False,
            "startedTotal": 2,
            "completedTotal": 0,
            "rejectedTotal": 0,
        }
        await _cancel(task)

    asyncio.run(scenario())
