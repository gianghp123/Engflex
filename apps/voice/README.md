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
  config.py       # Role-named settings (stt/llm/voice)
scripts/          # Prompt red-team set, fake-mic audio helper
tests/            # Pytest suites
```

## Quickstart

```bash
uv run python src/bot.py   # entry point (:7860)
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
