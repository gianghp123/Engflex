FROM ghcr.io/astral-sh/uv:latest AS uv
# digest: sha256:3af4716e991d6956a41e573eab705d0ee08500cd829ed30293eb8472f372c65a

FROM python:3.12-slim
COPY --from=uv /uv /uvx /bin/

ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 UV_COMPILE_BYTECODE=1
WORKDIR /app

COPY apps/voice/pyproject.toml apps/voice/uv.lock ./
RUN uv sync --frozen --no-dev --no-install-project

COPY apps/voice/src ./src
RUN uv sync --frozen --no-dev

EXPOSE 7860
CMD ["uv", "run", "src/bot.py", "-t", "webrtc", "--host", "0.0.0.0", "--port", "7860"]
