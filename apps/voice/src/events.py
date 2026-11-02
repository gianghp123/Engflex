"""Pipeline and transport event wiring: greeting, session timer, errors, finalize.

Deviation from the brief (verified against installed pipecat-ai 1.11.0):
``on_client_connected`` / ``on_client_disconnected`` are transport events
(``SmallWebRTCTransport`` registers them; ``PipelineWorker`` only registers
``on_pipeline_*`` / ``on_frame_*`` / ``on_*_timeout`` and warns + drops
anything else). So this module takes the ``transport`` explicitly and
registers the client handlers on it. Behaviour is unchanged: greeting on
connect, session timer, fatal error surface, finalize callback on finish.
"""

import asyncio
import re
import time

from loguru import logger
from pipecat.frames.frames import EndTaskFrame, EndWorkerFrame, ErrorFrame, LLMRunFrame
from pipecat.pipeline.worker import PipelineWorker
from pipecat.processors.frame_processor import FrameDirection
from pipecat.transports.base_transport import BaseTransport

from app.schemas import RunnerBody
from clients.go_turns import SnapshotPublisher
from transcript.capture import SnapshotLimitExceeded, TurnCollector


def get_custom_error_message(error: str, service_name: str = "Service") -> str:
    text = str(error)
    if "402" in text:
        return f"{service_name} unavailable: payment required."
    if "429" in text:
        return f"{service_name} unavailable: too many requests, try again soon."
    if "401" in text or "403" in text:
        return f"{service_name} unavailable: authentication failed."
    if "timeout" in text.lower() or "408" in text:
        return f"{service_name} timed out. Please try again."
    return f"{service_name} unavailable: an unexpected error occurred."


# The only string ever surfaced to the client. Provider codes, key states
# and timeouts are operator detail — they stay in the logs via
# get_custom_error_message. Learners get one neutral line regardless of
# cause, so no backend internals leak into the room modal.
USER_FACING_ERROR = "The voice service is temporarily unavailable. Please try again."


# Permanent failures never recover on their own: bad/expired keys (401),
# missing entitlements (403), unpaid bills (402), unknown models. Everything
# else (429 rate limits, timeouts, 5xx, transport blips) may clear, so it
# stays non-fatal and the session survives. Word boundaries keep codes like
# 1403 from matching.
_PERMANENT_ERROR = re.compile(
    r"\b40[01234]\b|invalid_api_key|model_not_found|permission_denied|permission denied|subscription",
    re.IGNORECASE,
)


def is_permanent_error(error: str) -> bool:
    """True when the session can never succeed without human action."""
    return _PERMANENT_ERROR.search(str(error)) is not None


def register_event_handlers(
    worker: PipelineWorker,
    transport: BaseTransport,
    user_aggregator,
    assistant_aggregator,
    *,
    body: RunnerBody,
    start_time: float,
    collector: TurnCollector,
    publisher: SnapshotPublisher,
) -> None:
    timer_handle: dict[str, asyncio.Task[None]] = {}

    @transport.event_handler("on_client_connected")
    async def on_client_connected(_transport: BaseTransport, _client):
        logger.info("client connected", conversation_id=body.conversationId)
        # No context priming here: the session prompt travels via the LLM
        # service's system_instruction and the opener is seeded in
        # build_context, so the first run is exactly one system + one user
        # message — the shape this upstream accepts.
        await worker.queue_frames([LLMRunFrame()])

        async def session_timer():
            await asyncio.sleep(body.maxDuration)
            logger.info("max duration reached", conversation_id=body.conversationId)
            await worker.queue_frame(EndTaskFrame(), FrameDirection.UPSTREAM)

        timer_handle["task"] = asyncio.create_task(session_timer())

    @transport.event_handler("on_client_disconnected")
    async def on_client_disconnected(_transport: BaseTransport, _client):
        logger.info("client disconnected", conversation_id=body.conversationId)
        if (pending := timer_handle.get("task")) is not None:
            pending.cancel()
        await worker.cancel()

    @worker.event_handler("on_pipeline_error")
    async def on_pipeline_error(_worker: PipelineWorker, frame):
        error_text = frame.error if isinstance(frame, ErrorFrame) else str(frame)
        # Transient blips (e.g. a single TTS context completing with no
        # audio on cold start) arrive with fatal=False and recover on their
        # own — pipecat only marks a processor unusable after repeated
        # ones. Ending the session here killed sessions that had already
        # recovered, so non-fatal frames are logged and left alone; only
        # fatal frames surface to the client and end the pipeline.
        # Note: the RTVI processor still forwards every ErrorFrame to the
        # client, including these non-fatal ones — the frontend's error
        # listener ignores fatal:false and only modals on fatal/unknown.
        if isinstance(frame, ErrorFrame) and not frame.fatal:
            # Permanent failures (auth, payment, unknown model) would leave
            # the session mute until the duration timer fires, so escalate
            # those through the fatal path below — sona-voice style, but
            # selective: blanket escalation killed recovered sessions (see
            # tests/test_pipeline_errors.py).
            if is_permanent_error(error_text):
                message = get_custom_error_message(error_text)
                logger.error(
                    "permanent pipeline error; ending session", error=error_text, message=message
                )
                await worker.queue_frames([ErrorFrame(USER_FACING_ERROR), EndWorkerFrame()])
                return
            logger.warning("transient pipeline error; session continues", error=error_text)
            return
        message = get_custom_error_message(error_text)
        logger.error("pipeline error", error=error_text, message=message)
        # No `fatal=True` (deprecated since 1.8.0): the RTVI processor forwards
        # every ErrorFrame to the client, and EndWorkerFrame ends the pipeline
        # so on_pipeline_finished still finalizes the session.
        await worker.queue_frames([ErrorFrame(USER_FACING_ERROR), EndWorkerFrame()])

    @user_aggregator.event_handler("on_user_turn_message_added")
    async def on_user_turn(_aggregator, message) -> None:
        # message.content is ALWAYS populated here. Do NOT use
        # on_user_turn_stopped for text: its content is None in realtime mode.
        # User turns take the default was_interrupted=False: the user-side
        # message type has no interruption field, and consecutive fragments
        # already merge in the collector.
        try:
            collector.add_user_text(message.content)
        except SnapshotLimitExceeded:
            await worker.queue_frames([ErrorFrame(USER_FACING_ERROR), EndWorkerFrame()])

    @assistant_aggregator.event_handler("on_assistant_turn_stopped")
    async def on_assistant_turn(_aggregator, message) -> None:
        # message.content may be empty on a pre-token interruption;
        # message.interrupted carries the flag (no separate frame watch).
        # InterruptionFrame exists in pipecat.frames.frames; there is no
        # StartInterruptionFrame in 1.11.0.
        try:
            collector.add_bot_text(message.content, was_interrupted=message.interrupted)
        except SnapshotLimitExceeded:
            await worker.queue_frames([ErrorFrame(USER_FACING_ERROR), EndWorkerFrame()])

    @worker.event_handler("on_pipeline_finished")
    async def on_pipeline_finished(_worker: PipelineWorker, _frame):
        if (pending := timer_handle.get("task")) is not None:
            pending.cancel()
        duration = max(0, int(time.time() - start_time))
        await publisher.finalize(duration)
