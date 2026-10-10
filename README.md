# 🎙️ EngFlex — AI English Learning Platform

<div align="center">
  <img src="apps/web/public/logo-lockup.svg" alt="EngFlex — Mo logo and wordmark" width="280" style="display:block;margin-left:auto;margin-right:auto;" />
</div>

<p align="center">
  <strong>Learn from content. Speak with AI. Track your progress.</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/status-active-brightgreen" alt="Status" />
  <img src="https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/python-%3E%3D3.12-3776AB?logo=python" alt="Python" />
  <img src="https://img.shields.io/badge/react-19-61DAFB?logo=react" alt="React" />
</p>

---

> **EngFlex is under active development.** Graduation project (team of 3) unifying
> dictation/shadowing practice and AI voice conversation around one learner
> profile and shared data layer. See [Roadmap](#roadmap) for scope.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Vite + React 19, TypeScript, TanStack Start/Router/Query/Store, shadcn/ui + Tailwind CSS |
| **Auth** | Clerk Authentication |
| **Voice Client** | Pipecat JS client over WebRTC, direct FE ↔ voice engine (audio never proxies through Go) |
| **API Backend** | Go + Gin + GORM |
| **Database** | PostgreSQL, Goose migrations (SQL is the only schema mechanism — never AutoMigrate) |
| **Contracts** | tygo-generated TypeScript (Go DTOs are the source of truth) |
| **Speech Pipeline** | Python + Pipecat (uv-managed, stateless) |
| **STT** | Deepgram |
| **LLM** | OpenAI-compatible API |
| **TTS** | Fish Audio (per-persona gender voice selection) |
| **Pronunciation** | Acoustic scoring on the analyze path |
| **i18n** | Paraglide + inlang message format, locales `vi` (default) + `en` |
| **Mocks** | mockery v3 (generated only) |
| **API Docs** | Swagger |

---

## Key Features

| Feature | Description |
|---------|-------------|
| **Real-time voice chat** | Deepgram STT → LLM → Fish Audio TTS over WebRTC (Pipecat) |
| **Free talk + roleplay** | Casual Flexi tutor or fixed personas (prompt-only, no hidden state) with boundary rules + live off-topic redirect |
| **On-demand feedback** | Per-turn analysis (grammar spans, relevance, alternatives, coaching tip) — async, never pauses the conversation |
| **Pronunciation scoring** | Acoustic assessment attached to analyzed turns |
| **Lessons** | Reading + writing (dictation engine), reusable across modules |
| **Video side-quests** | Self-selected video dictation + shadowing |
| **Vocabulary** | Save words from lessons/videos/conversations + manual add, basic SRS |
| **Progress** | Content list ordered by completion status (unfinished first) |
| **Onboarding/profile** | Level, goals, interests, study time → drives persona/context selection |
| **Gendered TTS voices** | Voice persona based on gender |
| **EN/VI interface** | Cookie-strategy locale, Vietnamese default |

---

## Deployment (Self-Hosted on Coolify)

Two deployment topologies were designed; this repository implements the
**self-hosted** one: a single machine running Coolify (Docker + Traefik),
provisioned end-to-end with Terraform. The
managed-services alternative is kept below as a design reference.

<div align="center">
  <img src="images/coolify_infra.png" alt="Self-hosted (Coolify) deployment topology" width="760" />
  <p><em>Self-hosted topology</em></p>
</div>

<div align="center">
  <img src="images/cloud_native_infra.png" alt="Cloud-native alternative topology: Vercel + Neon + Modal + ExpressTURN" width="760" />
  <p><em>Cloud-native</em></p>
</div>


### What `make deploy` runs

```text
terraform apply              # Coolify project, Postgres, Redis, MinIO service, 4 apps, env vars, domains
→ goose migrations           # host-run; SQL migrations are the only schema mechanism
→ ensure pronounce artifact  # build + push the frozen OpenPronounce image when missing or when the recipe changed
→ trigger 4 app deploys      # api, web, voice (+ coturn relay), pronounce — built from the local git mirror
→ health polls               # api/web external; voice/pronounce internal
```

### Prerequisites

- Coolify at `http://localhost:8000` with an API token (Coolify UI → Security →
  API Tokens; database/service creation routes require `deploy` on the token).
- Tools: `terraform` (>= 1.6), `goose` (`~/go/bin/goose`), `mkcert` (HTTPS), Docker.
- `openssl rand -hex 24` for generating the secret values below.

### Bootstrap

```bash
cp infrastructure/environments/selfhost/terraform.tfvars.example \
   infrastructure/environments/selfhost/terraform.tfvars

make deploy        # terraform apply → goose migrations → ensure pronounce artifact → trigger api/web/voice/pronounce → health polls
make deploy-seed   # create the lesson-audio bucket, open anonymous reads, TTS-seed audio
```

### Ops notes

- Sentry projects + DSNs are provisioned by Terraform (`infrastructure/modules/sentry`;
  projects `engflex-{api,web,voice}-production` in org `giang-pham-90`). The provider
  auth token is an org-scoped internal-integration token kept only in the gitignored
  `terraform.tfvars`; `make deploy` provisions/refreshes the projects and injects the
  DSNs into the app env vars (web's bakes at image build).

---

## Architecture Overview

> Diagrams below are Mermaid sources. To get the hand-drawn look, paste any
> block into Excalidraw via its diagram tool (Mermaid-to-Excalidraw import).

```mermaid
flowchart TB
    FE["Frontend<br/>Vite + React + TanStack<br/>Clerk + shadcn/ui"]
    BE["API control plane<br/>Go + Gin + GORM"]
    DB[("PostgreSQL<br/>goose migrations")]
    CLERK["Clerk Auth"]
    AIV["Voice engine<br/>Python + Pipecat<br/>stateless"]
    STT["Deepgram STT"]
    LLM["LLM API<br/>OpenAI-compatible"]
    TTS["Fish Audio TTS"]

    FE -->|"REST/JSON: business logics"| BE
    BE --> DB
    BE -->|"verify session"| CLERK
    BE -->|"provision: persona / context / token"| AIV
    FE <-->|"WebRTC direct, realtime audio<br/>"| AIV
    AIV --> STT
    AIV --> LLM
    AIV --> TTS
    AIV -->|"transcript + feedback<br/>(async batch)"| BE
```

### Voice Session Lifecycle

```mermaid
sequenceDiagram
    participant L as Learner
    participant W as Web FE
    participant A as API Go
    participant V as Voice Engine
    participant AI as AI Services

    L->>W: Open voice session
    W->>A: POST conversations
    A->>V: Provision session
    V-->>W: WebRTC offer and token
    W->>V: WebRTC audio
    V-->>W: WebRTC audio

    loop Every turn
        V->>AI: STT
        AI-->>V: Transcript
        V->>AI: LLM roleplay
        AI-->>V: Response
        V->>AI: TTS
        AI-->>V: Audio
        V-->>L: Spoken reply
    end

    L->>W: Tap feedback
    W->>V: Analyze utterance
    V->>AI: One shot analysis
    AI-->>V: Feedback result
    V-->>A: Feedback payload
    A->>A: Store feedback and update progress
```

---

## Project Structure

```
engflex-ultimate/
├── apps/api/                       # Go control plane (Gin + GORM) — see apps/api/README.md
│   ├── cmd/server/main.go          # Entrypoint: wires config + router only
│   ├── config/                     # Auth, cors, database, log, server, voice
│   ├── internal/
│   │   ├── common/                 # AppError, ApiResponse envelope, enums, limits
│   │   ├── database/
│   │   │   ├── migrations/         # Goose SQL — source of truth (one migration per table)
│   │   │   ├── models/             # GORM structs + TableName() (no gorm.Model)
│   │   │   └── repositories/       # One repo per table + mocks/
│   │   ├── logger/                 # slog setup (request_id + user_id in ctx)
│   │   ├── modules/<domain>/       # attempts, conversations, lessons, profiles, scenarios, videos, vocabulary
│   │   │   ├── module.go           # RegisterRoutes — wiring only
│   │   │   ├── controllers/        # Gin handlers, no SQL
│   │   │   ├── services/           # Business logic, returns *Model + error
│   │   │   └── dtos/requests|responses  # Wire DTOs (tygo inputs)
│   │   ├── server/middleware/      # RequireAuth, RequestID, RequestLogger, Recovery
│   │   └── utils/                  # Map/MapSlice, pagination, IDs
│   ├── tygo.yaml                   # Go DTOs → packages/contracts
│   └── Makefile                    # migrate, mocks, contracts, swagger
│
├── apps/voice/                     # Python Pipecat engine (uv-managed, stateless) — see apps/voice/README.md
│   ├── src/
│   │   ├── app/                    # Runner routes + session schemas
│   │   ├── pipeline/               # STT/LLM/TTS factories, tasks, handlers
│   │   ├── transcript/             # Capture + analyze (feedback, pronunciation)
│   │   ├── clients/                # Go callbacks (turn batch ingest)
│   │   ├── prompts.py              # Free-talk + roleplay templates (engine owns them)
│   │   ├── bot.py                  # Pipeline entry point
│   │   └── config.py               # Role-named settings (stt/llm/voice, never vendors)
│   ├── scripts/                    # redteam_roleplay.py, make_fake_mic_audio.py
│   ├── tests/                      # Per-feature suites (pytest, 120+ tests)
│   └── pyproject.toml              # uv deps (torch CPU index)
│
├── apps/web/                       # TanStack Start frontend — see apps/web/README.md
│   ├── src/
│   │   ├── routes/                 # Directory convention (_app/route.tsx, _app/<domain>/…)
│   │   ├── features/<domain>/      # attempts, dashboard, dictation, lessons, onboarding, …
│   │   ├── components/common|layout|ui  # Shared + shadcn primitives
│   │   ├── app/                    # APP_ROUTES + API_ROUTES (only URLs allowed)
│   │   ├── lib/                    # api.ts envelope client, auth-guard, utils
│   │   └── paraglide/              # Generated-only i18n output, never hand-edit
│   ├── messages/{en,vi}/<domain>.json  # inlang sources (vi is base locale)
│   └── scripts/check-conventions.mjs   # Message ids, dynamic keys, hardcoded URLs
│
└── packages/contracts/             # Generated TS (tygo output) + hand-written index.ts barrel
```

---

## Installation

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | >= 1.25 | API backend |
| [uv](https://docs.astral.sh/uv/) | latest | Voice engine deps + runner |
| [Node.js](https://nodejs.org/) | >= 20 | Web frontend |
| [pnpm](https://pnpm.io/) | >= 9 | Monorepo workspaces |
| [Docker](https://docs.docker.com/get-docker/) | Compose v2 | Dev infrastructure containers (Postgres, Redis, MinIO) + deployments |
| [PostgreSQL](https://www.postgresql.org/) | 15+ | Database |
| [goose](https://github.com/pressly/goose) | v3 | Migrations |
| [mockery](https://vektra.github.io/mockery/) | v3 | Repo mocks |

---

## Environment Variables

Each service reads its own `.env` — copy from `.env.example` and fill in values
(never commit real keys):

| Service | Notable keys |
|---------|--------------|
| API (`apps/api`) | `DATABASE_URL`, Clerk keys, voice-engine URL + internal secret |
| Voice (`apps/voice`) | `STT_API_KEY`, `LLM_API_KEY` / `LLM_BASE_URL` / `LLM_MODEL`, `TTS_API_KEY`, `VOICE_MODEL`, `MALE_VOICE_ID` / `FEMALE_VOICE_ID`, `INTERNAL_SECRET` |
| Web (`apps/web`) | Clerk publishable key, API base URL |

---

## Running Locally

The whole dev stack runs on your machine: a small Docker compose stack
(Postgres, Redis, MinIO, optional pronunciation model host) plus the three app
processes.

### 1. Workspace dependencies (once)

```bash
pnpm install               # from the repo root (pnpm workspace)
```

### 2. Dev infrastructure (Docker)

```bash
cd apps/api
cp .env.example .env       # compose reads DB_* / MINIO_* / REDIS_* from here too
docker compose up -d postgres redis minio
```

Optional, for pronunciation scoring — the OpenPronounce model host. The first
`up` builds from the pinned upstream tag (one-time: ~5 GB, needs network):

```bash
docker compose up -d openpronounce
make pronounce-warm        # load the wav2vec2 weights now (~20 s) instead of on the first request
```

### 3. App env files

```bash
cp apps/api/.env.example apps/api/.env
cp apps/voice/.env.example apps/voice/.env
cp apps/web/.env.example apps/web/.env
```

Fill in the keys: Clerk (api + web — the API refuses to boot without
`CLERK_SECRET_KEY`), and LLM/STT/TTS (api scoring + voice engine). Keep
`INTERNAL_SECRET` identical in api and voice; the `.env.example` files ship the
same dev value. The web app talks to `http://localhost:8000` by default;
override with `VITE_API_URL` if the API runs elsewhere.

### 4. Database + migrations

```bash
cd apps/api
make migration-up          # goose up, DATABASE_URL from .env
```

### 5. Run the apps (one terminal each)

| Terminal | Command | Serves |
|----------|---------|--------|
| API | `cd apps/api && go run ./cmd/server` | `http://localhost:8000`; Swagger UI at `/swagger` (regenerate: `make swagger`) |
| Voice engine | `cd apps/voice && uv run python src/bot.py` | `:7860`; the browser connects over WebRTC **directly** here — Go never proxies audio |
| Web | `pnpm --filter web dev` | `http://localhost:3000` |

### Local ports

| Port | Service |
|------|---------|
| 3000 | Web (Vite dev) |
| 8000 | API (Go) |
| 7860 | Voice engine (Pipecat) |
| 8001 | Pronunciation (OpenPronounce container) |
| 9000 / 9010 | MinIO API / console |
| 5777 | Postgres (host port; set by `DB_PORT` in `apps/api/.env`) |
| 6379 | Redis |

### Seed lesson content (optional)

```bash
cd apps/api && make seed-audio   # TTS-generate lesson audio into MinIO (needs a migrated DB + minio up)
```

### Gates (run before every change)

```bash
# API
cd apps/api && go test ./... -count=1 && go vet ./... && test -z "$(gofmt -l internal/)"

# Voice
cd apps/voice && uv run pytest -q && uv run ruff check src tests

# Web
cd apps/web && pnpm check && pnpm exec tsc --noEmit && pnpm check:conventions && pnpm build

# End-to-end (Playwright — needs its own env, see tests/e2e/README.md)
pnpm --filter e2e test:e2e
```

---

## API Development Commands

```bash
cd apps/api
make migration-create name=...  # new goose migration (pre-deploy: edit in place + down/up)
make migration-up / migration-down
make mocks                       # regenerate mockery mocks
make contracts                     # tygo generate → packages/contracts
make swagger                       # swag init -g cmd/server/main.go
```

<!-----

## Roadmap

### MVP scope (per spec)

* [ ] Onboarding/profile form (level, goals, interests, study time)
* [ ] Lessons — reading + writing on the dictation engine
* [ ] Video side-quests — dictation + shadowing
* [ ] AI voice — free talk + fixed persona roleplays (prompt-only)
* [ ] Per-turn async feedback + pronunciation scoring
* [ ] Vocabulary — save from lessons/videos/conversations + basic SRS
* [ ] Progress — completion-ordered content list

### Stretch

* [ ] Auto takeaway summary at session end
* [ ] Objective tracking inside roleplays
* [ ] Streaks, skill trends, personalized recommendations
* [ ] Admin/CMS content UI (seed via script/DB for now)
```-->
