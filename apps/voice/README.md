# EngFlex Voice Engine — Python Pipecat Service

Stateless realtime voice pipeline (STT → LLM → TTS over WebRTC). It owns the
conversation prompts and the audio path; Go only provisions sessions and ingests
results afterwards. Pronunciation scoring lives in the external OpenPronounce
container (called by Go), not here. Always run via `uv run`.

## Layout

```text
src/
  app/            # Runner routes + session schemas
  pipeline/       # STT/LLM/TTS factories, bot tasks, handlers
  transcript/     # Capture + analyze (LLM feedback)
  clients/        # Go callbacks (turn batch ingest)
  prompts.py      # Free-talk + roleplay templates
  bot.py          # Pipeline entry point
  server.py       # Production dispatcher (auth, capacity, drain)
  session.py      # Dispatcher session state (leases, TTL, sweep, drain)
  config.py       # Role-named settings (stt/llm/voice)
scripts/          # Prompt red-team set, fake-mic audio helper
tests/            # Pytest suites
```

## Quickstart

```bash
uv run python src/bot.py   # dev runner (:7860, no auth — local only)
uv run src/server.py  # production dispatcher (:7860, X-Internal-Secret)
uv run pytest -q           # tests
uv run ruff check src tests
```

## Configuration (`.env`)

| Key | Role |
|-----|------|
| `STT_API_KEY` | Speech-to-text |
| `LLM_API_KEY` / `LLM_BASE_URL` / `LLM_MODEL` | Chat model (any OpenAI-compatible endpoint) |
| `TTS_API_KEY` / `VOICE_MODEL` | Speech synthesis model |
| `MALE_VOICE_ID` / `FEMALE_VOICE_ID` | Per-gender voice identity, selected by persona gender |
| `INTERNAL_SECRET` | Authenticates Go callbacks |
| `MAX_DURATION_SEC` / `IDLE_TIMEOUT_SEC` | Session bounds |
| `MAX_SESSIONS` / `MAX_PENDING` | Dispatcher capacity / start-flood guard (pending derives `max(1, MAX_SESSIONS // 4)`) |
| `PENDING_TTL_SEC` | Offer must arrive within, else the lease is reclaimed |
| `DRAIN_GRACE_SEC` | SIGTERM window: live sessions may run out within it |
| `PIPECAT_ICE_SERVERS` | ICE list (JSON array, object, or comma URLs); malformed fails at boot |
| `VOICE_HOST` / `VOICE_PORT` | Dispatcher listener |
| `VOICE_VERSION` | Reported by `GET /status` |
