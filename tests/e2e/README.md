# E2E (Playwright, live backend)

Voice-room flows against a real Go API + Postgres. The voice engine is
**not** started: `VOICE_SERVICE_URL` points at a dead port so
setup-failure paths are deterministic. True mid-call WebRTC disconnects
need a live engine + media and are out of scope.

## Prerequisites

`playwright.config.ts` loads `tests/e2e/.env` (then `apps/web/.env.local`
as fallback) via Vite's `loadEnv`, so put e2e values in `tests/e2e/.env` —
shell env still wins on conflict. Only one value has no file fallback:

```sh
# tests/e2e/.env (already gitignored; see .env.example)
VITE_CLERK_PUBLISHABLE_KEY=...   # also used as CLERK_PUBLISHABLE_KEY
CLERK_SECRET_KEY=...             # same Clerk instance as the apps
# Dedicated e2e database, never the dev one (see below).
DATABASE_URL=postgres://postgres:postgres@localhost:5777/engflex_e2e?sslmode=disable
E2E_CLERK_USER_EMAIL="e2e+clerk_test@…"  # pre-created test user, nothing else works without it
```

1. Postgres running (no manual migrate needed):
   ```sh
   cd apps/api && docker compose up -d postgres
   ```

`DATABASE_URL` must point at a **dedicated e2e database**. Before the Go API
boots, `ensure-e2e-db.mjs` (chained onto the API's `webServer.command`)
creates it if missing and migrates it, and the config points the server at
it via derived `DB_*` vars. It has to live in the webServer command and not
in a setup hook: Playwright starts webServers *before* every setup hook, and
the API refuses to boot without its database. Two reasons for the dedicated
DB: tests must never touch dev data, and goose tracks applied versions —
edited migrations never re-run on an already-migrated dev DB, which is
exactly the stale-schema failure that bit the first run (`relation
"feedbacks" does not exist`).

> Makefile trap: `apps/api/Makefile` must not `include .env` — a makefile
> assignment overrides the environment, so the `.env` `DATABASE_URL` would
> silently clobber the e2e one and the suite would migrate the dev database
> while the API boots against an empty e2e DB (`relation "conversations"
> does not exist`). The Makefile reads only the `DATABASE_URL` line with
> `?=` instead.

## Run

```sh
cd tests/e2e
pnpm test:e2e            # headless, boots Go API + vite automatically
pnpm test:e2e:headed     # headed browser (debugging)
pnpm test:e2e:live       # live engine + fake mic (Analyze flow)
pnpm test:e2e:recovery   # live engine + fake mic (transcript recovery flow)
```

The Go server starts via `webServer` with short voice timeouts
(`VOICE_*_TIMEOUT_MS`) so failure cases resolve in seconds. Tests create
real conversations through the API as the signed-in Clerk user and delete
them afterwards; nothing else in the database is touched.

## Live suites (`playwright.live.config.ts`, `playwright.recovery.config.ts`)

Two live suites, one shared stack (`live-stack.ts`) with dedicated ports
(API 8010, web 3010, engine 7861) and `reuseExistingServer: false`
everywhere: a dev server on 8000 carries its own `VOICE_SERVICE_URL` and would
silently make the run prove nothing. Extra prerequisites: `apps/voice/.env`
with real `STT_API_KEY`/`LLM_API_KEY` and the voice model in `models/`.

- `voice-live.spec.ts` (`pnpm test:e2e:live`): a real engine, a real WebRTC
  session, and a real Analyze round trip — fake mic → Deepgram → collector →
  live turn post → Go poll → Analyze button on the live row → engine
  `/analyze` → diagnostics card, with DB assertions that the feedback is keyed
  by the learner turn's UUID at the matching position. Nothing about recovery
  is forced here, so this suite also proves the correction feature leaves normal
  turns completely alone.
- `voice-recovery.spec.ts` (`pnpm test:e2e:recovery`): on-demand transcript
  correction, driven exactly as a learner drives it — the fake-mic turn commits
  on its own with **no card** (nothing is forced, so this also proves no signal
  interrupts the learner mid-sentence) → **Review transcript** opens the card →
  an edit rewrites that turn *in place* at the same position (DB ground truth)
  and the panel redraws to match → **Speak again** makes the engine listen while
  the field stays editable, a spoken attempt becomes the field, a second attempt
  is accepted, and sending corrects the turn in place again → **Dismiss** closes
  the window and normal turns resume.

The two configs are now identical: no engine env override exists, because the
learner is the only trigger. Nothing in the engine decides on the learner's
behalf that a turn was misheard — a backend uncertainty signal is not a reason
to interrupt someone mid-sentence — so there is no engine-side switch to force.

Fake mic: Chromium's built-in fake capture is a 440 Hz tone, which STT
rejects, so the suite feeds `--use-file-for-fake-audio-capture` a WAV
synthesized offline by `apps/voice/scripts/make_fake_mic_audio.py`
(a local Piper voice, no network and no key — the engine's own TTS is
independent of it), generated into `playwright/.fake-mic/` (gitignored) on
first run. Flag pairing matters: `--use-fake-device-for-media-stream` selects
the fake device (without it Chrome uses the real mic and the file flag is
silently ignored — this cost several runs, which transcribed the room:
TOEIC practice, narration, silence), `--use-fake-ui-for-media-stream` only
auto-grants the permission prompt. Load-bearing fixture details:

- **48 kHz mono 16-bit.** Chromium's capture pipeline runs at 48 kHz; feeding
  the voice's native 22050 Hz plays ~2.2x fast and STT hears garbage
  (observed: "on contact we check it right" for "I go to the office
  yesterday"). The script resamples (linear, stdlib-only).
- **60 s of leading digital silence.** The first in-session instance must land
  after the greeting (LLM + TTS, up to 60 s+ on a slow proxied endpoint) and
  after Chrome's AEC/AGC and the Opus encoder converge: a VAD episode
  overlapping bot playback transcribed to nothing (likely echo-mix). A
  −45 dBFS comfort-noise floor was tried first and abandoned: streaming STT
  hallucinated a sustained poetic narrative on it — a new "turn" every few
  seconds — drowning the real instances, while absolute silence yields finals
  only on speech (plus Opus DTX sends no packets for it). The sentence starts
  with a sacrificial "Well," (onset clipping eats the filler) and repeats 3×
  with 3 s gaps; the survivors carry "yesterday". Ordering needs no luck: the
  test asserts group ordinal == DB position structurally, never "greeting is
  position 1".
- **Looped playback (no `%noloop`).** The file starts playing when the mic
  track is acquired, well before the session's STT pipeline is listening — a
  play-once file speaks into the void and the engine then sits in "No audio
  frame received" (observed). Repetitions become separate user turns; the test
  analyzes the first clean one.

UI assertions match loosely (`yesterday`, not the full sentence): streaming
STT returns lowercase unpunctuated text, and a mangled fragment can precede
the clean transcription. The test clicks the first row *containing* the anchor
word, records its sibling ordinal, and the DB asserts the feedback sits on the
turn at exactly that position with a matching UUID key. No assumption that the
greeting is position 1 — file-playback phase (browser launch vs. session
start, ~20 s apart) decides the interleaving, so the invariant is checked
structurally (group ordinal == DB position) instead of positionally.

## Debugging a live failure

`live-boundaries.mjs <run-log> <conversation-id>` prints one verdict per
pipeline boundary from the run's combined log plus the DB:

```text
session → webrtc → pcm → vad → stt-text → db → analyze
```

The conversation id is printed by the spec as `e2e: live conversation <id>`.
Failed live conversations are kept (not deleted) so the `db`/`analyze` rows
stay queryable. The table names the first failing link with evidence — every
fix below was found this way.

## Known traps (each cost at least one live run)

These are load-bearing; re-read before touching the suite, the fixture
script, or the engine's boot path.

1. **Fake device flag.** `--use-file-for-fake-audio-capture` is silently
   ignored without `--use-fake-device-for-media-stream` — Chrome uses the
   real mic and the engine transcribes the room. (`--use-fake-ui-…` only
   grants permission.)
2. **Side-effect import.** `bot.py`'s routes import registers `/analyze`; an
   over-eager `ruff check --fix` once deleted it as F401-unused and every
   Analyze returned 503. Guarded by `apps/voice/tests/test_bot_routes.py`.
3. **System-only analysis requests.** The upstream provider 400s chat
   requests with no user message (same constraint as the session's `"Hello!"`
   opener), so `analysis.py` appends the utterance as a user message.
   A live 503 on Analyze with no engine warning → check this first.
4. **Import order poisons engine env.** `pipecat.runner.run` calls
   `load_dotenv(override=True)` at import time, clobbering process env
   (e.g. Playwright's `API_URL`) with `.env` values. `from config import
   settings` must stay the first local import in `bot.py` (commented there);
   the boot-time `engine config` log line prints settings next to raw env as
   a tripwire — `api_url=…8000` + `sentry=True` under the live config means
   this regressed.
5. **Comfort noise hallucinates.** A −45 dBFS floor made streaming STT emit
   sustained narrative turns; digital silence yields finals only on speech.
6. **Makefile `include .env`.** Same class of bug on the Go side (see the
   trap note under Prerequisites).
7. **Phase is uncontrolled.** File playback starts at browser launch; the
   session starts ~20 s later. Never assert absolute positions — the
   structural invariant is group ordinal == DB position.
8. **`process_frame` must call `super()`.** A custom `FrameProcessor` that
   overrides `process_frame` without `await super().process_frame(frame,
   direction)` never runs the base `StartFrame` handler, so it never creates
   its own process task: its input queue is never drained and it silently
   swallows every frame it is handed. Symptom: STT finals *and* the greeting
   `LLMRunFrame` vanish with no error anywhere, and audio still arrives (the
   VAD/STT services upstream keep working). Found by the recovery suite.
9. **The raw STT text commits on the normal path.** Only the turn under review
   is rewritten. Later positions are loop audio and are *supposed* to commit, so
   assert the correction at position 1, never "no turn carries the raw text".
10. **Barge-in is normal.** The looped fixture keeps "speaking" after a turn
   ends, so a bot reply can be cut off mid-sentence by the next VAD episode.
   Assert conversation outcomes in the DB, not transient UI rows.

## Layout

- `global.setup.ts` — `clerkSetup()`, server-side sign-in, storage state
  to `playwright/.clerk/user.json` (gitignored).
- `ensure-e2e-db.mjs` — create + migrate the e2e DB, chained onto the API's
  `webServer.command` (see above for why it cannot be a setup step).
- `helpers.ts` — authed API calls via server-minted Clerk JWTs, turn/feedback
  seeding and cleanup over `DATABASE_URL` (`pg`), plus `dbListTurns` /
  `dbLearnerTurns` read-backs for the live suites.
- `live-stack.ts` — shared live stack (env loading, ports, DB parts, Chrome
  flags, `defineLiveConfig`) for both live configs.
- `live-boundaries.mjs` — boundary report (see Debugging above).
- `voice-room.spec.ts` — service-fail modal, ended-session review, analyze
  error + seeded feedback card.
- `voice-live.spec.ts` + `fake-mic.ts` + `playwright.live.config.ts` — the
  live Analyze suite above.
- `voice-recovery.spec.ts` + `playwright.recovery.config.ts` — the live
  recovery suite above.

Desktop Chromium only, one worker (serial against the live backend).
Default suite uses the tone-based fake device
(`--use-fake-device-for-media-stream` / `--use-fake-ui-for-media-stream`) so
`initDevicesOnMount` never blocks; both live suites use the file-based fake
audio capture instead.
