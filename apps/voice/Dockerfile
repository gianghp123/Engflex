FROM ghcr.io/astral-sh/uv:latest AS uv
# digest: sha256:3af4716e991d6956a41e573eab705d0ee08500cd829ed30293eb8472f372c65a

FROM python:3.12-slim
COPY --from=uv /uv /uvx /bin/
# curl is required for Coolify's container healthcheck (same rule as the api
# and pronounce images); python-slim ships neither curl nor wget.
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*

ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 UV_COMPILE_BYTECODE=1
WORKDIR /app

COPY apps/voice/pyproject.toml apps/voice/uv.lock ./
# BuildKit cache mount: uv downloads/rebuilt wheels survive even when this
# layer misses (post-cleanup rebuilds), so a cache-miss reinstall costs
# minutes, not the ~11 min cold download observed 2026-10-11.
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --no-install-project

COPY apps/voice/src ./src
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

EXPOSE 7860
CMD ["uv", "run", "src/bot.py", "-t", "webrtc", "--host", "0.0.0.0", "--port", "7860"]
