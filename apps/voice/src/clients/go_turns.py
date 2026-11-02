"""Coalesce immutable full snapshots; one sender task per voice session."""
import asyncio

from clients.go_callbacks import finalize_session, post_callback
from transcript.capture import TurnCollector


class SnapshotPublisher:
    def __init__(self, conversation_id: str, engine_session_id: str,
                 collector: TurnCollector):
        self.conversation_id = conversation_id
        self.engine_session_id = engine_session_id
        self.collector = collector
        self.pending: dict | None = None
        self.task: asyncio.Task | None = None
        self.closed = False
        self.final_task: asyncio.Task | None = None

    def snapshot(self) -> dict:
        revision, records = self.collector.snapshot()
        return {"engineSessionId": self.engine_session_id, "revision": revision,
                "turns": [{"position": r.position, "role": r.role, "text": r.text,
                           "wasInterrupted": r.was_interrupted} for r in records]}

    def notify(self) -> None:
        if self.closed:
            return
        self.pending = self.snapshot()
        if self.task is None or self.task.done():
            self.task = asyncio.create_task(self._drain())

    async def _drain(self) -> None:
        while self.pending is not None:
            snapshot, self.pending = self.pending, None
            await post_callback(self.conversation_id, "turns", snapshot)

    async def finalize(self, duration_sec: int) -> bool:
        if self.final_task is None:
            self.closed = True
            self.collector.finalized = True
            self.pending = None
            snapshot = self.snapshot()
            async def flush():
                if self.task is not None:
                    await self.task
                return await finalize_session(self.conversation_id, duration_sec, snapshot)
            self.final_task = asyncio.create_task(flush())
        return await self.final_task
