FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    HF_HOME=/models \
    PORT=8000

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg espeak-ng libsndfile1 git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
RUN git clone --depth 1 --branch v0.3.0 https://github.com/Halleck45/OpenPronounce.git .

# CPU wheels of torch are much smaller than the default CUDA ones.
RUN pip install torch --index-url https://download.pytorch.org/whl/cpu
# Upstream v0.3.0's Dockerfile `pip install ".[app]"` cannot resolve
# against current PyPI — pip backtracks through every uvicorn release
# and fails (uvicorn>=0.23 unpinned; verified 2026-10-08). Pin
# uvicorn[standard] below 0.31 to keep the resolution tractable.
RUN pip install ".[app]" "uvicorn[standard]<0.31"

RUN python -c "from openpronounce.speech import _load_models; _load_models()" \
    && python -c "from openpronounce.phones import _load_model; _load_model()"

EXPOSE 8000
CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000"]
