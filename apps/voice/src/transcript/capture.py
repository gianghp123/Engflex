"""In-memory transcript capture for one session.

The engine is the only writer of transcript turns, so it collects everything
here and posts one batch to Go at finalize. Consecutive user fragments (STT
finals that never triggered a bot reply) collapse into one turn so the
transcript reads like a conversation, not like STT chunks.

Deliberately absent: audio and word-metadata fields. Turn records stay
text-only; acoustic alignment belongs to the OpenPronounce scoring container,
not the transcript.
"""

import json
from collections.abc import Callable
from dataclasses import dataclass, replace


class SnapshotLimitExceeded(ValueError):
    pass


@dataclass
class TurnRecord:
    position: int
    role: str
    text: str
    was_interrupted: bool = False


class TurnCollector:
    def __init__(self) -> None:
        self._records: list[TurnRecord] = []
        self.revision = 1
        self.on_change: Callable[[], None] | None = None
        self.on_limit: Callable[[], None] | None = None
        self.accepting = True
        self.finalized = False
        self._open_user: TurnRecord | None = None
        # Slot a correction reserved for the bot's next reply. It survives an
        # interrupted (dying) reply and is consumed by the first complete one.
        self._reserved_bot_position: int | None = None
        # The turn the modal is looking at, set by review and consumed by send.
        # A position, not a record: the collector owns the records, so a stale
        # mark resolves to None instead of rewriting a resurrected object.
        self._reviewed_position: int | None = None

    def add_user_text(self, text: str, was_interrupted: bool = False) -> TurnRecord | None:
        """Record learner speech.

        Consecutive fragments merge into one turn so the transcript reads like a
        conversation. A correction never comes through here — it goes through
        `begin_correction`, which rewrites an existing turn in place.
        """
        cleaned = (text or "").strip()
        if not cleaned:
            return None
        self._check_limit(cleaned, merging=self._open_user)
        if self._open_user is not None:
            self._open_user.text = f"{self._open_user.text} {cleaned}".strip()
            self._open_user.was_interrupted = self._open_user.was_interrupted or was_interrupted
            self._changed()
            return self._open_user
        record = TurnRecord(
            position=self._next_position(),
            role="user",
            text=cleaned,
            was_interrupted=was_interrupted,
        )
        self._records.append(record)
        self._open_user = record
        self._changed()
        return record

    def begin_correction(self, text: str) -> int | None:
        """Rewrite the latest user turn and reserve the bot's reply slot.

        Returns the corrected turn's position, or None when there is no user
        turn to correct. The reservation is what keeps positions stable: the
        interruption we queue to stop the stale reply records a dying reply
        *after* this call, and the regenerated reply must still land on the
        slot the stale one held rather than one past it — otherwise every
        later position drifts and Go's `(conversation_id, position)` key
        desyncs from the transcript.
        """
        cleaned = (text or "").strip()
        if not cleaned:
            return None
        target: TurnRecord | None = None
        for record in self._records:
            if record.role == "user":
                target = record
        if target is None:
            return None
        self._check_limit(cleaned, replacing=True)
        target.text = cleaned
        # Drop the stale reply and anything recorded after the target turn.
        self._records = [r for r in self._records if r.position <= target.position]
        self._open_user = target
        self._reserved_bot_position = target.position + 1
        self._changed()
        return target.position

    def add_bot_text(self, text: str, was_interrupted: bool = False) -> TurnRecord | None:
        cleaned = (text or "").strip()
        if not cleaned:
            return None
        reserved = self._reserved_bot_position
        if reserved is not None and was_interrupted:
            # A dying reply while a correction holds a slot answered text that
            # no longer exists and is about to be replaced on that very slot.
            # Recording it would only churn a row we are about to overwrite.
            return None
        self._check_limit(cleaned, replacing=reserved is not None)
        self._open_user = None
        slot = reserved if reserved is not None else self._next_position()
        record = TurnRecord(
            position=slot,
            role="ai",
            text=cleaned,
            was_interrupted=was_interrupted,
        )
        if len(self._records) == slot:
            # The reserved slot is already there, from the stale reply.
            self._records[slot - 1] = record
        else:
            self._records.append(record)
        if reserved is not None:
            # This reply claimed the reservation, so the reservation is released.
            self._reserved_bot_position = None
        self._changed()
        return record

    def latest_user_record(self) -> TurnRecord | None:
        """The most recent learner turn, or None when they have not spoken."""
        latest: TurnRecord | None = None
        for record in self._records:
            if record.role == "user":
                latest = record
        return latest

    def records(self) -> list[TurnRecord]:
        return list(self._records)

    def mark_reviewed(self, position: int) -> None:
        """Remember which turn the modal is looking at (set by review)."""
        self._reviewed_position = position

    def clear_reviewed(self) -> None:
        self._reviewed_position = None

    @property
    def reviewed_position(self) -> int | None:
        return self._reviewed_position

    def clear(self) -> None:
        self._records.clear()
        self._open_user = None
        self._reviewed_position = None
        self._changed()

    def _next_position(self) -> int:
        return len(self._records) + 1

    def snapshot(self) -> tuple[int, tuple[TurnRecord, ...]]:
        return self.revision, tuple(replace(r) for r in self._records)

    def _changed(self) -> None:
        self.revision += 1
        if self.on_change is not None:
            self.on_change()

    def _check_limit(self, text: str, *, merging: TurnRecord | None = None,
                     replacing: bool = False) -> None:
        if self.finalized:
            raise SnapshotLimitExceeded("transcript finalized")
        length = len(text) + (len(merging.text) + 1 if merging else 0)
        count = len(self._records) + (0 if merging or replacing else 1)
        # Reserve JSON escaping/metadata overhead conservatively before a
        # mutation, so an oversized snapshot is never silently truncated.
        wire_bytes = sum(len(json.dumps(r.text).encode("utf-8")) for r in self._records)
        wire_bytes += len(json.dumps(text).encode("utf-8")) + count * 128
        if not self.accepting or length > 10000 or count > 500 or wire_bytes > (4 << 20):
            self.accepting = False
            if self.on_limit is not None:
                self.on_limit()
            raise SnapshotLimitExceeded("transcript snapshot limit exceeded")
