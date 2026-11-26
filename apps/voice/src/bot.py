import asyncio
import os
import time

# NOTE (import order is structural, not lucky).
# `settings` must be constructed before anything pulls in pipecat.runner.run,
# because that module calls load_dotenv(override=True) at import time, which
# overwrites process-supplied env (e.g. Playwright's API_URL) with .env values.
# The dev path (register_dev_routes / __main__ below) imports that module
# lazily, AFTER the config import above has run; importing `bot` itself never
# pulls the runner in. tests/test_bot_routes.py pins both properties.
# isort: off
from config import settings  # must precede every pipecat import, see above
# isort: on

from loguru import logger
from pipecat.runner.types import RunnerArguments, SmallWebRTCRunnerArguments
from pipecat.transports.base_transport import BaseTransport, TransportParams
from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport

from app.routes import register_routes
from app.schemas import RunnerBody
from events import register_event_handlers
from observability import bind_session, init_sentry
from pipeline.services import build_context
from pipeline.tasks import create_voice_bot_worker
from prompts import resolve_level
from session import SessionRegistry
from transcript.capture import TurnCollector
from transcript.sessions import SessionControl, register, unregister

init_sentry()

# Non-secret effective config, one line at boot. Values are interpolated into
# the message (not kwargs) because the log format does not render extras.
# The env-vs-settings pair catches env poisoning (see the import-order note
# above) directly in the run log.
logger.info(
    f"engine config api_url={settings.api_url} "
    f"llm={settings.llm_model} sentry={bool(settings.sentry_dsn)} "
    f"env_API_URL={os.environ.get('API_URL')!r}"
)


async def bot(runner_args: RunnerArguments):
    if not isinstance(runner_args, SmallWebRTCRunnerArguments):
        logger.error("unsupported runner args", kind=type(runner_args).__name__)
        return
    if not runner_args.body:
        raise ValueError("runner body is required")

    body = RunnerBody.model_validate(runner_args.body)
    log = bind_session(body.userId, body.conversationId)
    log.info("session starting", max_duration=body.maxDuration)

    transport: BaseTransport = SmallWebRTCTransport(
        webrtc_connection=runner_args.webrtc_connection,
        params=TransportParams(audio_in_enabled=True, audio_out_enabled=True),
    )

    start_time = time.time()
    context = build_context()
    collector = TurnCollector()
    level = resolve_level(
        body.learner.level if body.learner else None,
        body.scenario.cefrLevel if body.scenario else None,
    )
    worker, user_aggregator, assistant_aggregator = await create_voice_bot_worker(
        transport,
        context,
        max_duration=body.maxDuration,
        conversation_id=body.conversationId,
        collector=collector,
        persona=body.persona,
        scenario=body.scenario,
        level=level,
    )
    register(
        body.conversationId,
        SessionControl(
            conversation_id=body.conversationId,
            worker=worker,
            collector=collector,
        ),
    )
    try:
        register_event_handlers(
            worker,
            transport,
            user_aggregator,
            assistant_aggregator,
            body=body,
            start_time=start_time,
            collector=collector,
        )

        from pipecat.workers.runner import WorkerRunner

        runner = WorkerRunner(handle_sigint=False)
        await runner.add_workers(worker)
        await runner.run()
        log.info("session finished", duration_sec=int(time.time() - start_time))
    finally:
        # A crashed session must not leave a handle behind: the transcript
        # routes would then act on a dead pipeline.
        unregister(body.conversationId)


async def run_session(runner_args: RunnerArguments, *, registry: SessionRegistry) -> None:
    """Run one dispatcher session: the pipeline, then guaranteed cleanup.

    server.py's offer callback starts this as a task. Whatever bot() does —
    clean finish, crash, cancellation — the connection is closed, the lease
    released, and the live entry removed before the task ends.
    """
    connection = getattr(runner_args, "webrtc_connection", None)
    reason = "finished"
    try:
        await bot(runner_args)
    except asyncio.CancelledError:
        reason = "cancelled"
        raise
    except Exception:  # noqa: BLE001 - logged here so the task never dies silent
        reason = "error"
        logger.exception("session crashed", session_id=runner_args.session_id)
    finally:
        if connection is not None:
            try:
                if connection.pc.connectionState != "closed":
                    await connection.disconnect()
            except Exception:  # noqa: BLE001 - cleanup must never mask the reason
                logger.warning("connection close failed", session_id=runner_args.session_id)
        registry.end_live(runner_args.session_id, reason)


def register_dev_routes() -> None:
    """Mount the data routes on the pipecat dev runner's shared app.

    Lazy on purpose: importing pipecat.runner.run here (after `settings` was
    constructed at module import above) is what keeps load_dotenv from
    poisoning the process env. The production dispatcher never calls this.
    """
    # Named for what it is: the runner's FastAPI singleton, not our app package.
    from pipecat.runner.run import app as runner_app

    register_routes(runner_app)


if __name__ == "__main__":
    register_dev_routes()

    from pipecat.runner.run import main

    main()
