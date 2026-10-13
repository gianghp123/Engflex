"""Pipecat service builders: STT, LLM, TTS, and LLM context.

STT and TTS return pipecat's service base classes: which provider serves a
role is configuration (`STT_API_KEY`, `VOICE_MODEL`), so callers should
not depend on a vendor type. The LLM builder stays concrete — see below.
"""

from pipecat.processors.aggregators.llm_context import LLMContext
from pipecat.services.deepgram.stt import DeepgramSTTService
from pipecat.services.fish.tts import FishAudioTTSService
from pipecat.services.openai.llm import OpenAILLMService
from pipecat.services.stt_service import STTService
from pipecat.services.tts_service import TTSService

from app.schemas import PersonaBody, ScenarioBody
from config import settings
from observability import get_metrics
from prompts import build_system_prompt


def build_stt() -> STTService:
    # 1.11.0 moved interim_results/punctuate/endpointing into Settings
    # (direct kwargs no longer exist; live_options is deprecated).
    return DeepgramSTTService(
        api_key=settings.stt_api_key,
        settings=DeepgramSTTService.Settings(
            interim_results=True,
            punctuate=False,
            endpointing=300,
        ),
        metrics=get_metrics(),
    )


def build_llm(
    *,
    conversation_id: str | None = None,
    persona: PersonaBody | None = None,
    scenario: ScenarioBody | None = None,
    level: str | None = None,
) -> OpenAILLMService:
    # The LLM return type stays concrete: `LLMService` is generic over its
    # adapter and that parameter is invariant, so the base class cannot
    # describe a provider without naming one. OpenAI-compatible is the
    # interface we drive (any gateway speaking it works via LLM_BASE_URL).
    extra = (
        {"reasoning_effort": settings.llm_reasoning_effort} if settings.llm_reasoning_effort else {}
    )
    headers = {"User-Agent": settings.llm_user_agent}
    if conversation_id:
        # Stable per-conversation session id for proxied endpoints that use
        # it for routing / prompt caching (e.g. opencode.ai x-opencode-session).
        headers["x-opencode-session"] = conversation_id
    return OpenAILLMService(
        api_key=settings.llm_api_key,
        base_url=settings.llm_base_url,
        settings=OpenAILLMService.Settings(
            model=settings.llm_model,
            extra=extra,
            max_tokens=200,
            system_instruction=build_system_prompt(
                level or "B1", persona=persona, scenario=scenario
            ),
        ),
        default_headers=headers,
        metrics=get_metrics(),
    )


def build_tts(persona: PersonaBody | None = None) -> TTSService:
    # Hosted TTS: VOICE_MODEL is the synthesis model, the voice identity is
    # selected by persona gender (MALE_VOICE_ID / FEMALE_VOICE_ID). Sessions
    # without a usable gender (freetalk has no persona) default to the female
    # voice — never the provider default: reference_id=None lets Fish pick a
    # voice per connection, which flips gender audibly on reconnects. Nothing here names a vendor, so pointing at a different
    # engine is a config change.
    voice: str | None = settings.female_voice_id or settings.voice_id or None
    gender = (persona.gender if persona else "").strip().lower()
    if gender == "male" and settings.male_voice_id:
        voice = settings.male_voice_id
    elif gender == "female" and settings.female_voice_id:
        voice = settings.female_voice_id
    return FishAudioTTSService(
        api_key=settings.tts_api_key,
        settings=FishAudioTTSService.Settings(
            model=settings.voice_model,
            voice=voice,
        ),
        metrics=get_metrics(),
    )


def build_context() -> LLMContext:
    # One seeded user opener, zero system messages. Two upstream constraints:
    # requests with no user message are rejected with 400, and stuffing the
    # prompt into the context is deprecated since pipecat 1.9 — the tutor
    # prompt travels via the LLM service's system_instruction instead (single
    # system message).
    # The opener is a bare greeting on purpose: anything more specific makes
    # the model *reply* to it ("No problem—…") instead of delivering its own
    # opening line.
    return LLMContext(messages=[{"role": "user", "content": "Hello!"}])
