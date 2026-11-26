"""Wiring tests for the bot entrypoint.

Two independent regression guards, both about import wiring in `bot.py`:

1. Importing `bot` must never pull in `pipecat.runner.run`: that module
   calls load_dotenv(override=True) at import time, which overwrites
   process env (e.g. Playwright's API_URL) with .env values. The dev path
   imports it lazily, AFTER `settings` is constructed.
2. The full dev path (register_dev_routes) mounts the data routes on the
   runner's shared app, and process env still beats .env through it.
"""

import os
import subprocess
import sys
from pathlib import Path

from fastapi import FastAPI
from fastapi.routing import APIRoute

import bot  # noqa: F401 -- imported for the side effect under test

SRC = Path(__file__).resolve().parents[1] / "src"


def test_bot_import_never_pulls_the_dev_runner(tmp_path):
    """The env-poisoning module must not be reachable via a plain import.

    Runs in a subprocess: other test modules legitimately import
    pipecat.runner.run, so the in-process sys.modules would be polluted and
    the assertion meaningless here.
    """
    result = subprocess.run(
        [
            sys.executable,
            "-c",
            "import sys; import bot; print('pipecat.runner.run' in sys.modules)",
        ],
        cwd=tmp_path,
        env={**os.environ, "PYTHONPATH": str(SRC)},
        capture_output=True,
        text=True,
        timeout=120,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert "False" in result.stdout, result.stdout


def test_register_routes_mounts_the_data_routes():
    from app.routes import register_routes

    app = FastAPI()
    register_routes(app)
    paths = {route.path for route in app.routes if isinstance(route, APIRoute)}
    assert {"/analyze", "/transcript", "/transcribe"} <= paths


def test_dev_path_mounts_routes_and_env_beats_dotenv(tmp_path):
    """The dev runner wiring, end to end in a fresh interpreter.

    The temp cwd holds a `.env` that disagrees with the process env on
    purpose: `settings` must be built before register_dev_routes' lazy
    import of pipecat.runner.run lets load_dotenv(override=True) run.
    """
    (tmp_path / ".env").write_text(
        "STT_API_KEY=x\nTTS_API_KEY=y\nINTERNAL_SECRET=z\nAPI_URL=http://from-dotenv:9999\n"
    )
    result = subprocess.run(
        [
            sys.executable,
            "-c",
            (
                "import bot; bot.register_dev_routes(); "
                "from pipecat.runner.run import app as runner_app; "
                "from config import settings; "
                "print(settings.api_url); "
                "print(any(getattr(r, 'path', None) == '/analyze' for r in runner_app.routes))"
            ),
        ],
        cwd=tmp_path,
        env={**os.environ, "PYTHONPATH": str(SRC), "API_URL": "http://from-process:1234"},
        capture_output=True,
        text=True,
        timeout=120,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert "http://from-process:1234" in result.stdout, result.stdout
    assert "True" in result.stdout, result.stdout


def test_level_precedence_learner_over_scenario():
    from prompts import resolve_level

    assert resolve_level("B1", "C1") == "B1"


def test_bot_forwards_persona_scenario():
    from app.schemas import RunnerBody

    body = RunnerBody.model_validate(
        {
            "userId": "u1",
            "conversationId": "c1",
            "maxDuration": 300,
            "persona": {"name": "Amelia", "roleTitle": "Hiring director"},
            "scenario": {"title": "T", "objective": "O.", "cefrLevel": "B2"},
            "learner": {"level": "B1"},
        }
    )
    assert body.persona is not None and body.persona.name == "Amelia"
    assert body.scenario is not None and body.scenario.cefrLevel == "B2"
    assert body.learner is not None and body.learner.level == "B1"
