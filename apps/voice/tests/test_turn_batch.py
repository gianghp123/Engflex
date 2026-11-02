import asyncio
from unittest.mock import AsyncMock

import pytest

from clients import go_callbacks, go_turns
from transcript.capture import SnapshotLimitExceeded, TurnCollector


def test_snapshot_copies_corrections_and_truncations():
    c = TurnCollector()
    c.add_user_text("wrong")
    c.add_bot_text("stale answer")
    before, records = c.snapshot()
    c.begin_correction("correct")
    after, corrected = c.snapshot()
    assert after > before
    assert [r.text for r in records] == ["wrong", "stale answer"]
    assert [r.text for r in corrected] == ["correct"]


def test_publisher_coalesces_and_finalizes_full_snapshot(monkeypatch):
    sent = []
    async def post(conversation, operation, payload):
        sent.append((operation, payload))
        await asyncio.sleep(0)
        return True
    async def final(conversation, duration, payload):
        sent.append(("finalize", {**payload, "durationSec": duration}))
        return True
    monkeypatch.setattr(go_turns, "post_callback", post)
    monkeypatch.setattr(go_turns, "finalize_session", final)
    async def run():
        c = TurnCollector()
        p = go_turns.SnapshotPublisher("c1", "s1", c)
        c.on_change = p.notify
        for text in ["one", "two", "three"]:
            c.add_user_text(text)
        await p.task
        c.add_bot_text("answer")
        assert await p.finalize(0)
        assert len(sent) == 2
        assert sent[-1][1]["engineSessionId"] == "s1"
        assert sent[-1][1]["durationSec"] == 0
        assert len(sent[-1][1]["turns"]) == 2
    asyncio.run(run())


def test_over_limit_refuses_without_truncating():
    c = TurnCollector()
    c.add_user_text("original")
    _, before = c.snapshot()
    with pytest.raises(SnapshotLimitExceeded):
        c.add_user_text("x" * 10000)
    assert c.snapshot()[1] == before
    assert not c.accepting


@pytest.mark.parametrize("statuses,expected,calls", [([500, 429, 200], True, 3), ([401], False, 1), ([422], False, 1), ([503]*5, False, 5)])
def test_callback_retries_only_transient_errors(monkeypatch, statuses, expected, calls):
    class Response:
        def __init__(self, status): self.status = status
        async def __aenter__(self): return self
        async def __aexit__(self, *args): return False
        async def read(self): return b""
    class Session:
        closed = False
        def __init__(self): self.calls = 0
        def post(self, *args, **kwargs):
            self.calls += 1
            return Response(statuses[self.calls-1])
    session = Session()
    monkeypatch.setattr(go_callbacks, "_session", session)
    monkeypatch.setattr(go_callbacks.asyncio, "sleep", AsyncMock())
    assert asyncio.run(go_callbacks.post_callback("c1", "turns", {"revision": 1})) is expected
    assert session.calls == calls
