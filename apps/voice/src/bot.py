import os
import time

# NOTE (import order is load-bearing, and now ENFORCED rather than lucky).
# `settings` must be constructed before anything pulls in pipecat.runner.run,
# because that module calls load_dotenv(override=True) at import time, which
# overwrites process-supplied env (e.g. Playwright's API_URL) with .env values.
# This used to rely on isort: `from config import ...` sorted before the other
# from-imports. Renaming the routes module to `app` broke that silently — `app`
# sorts before `config`, so the side-effect import ran first and the live e2e
# silently talked to .env's API_URL instead of the test's. Hence the isort: off
# region below, and the regression test in tests/test_bot_routes.py.
# isort: off
from config import settings  # must precede every pipecat import, see above
# isort: on

from loguru import logger
from pipecat.runner.types import RunnerArguments, SmallWebRTCRunnerArguments
from pipecat.transports.base_transport import BaseTransport, TransportParams
from pipecat.transports.smallwebrtc.transport import SmallWebRTCTransport

from app.routes import analyze  # noqa: F401 -- side effect: registers /analyze
from app.schemas import RunnerBody
from clients.go_turns import SnapshotPublisher
from events import register_event_handlers
from observability import bind_session, init_sentry
from pipeline.services import build_context
from pipeline.tasks import create_voice_bot_worker
from prompts import resolve_level
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
    publisher = SnapshotPublisher(body.conversationId, runner_args.session_id, collector)
    collector.on_change = publisher.notify
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
            publisher=publisher,
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
            publisher=publisher,
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


if __name__ == "__main__":
    from pipecat.runner.run import main

    main()
