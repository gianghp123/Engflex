#!/usr/bin/env bash
# Seed lesson audio into the deployed Postgres + MinIO. Run AFTER
# `make deploy`: `make deploy-seed`.
# Every Terraform output is captured BEFORE the `cd apps/api` — terraform's
# relative -chdir path only resolves from the repo root.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

TF_DIR="infrastructure/environments/selfhost"
BUCKET="lesson-audio"

command -v terraform >/dev/null || { echo "ERROR: terraform not on PATH" >&2; exit 1; }
command -v docker >/dev/null || { echo "ERROR: docker not on PATH" >&2; exit 1; }
command -v uv >/dev/null || { echo "ERROR: uv not on PATH" >&2; exit 1; }

tf() { terraform -chdir="$TF_DIR" "$@"; }

echo "== capture outputs =="
# Plain HTTP by intent (tunnel-only + localhost): minio speaks http locally.
MINIO_ENDPOINT="http://$(tf output -raw minio_host)"
MINIO_ROOT_USER="$(tf output -raw minio_root_user)"
MINIO_ROOT_PASSWORD="$(tf output -raw minio_root_password)"
DATABASE_URL="$(tf output -raw postgres_public_url)"

echo "== bucket + anonymous-download policy =="
# seed_audio.py FAILS if the bucket does not exist, so the bucket is created
# here; the anonymous-download policy is what makes media URLs servable
# without credentials (Task 9's unauthenticated probe pins it).
# NOTE: the minio/mc image's entrypoint IS `mc`, so the multi-step flow runs
# via --entrypoint sh. Plain http locally — no --insecure needed.
docker run --rm --network host \
  -e MC_CONFIG_DIR=/tmp/.mc \
  -e MINIO_ENDPOINT="$MINIO_ENDPOINT" \
  -e MINIO_ROOT_USER="$MINIO_ROOT_USER" \
  -e MINIO_ROOT_PASSWORD="$MINIO_ROOT_PASSWORD" \
  -e BUCKET="$BUCKET" \
  --entrypoint sh minio/mc -c '
    mc alias set deploy "$MINIO_ENDPOINT" "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" &&
    mc mb --ignore-existing "deploy/$BUCKET" &&
    mc anonymous set download "deploy/$BUCKET"
  '

echo "== seed audio =="
cd "$REPO_ROOT/apps/api"
DATABASE_URL="$DATABASE_URL" \
MINIO_ENDPOINT="$MINIO_ENDPOINT" \
MINIO_ROOT_USER="$MINIO_ROOT_USER" \
MINIO_ROOT_PASSWORD="$MINIO_ROOT_PASSWORD" \
MINIO_BUCKET="$BUCKET" \
uv run --with edge-tts --with minio --with "psycopg[binary]" --with python-dotenv scripts/seed_audio.py

echo "seed complete: bucket $BUCKET"
