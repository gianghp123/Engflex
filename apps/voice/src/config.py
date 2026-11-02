from pydantic import AliasChoices, Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Engine configuration, read from the environment.

    Naming rule: every setting names the ROLE (stt, llm, voice, models),
    never the vendor that happens to serve it — swapping a provider must not
    mean touching code or config names. Renamed settings still accept their
    previous env name, so an existing `.env` keeps working while it is
    migrated.
    """

    model_config = SettingsConfigDict(env_file=".env", extra="ignore", populate_by_name=True)

    # Speech to text
    stt_api_key: str = Field(validation_alias=AliasChoices("STT_API_KEY", "DEEPGRAM_API_KEY"))

    # Text to speech
    tts_api_key: str
    # Synthesis rate of the bot's audio. It comes from the TTS engine, not
    # from the transport: the engine renders at this rate and the WebRTC
    # output resamples as needed. Must be a rate the configured engine
    # accepts (Fish PCM: 8000/16000/24000/32000/44100) — a mismatch is
    # silent, the service just never emits audio.
    audio_out_sample_rate: int = 24000
    # The synthesis MODEL (e.g. s2.1-pro-free). No legacy alias: the old
    # PIPER_VOICE held a local voice NAME, which is a different concept from
    # a model id and would silently land in the wrong slot.
    voice_model: str = "s2.1-pro-free"
    # Optional voice identity (Fish `reference_id`: a cloned/stored voice).
    # Empty = the provider's default voice.
    voice_id: str = ""
    # Per-gender voice identities, selected by persona gender.
    # Empty = fall back to `voice_id`, then the provider default.
    male_voice_id: str = ""
    female_voice_id: str = ""

    # Language model (any OpenAI-compatible endpoint)
    llm_api_key: str = Field(validation_alias=AliasChoices("LLM_API_KEY", "OPENAI_API_KEY"))
    llm_model: str = Field(
        default="gpt-4o-mini", validation_alias=AliasChoices("LLM_MODEL", "LLM_NAME")
    )
    llm_base_url: str = Field(
        default="https://api.openai.com/v1",
        validation_alias=AliasChoices("LLM_BASE_URL", "OPENAI_BASE_URL"),
    )
    # Extra top-level chat-completions param (e.g. reasoning_effort=none for
    # reasoning models). Empty = not sent. Pipecat merges Settings.extra
    # into the request params verbatim.
    llm_reasoning_effort: str = ""
    # Client identity for proxied OpenAI-compatible endpoints that require
    # it (e.g. opencode.ai/zen: custom user agent + per-session header).
    llm_user_agent: str = Field(
        default="engflex-voice/0.1",
        validation_alias=AliasChoices("LLM_USER_AGENT", "OPENCODE_USER_AGENT"),
    )

    # Go control plane
    api_url: str = "http://localhost:8000"
    internal_secret: str

    # Session
    max_duration_sec: int = 300
    idle_timeout_sec: int = 300
    # Per-attempt callback timeout; total retry budget is capped at 30 seconds.
    turns_post_timeout_sec: int = 30

    # Analysis (a separate SDK client: the session's LLM service is streaming).
    analyze_timeout_sec: int = 60
    analyze_session_prefix: str = "engflex-analyze"
    # Turn off to force the extraction path without a code change.
    analysis_structured_output: bool = True

    # Observability (optional)
    sentry_dsn: str = ""


# Values come from the environment via pydantic-settings; `model_validate`
# is the type-checked way to express "construct from the environment" (a bare
# `Settings()` reads as missing the three required arguments).
settings = Settings.model_validate({})
