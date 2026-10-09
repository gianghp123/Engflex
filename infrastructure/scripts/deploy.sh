#!/usr/bin/env bash
# Deploy: terraform apply → goose migrations (with postgres readiness wait)
# → trigger all four Coolify app deploys (api, web, voice, pronounce) →
# external health polls (api/web) + internal status probes (voice/pronounce).
# Run via `make deploy` from the repo root (or directly from anywhere).
# Exit non-zero on the first failure, with the failing stage named.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

TF_DIR="infrastructure/environments/selfhost"
MIGRATIONS_DIR="apps/api/internal/database/migrations"
GOOSE="${GOOSE:-$HOME/go/bin/goose}"

command -v terraform >/dev/null || { echo "ERROR: terraform not on PATH" >&2; exit 1; }
[ -x "$GOOSE" ] || { echo "ERROR: goose not found at $GOOSE (set \$GOOSE, or install to ~/go/bin/goose)" >&2; exit 1; }
command -v curl >/dev/null || { echo "ERROR: curl not on PATH" >&2; exit 1; }

tf() { terraform -chdir="$TF_DIR" "$@"; }

echo "== terraform =="
tf init -input=false
tf apply -input=false -auto-approve

echo "== capture outputs =="
TOKEN="$(tf output -raw coolify_token)"
COOLIFY="$(tf output -raw coolify_endpoint)"
DATABASE_URL="$(tf output -raw postgres_public_url)"
API_UUID="$(tf output -raw api_app_uuid)"
WEB_UUID="$(tf output -raw web_app_uuid)"
API_HEALTH_URL="$(tf output -raw api_health_url)"
WEB_HOST="$(tf output -raw web_host)"

echo "== migrations =="
echo "waiting for postgres to accept connections…"
db_deadline=$((SECONDS + 180))
until "$GOOSE" -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" status >/dev/null 2>&1; do
  if [ "$SECONDS" -ge "$db_deadline" ]; then
    echo "ERROR: TIMEOUT after 180s waiting for postgres to accept connections" >&2
    exit 1
  fi
  echo "waiting for postgres…"
  sleep 5
done
"$GOOSE" -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" up

echo "== trigger Coolify deploys =="
VOICE_UUID="$(tf output -raw voice_app_uuid)"
PRONOUNCE_UUID="$(tf output -raw pronounce_app_uuid)"
for UUID in "$API_UUID" "$WEB_UUID" "$VOICE_UUID" "$PRONOUNCE_UUID"; do
  if ! curl -fsS -X POST "$COOLIFY/api/v1/deploy?uuid=$UUID" \
      -H "Authorization: Bearer $TOKEN" >/dev/null; then
    echo "ERROR: deploy trigger failed for app $UUID" >&2
    exit 1
  fi
  echo "triggered $UUID"
done

echo "== health =="
poll() { # $1=url $2=label $3=deadline_sec (default 900)
  local url="$1" label="$2" deadline_sec="${3:-900}"
  local deadline=$((SECONDS + deadline_sec))
  until curl -fsS --max-time 5 -o /dev/null "$url"; do
    if [ "$SECONDS" -ge "$deadline" ]; then
      echo "ERROR: TIMEOUT after ${deadline_sec}s waiting for $label ($url)" >&2
      exit 1
    fi
    echo "waiting for $label…"
    sleep 5
  done
  echo "$label healthy"
}
poll_internal() { # $1=url $2=label $3=deadline_sec (default 900)
  local url="$1" label="$2" deadline_sec="${3:-900}"
  local deadline=$((SECONDS + deadline_sec))
  until docker run --rm --network coolify curlimages/curl -fsS --max-time 10 -o /dev/null "$url"; do
    if [ "$SECONDS" -ge "$deadline" ]; then
      echo "ERROR: TIMEOUT after ${deadline_sec}s waiting for $label ($url)" >&2
      exit 1
    fi
    echo "waiting for $label…"
    sleep 10
  done
  echo "$label healthy"
}
poll "$API_HEALTH_URL" "api" 1200
poll "https://$WEB_HOST/" "web" 1200

echo "== voice internal status =="
poll_internal "http://engflex-voice:7860/status" "voice" 900

echo "== pronounce internal health =="
poll_internal "http://engflex-pronounce:8000/health" "pronounce" 3600

echo "deploy complete: https://$WEB_HOST/"
