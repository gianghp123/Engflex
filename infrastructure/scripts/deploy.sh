#!/usr/bin/env bash
# Deploy: terraform plan → apply (iff changed; rollout is TF-owned — apps
# via coolify_deployment resources, services via coolify_resource_action
# with env-hash triggers; empty diffs no-op) → goose migrations (with
# postgres readiness wait) → ensure frozen pronounce artifact → external
# health polls (api/web) + internal status probes (voice/pronounce).
# Run via `make deploy` from the repo root (or directly from anywhere).
# FORCE=1 make deploy → force-replace all four app deployments.
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
# App rollout lives in TF (coolify_deployment resources in deployments.tf):
# env/code/structural changes redeploy exactly the affected apps, empty diffs
# no-op. Only services (provider has no service-deployment resource) still
# trigger via curl below. Origin reachability is required at plan time (the
# external data source reads origin HEAD) — same network the providers need.

# FORCE=1 make deploy → force-replace all four app deployments.
FORCE="${FORCE:-0}"
REPLACE_ARGS=()
if [ "$FORCE" = "1" ]; then
  for r in api web voice pronounce; do
    REPLACE_ARGS+=("-replace=coolify_deployment.$r")
  done
fi

tf() { terraform -chdir="$TF_DIR" "$@"; }

echo "== ensure pronounce artifact (frozen image) =="
# Runs BEFORE terraform: the pronounce TF deployment (fired during apply)
# pulls this tag, so the artifact must exist before apply starts.
bash "$REPO_ROOT/infrastructure/scripts/ensure-pronounce-image.sh"

echo "== terraform =="
tf init -input=false
PLAN_FILE="$(mktemp "${TMPDIR:-/tmp}/engflex-tfplan.XXXXXX")"
DB_WAIT_LOG="$(mktemp "${TMPDIR:-/tmp}/engflex-dbwait.XXXXXX")"
trap 'rm -f "$PLAN_FILE" "$DB_WAIT_LOG"' EXIT
# detailed-exitcode: 0 = empty, 2 = changes. (set -e would abort on 2.)
set +e
tf plan -input=false -out="$PLAN_FILE" -detailed-exitcode
PLAN_EXIT=$?
set -e
case "$PLAN_EXIT" in
  0) echo "terraform: plan empty, nothing to apply";;
  2) echo "terraform: changes pending";;
  *) echo "ERROR: terraform plan failed (exit $PLAN_EXIT)" >&2; exit 1;;
esac

if [ "$PLAN_EXIT" = "2" ]; then
  echo "== terraform apply =="
  tf apply -input=false -auto-approve "${REPLACE_ARGS[@]}" "$PLAN_FILE"
fi

echo "== capture outputs =="
DATABASE_URL="$(tf output -raw postgres_public_url)"
API_HEALTH_URL="$(tf output -raw api_health_url)"
WEB_HOST="$(tf output -raw web_host)"
TUNNEL_HOST="$(tf output -raw tunnel_hostname)"

echo "== migrations =="
echo "waiting for postgres to accept connections…"
# 600s: on a full recreate every container starts async behind API-accept,
# so the gate must outlast scheduling lag. Outputs are captured post-apply:
# UUIDs rotate on destroy/recreate, so pre-apply values point at dead
# resources.
db_deadline=$((SECONDS + 600))
# The probe's stderr is LOGGED, never suppressed: silent retries destroy the
# only evidence when the gate fails. Each retry prints the last error; a
# timeout dumps the tail.
until "$GOOSE" -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" status >"$DB_WAIT_LOG" 2>&1; do
  if [ "$SECONDS" -ge "$db_deadline" ]; then
    echo "ERROR: TIMEOUT after 600s waiting for postgres to accept connections" >&2
    tail -20 "$DB_WAIT_LOG" >&2
    exit 1
  fi
  echo "waiting for postgres… ($(tail -1 "$DB_WAIT_LOG"))"
  sleep 5
done
"$GOOSE" -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" up

echo "== health =="
poll() { # $1=url $2=label $3=deadline_sec (default 900)
  local url="$1" label="$2" deadline_sec="${3:-900}"
  local deadline=$((SECONDS + deadline_sec))
  # Plain-HTTP local routes (tunnel-only + localhost): no TLS locally, so no
  # --cacert dance. The tunnel edge still terminates real certs publicly.
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
poll "http://$WEB_HOST/" "web" 1200

echo "== voice internal status =="
poll_internal "http://engflex-voice:7860/status" "voice" 900

echo "== pronounce internal health =="
poll_internal "http://engflex-pronounce:8000/health" "pronounce" 3600

if [ -n "$TUNNEL_HOST" ]; then
  echo "deploy complete: https://$TUNNEL_HOST/ (local: http://$WEB_HOST/)"
else
  echo "deploy complete: http://$WEB_HOST/"
fi
