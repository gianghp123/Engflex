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

# Boot wrapper (ours; upstream stays unmodified): warms the checkpoints and
# the inference/phonemizer paths BEFORE uvicorn binds :8000, so /health only
# passes when the scorer is warm and no learner pays the ~17-34 s cold start.
# Rationale and measured baseline live in pronounce_boot.py.
COPY infrastructure/composes/pronounce/pronounce_boot.py /app/pronounce_boot.py
CMD ["python3", "/app/pronounce_boot.py"]
