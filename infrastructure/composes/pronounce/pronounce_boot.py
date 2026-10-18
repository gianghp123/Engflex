"""Boot wrapper for the OpenPronounce container: warm models before serving.

Upstream (`server.py`) loads its wav2vec2 checkpoints lazily on the first
request, so the first learner attempt after any (re)start paid the full load
plus the first-inference warmup — 34 s observed in production
(2026-10-18T10:25:13Z: 21.6 s checkpoint loading, 12.7 s first scoring). This
wrapper drives the same model loaders and inferential paths the endpoints use
BEFORE ``uvicorn.run`` binds :8000, so ``/health`` only passes once scoring is
actually warm and Coolify's healthcheck keeps the service out of rotation
until then. Restarts therefore take ~15-25 s longer before serving, and no
user request ever pays the cold cost.

Upstream code stays unmodified (frozen v0.3.0 artifact); this file only
launches it. The TTS reference path is deliberately not warmed: it is
network-backed (gTTS) and cached per text, not part of the cold-start cost.

Verified against the artifact: with ``HF_HUB_OFFLINE=1`` and no network both
checkpoints load from the baked ``/models`` cache.
"""

import logging
import os
import time

import numpy as np
import uvicorn

PORT = int(os.environ.get("PORT", "8000"))
SAMPLE_RATE = 16000  # openpronounce: SAMPLING_RATE / audio.TARGET_SR
WARM_PHRASE = "warm up"

log = logging.getLogger("openpronounce.boot")


def _phase(name: str, fn) -> None:
    started = time.time()
    fn()
    log.info("prewarm: %s in %.1fs", name, time.time() - started)


def prewarm() -> None:
    # Imported here, not at module top: torch/transformers dominate the wait.
    from openpronounce import phones, speech

    silence = np.zeros(SAMPLE_RATE, dtype=np.float32)

    # One pass over each inferential path a request touches, in the order the
    # API calls them (see apps/api/internal/voice/pronounce_engine.go). The
    # prosody phases matter more than they look: librosa's pyin compiles its
    # numba kernel on first call (measured 9.6 s, then 0.03 s per call), and
    # compare_audio_with_text always runs it at the end.
    _phase("asr checkpoint + inference", lambda: speech.transcribe(silence, "en"))
    _phase("embedding checkpoint + inference", lambda: speech.extract_embeddings(silence, SAMPLE_RATE))
    _phase("phone checkpoint + inference", lambda: phones.recognize_phones(silence, SAMPLE_RATE, lang="en"))
    _phase("phonemizer path", lambda: speech.get_phonemes_with_word_mapping(WARM_PHRASE, "en"))
    _phase("prosody: pitch contour (pyin/numba jit)",
           lambda: speech.interpolate_f0(speech.extract_f0(silence, SAMPLE_RATE)))
    _phase("prosody: energy", lambda: speech.extract_energy(silence))


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    started = time.time()
    log.info("prewarming models before binding :%d", PORT)
    try:
        prewarm()
    except Exception:
        # Never serve cold on a broken cache/failed load: exit non-zero so the
        # restart policy retries and the deployment is visibly unhealthy.
        log.exception("prewarm failed - refusing to serve cold")
        raise SystemExit(1)
    log.info("prewarm complete in %.1fs - starting uvicorn", time.time() - started)

    from server import app  # upstream FastAPI app, unmodified

    uvicorn.run(app, host="0.0.0.0", port=PORT)


if __name__ == "__main__":
    main()
