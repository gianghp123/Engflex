# syntax=docker/dockerfile:1
# OpenPronounce is a pinned upstream (Halleck45/OpenPronounce @ v0.3.0) we do
# not modify. The finished image — apt deps, CPU torch, "[app]" + uvicorn,
# the baked wav2vec2 checkpoints, curl for Coolify's healthcheck — is frozen
# as an OCI artifact in the local registry. This build is therefore a
# localhost pull: seconds, no model download, no pip.
#
# The artifact is ensured automatically by `make deploy` and can be managed
# manually with `make pronounce-artifact` (see infrastructure/scripts/
# ensure-pronounce-image.sh; recipe: Dockerfile.artifact).
# NOTE: keep this tag in sync with ARTIFACT_TAG in ensure-pronounce-image.sh.
FROM 127.0.0.1:5000/engflex-pronounce:v0.3.0
