#!/usr/bin/env bash
# Ensure the frozen OpenPronounce artifact exists in the local registry and
# matches the current recipe. Builds and pushes it on first use or whenever
# infrastructure/composes/pronounce/Dockerfile.artifact changes.
#   --force       build+push unconditionally
#   --check-only  report what would happen, never build
# Env overrides (used by tests): RECIPE_FILE, BUILD_CONTEXT, ARTIFACT_TAG.
# NOTE: keep the ARTIFACT_TAG default in sync with the FROM tag in
# infrastructure/composes/pronounce/Dockerfile (stub). The recipe hash covers
# Dockerfile.artifact only — the recipe must stay self-contained (no COPY
# from the build context) for the hash to remain a complete change detector.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RECIPE_FILE="${RECIPE_FILE:-$REPO_ROOT/infrastructure/composes/pronounce/Dockerfile.artifact}"
BUILD_CONTEXT="${BUILD_CONTEXT:-$REPO_ROOT/infrastructure/composes/pronounce}"
ARTIFACT_TAG="${ARTIFACT_TAG:-127.0.0.1:5000/engflex-pronounce:v0.3.0}"

FORCE=0
CHECK_ONLY=0
for arg in "$@"; do
  case "$arg" in
    --force) FORCE=1 ;;
    --check-only) CHECK_ONLY=1 ;;
    *) echo "ERROR: unknown argument: $arg (expected --force/--check-only)" >&2; exit 1 ;;
  esac
done

[ -f "$RECIPE_FILE" ] || { echo "ERROR: recipe not found: $RECIPE_FILE" >&2; exit 1; }
[ -d "$BUILD_CONTEXT" ] || { echo "ERROR: build context not found: $BUILD_CONTEXT" >&2; exit 1; }

HASH=$(sha256sum "$RECIPE_FILE" | awk '{print $1}')

REASON=""
if [ "$FORCE" -eq 1 ]; then
  REASON="--force"
elif docker pull -q "$ARTIFACT_TAG" >/dev/null 2>&1; then
  # NOTE: `docker manifest inspect` is not used: it reports "no such
  # manifest" on this host even for tags the registry serves (verified
  # against both this artifact and the controller-seeded minio image).
  # The pull doubles as the presence check and ensures the image locally.
  CURRENT="$(docker image inspect --format '{{ index .Config.Labels "engflex.recipe-hash" }}' "$ARTIFACT_TAG" 2>/dev/null || true)"
  case "$CURRENT" in "<no value>"|"") CURRENT="" ;; esac
  if [ "$CURRENT" != "$HASH" ]; then
    REASON="recipe changed (have ${CURRENT:-none}, want $HASH)"
  fi
else
  REASON="missing from registry"
fi

if [ "$CHECK_ONLY" -eq 1 ]; then
  if [ -n "$REASON" ]; then
    echo "would build ($REASON)"
  else
    echo "up to date"
  fi
  exit 0
fi

if [ -z "$REASON" ]; then
  echo "pronounce artifact $ARTIFACT_TAG is up to date (recipe $HASH)"
  exit 0
fi

echo "building pronounce artifact $ARTIFACT_TAG ($REASON)…"
docker build --label "engflex.recipe-hash=$HASH" -f "$RECIPE_FILE" -t "$ARTIFACT_TAG" "$BUILD_CONTEXT"
echo "pushing $ARTIFACT_TAG…"
docker push "$ARTIFACT_TAG"
echo "confirming label…"
docker pull -q "$ARTIFACT_TAG" >/dev/null 2>&1
CONFIRMED="$(docker image inspect --format '{{ index .Config.Labels "engflex.recipe-hash" }}' "$ARTIFACT_TAG")"
[ "$CONFIRMED" = "$HASH" ] || { echo "ERROR: label mismatch after push (have ${CONFIRMED:-none}, want $HASH)" >&2; exit 1; }
echo "pronounce artifact $ARTIFACT_TAG ready (recipe $HASH)"
