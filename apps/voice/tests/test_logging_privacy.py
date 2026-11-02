import json

from loguru import logger

from observability import configure_logging


def test_default_logging_preserves_callback_identity_without_provider_payloads():
    messages = []
    configure_logging(messages.append)
    try:
        logger.debug("private debug transcript")
        logger.bind(conversation_id="c1", revision=3).error("callback unacknowledged")
        # Exercise the same record filter used by third-party warning logs.
        with logger.contextualize(text="private prepared TTS text"):
            logger.patch(lambda record: record.update(name="pipecat.services.tts_service")).warning("private prepared TTS text")
        records = [json.loads(message)["record"] for message in messages]
        assert len(records) == 2
        assert records[0]["extra"] == {"conversation_id": "c1", "revision": 3}
        assert records[1]["message"] == "voice dependency diagnostic"
        assert records[1]["extra"] == {}
        assert "private" not in "".join(messages)
    finally:
        configure_logging()
