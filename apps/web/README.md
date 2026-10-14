# EngFlex Web — TanStack Start Frontend

Vite + React 19 + TanStack Start/Router/Query/Store, Clerk auth, shadcn/ui, and
Paraglide i18n (`vi` default + `en`). Voice screens talk WebRTC directly to the
voice engine — Go is never in the audio path.

## Layout

```text
src/
  routes/                    # Directory convention; thin shells
  features/<domain>/         # attempts, dashboard, dictation, lessons, onboarding,
                             # profiles, reading, vocabulary, voice, writing
    components/              # Presentational components
    fixtures.ts / store.ts   # Mock data + cross-screen store
  components/common|layout|ui# Shared + layout + shadcn primitives
  app/                       # APP_ROUTES, API_ROUTES, breadcrumbs
  lib/                       # api.ts, auth-guard.ts, paraglide-middleware.ts
  paraglide/                 # Generated — never hand-edit
messages/{en,vi}/<domain>.json  # inlang sources
scripts/check-conventions.mjs   # Message ids, dynamic keys, hardcoded URLs
```

## Quickstart

```bash
pnpm install            # from repo root
pnpm --filter web dev   # Vite dev on :3000
```

## Gates

```bash
pnpm --filter web compile:messages   # before typecheck when copy changes
pnpm --filter web check              # biome
pnpm --filter web typecheck
pnpm --filter web check:conventions  # orphaned ids, en/vi divergence, hardcoded URLs
pnpm --filter web build
```
