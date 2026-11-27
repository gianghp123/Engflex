#!/usr/bin/env bash
# Run api + voice + web together, one tmux window each.
# Usage: ./scripts/dev.sh   (re-run to re-attach)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SESSION="${SESSION:-engflex}"

if tmux has-session -t "$SESSION" 2>/dev/null; then
  echo "attaching to existing session '$SESSION'"
  exec tmux attach-session -t "$SESSION"
fi

# Infra the api needs (postgres/redis/minio). Runs detached, no window needed.
# Skip with: SKIP_INFRA=1 ./scripts/dev.sh
if [[ "${SKIP_INFRA:-0}" != "1" ]]; then
  (cd "$ROOT/apps/api" && docker compose up -d postgres redis minio minio-init)
fi

tmux new-session -d -s "$SESSION" -n api -c "$ROOT/apps/api" \
  "go run cmd/server/main.go; echo 'api exited (code $?)'; exec bash"
tmux new-window -t "$SESSION" -n voice -c "$ROOT/apps/voice" \
  "uv run python src/bot.py; echo 'voice exited (code $?)'; exec bash"
tmux new-window -t "$SESSION" -n web -c "$ROOT/apps/web" \
  "pnpm dev; echo 'web exited (code $?)'; exec bash"

tmux select-window -t "$SESSION:api"
exec tmux attach-session -t "$SESSION"
