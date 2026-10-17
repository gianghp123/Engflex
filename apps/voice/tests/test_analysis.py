from typing import Any

import pytest
from pydantic import ValidationError

from transcript.analyze import (
    AnalyzeRequest,
    ContextTurn,
    TurnFeedback,
    _messages,
    build_feedback_prompt,
    extract_json_object,
    parse_feedback,
)


def _request() -> AnalyzeRequest:
    return AnalyzeRequest(
        conversationId="c1",
        turnId="t1",
        text="I go to the office yesterday.",
        context=[ContextTurn(role="ai", text="What did you do yesterday?")],
    )


def _payload(**overrides: object) -> str:
    import json

    data: dict[str, object] = {
        "corrected": "I went to the office yesterday.",
        "spans": [],
        "relevance": {"status": "relevant", "reason": None},
        "alternatives": {"language": None, "contextual": None},
        "tip": "Use the past tense for finished time.",
    }
    data.update(overrides)
    return json.dumps(data)


def test_messages_include_a_user_message():
    # The upstream provider rejects requests with no user message (the same
    # 400 that killed the greeting turn before the "Hello!" opener). The
    # system prompt already carries the utterance; the user message repeats
    # it as the review target so the request is never system-only.
    messages = _messages(_request())
    assert messages[0].get("role") == "system"
    user = [m for m in messages if m.get("role") == "user"]
    assert len(user) == 1
    assert "yesterday" in str(user[0].get("content", ""))


def test_prompt_demands_the_five_keys():
    prompt = build_feedback_prompt(_request())
    for key in ("corrected", "spans", "relevance", "alternatives", "tip"):
        assert key in prompt
    assert "yesterday" in prompt


def test_prompt_never_asks_for_phonemes():
    # Spans are word-level language spans, never phonemes: nothing on this
    # route produces acoustic measurements, so the contract must not ask for
    # them, or the model invents scores.
    assert "phoneme" not in build_feedback_prompt(_request()).lower()


def test_prompt_never_asks_for_a_speech_block():
    assert "speech" not in build_feedback_prompt(_request()).lower()


def test_prompt_carries_no_level_or_objective():
    # The conversation context already expresses the topic; the ask is simply
    # "is this correct English", which does not vary by level.
    req = AnalyzeRequest(conversationId="c1", turnId="t1", text="hello")
    prompt = build_feedback_prompt(req)
    assert "cefr" not in prompt.lower()
    assert "objective" not in prompt.lower()


def test_prompt_default_matches_the_english_spelling():
    # "" and "English" must render one brief, or the string every existing
    # reply was written against drifts the moment a learner picks English.
    english = _request().model_copy(update={"language": "English"})
    assert build_feedback_prompt(_request()) == build_feedback_prompt(english)


def test_prompt_localizes_only_the_prose():
    # The corrected utterance and the span text are the learner's own
    # target-language content: only the coaching prose localizes.
    req = _request().model_copy(update={"language": "Vietnamese"})
    prompt = build_feedback_prompt(req)
    assert "in Vietnamese only" in prompt
    assert "Keep corrected, every span text" in prompt
    assert "in Vietnamese only" not in build_feedback_prompt(_request())


def test_extract_returns_a_bare_object():
    raw = '{"corrected": "a", "spans": [], "relevance": {"status": "relevant", "reason": null}, "alternatives": {"language": null, "contextual": null}, "tip": "t"}'
    assert extract_json_object(raw) == raw


def test_extract_ignores_prose_around_the_object():
    raw = 'Sure!\n{"corrected": "a"}\nHope that helps.'
    assert extract_json_object(raw) == '{"corrected": "a"}'


def test_extract_raises_when_there_is_no_json_block():
    with pytest.raises(ValueError):
        extract_json_object("**Corrected sentence:** I went to the office.")


def test_parse_validates_a_clean_object():
    fb = parse_feedback(_payload())
    assert fb.corrected == "I went to the office yesterday."
    assert fb.spans == []
    assert fb.relevance.status == "relevant"
    assert fb.alternatives.language is None


def test_parse_accepts_a_span_with_occurrence():
    fb = parse_feedback(
        _payload(
            spans=[
                {
                    "text": "didn't went",
                    "occurrence": 2,
                    "status": "incorrect",
                    "correction": "didn't go",
                    "reason": "Use the base form after did or didn't.",
                }
            ]
        )
    )
    assert fb.spans[0].occurrence == 2
    assert fb.spans[0].correction == "didn't go"


def test_parse_survives_a_bolded_reply():
    # The model ignores response_format on this route and answers in
    # markdown with a JSON object inside.
    raw = "**Corrected sentence:** I **went** to the office yesterday.\n\n" + _payload()
    assert parse_feedback(raw).tip == "Use the past tense for finished time."


def test_parse_raises_on_an_unexpected_shape():
    with pytest.raises(ValidationError):
        parse_feedback('{"corrected": "a", "tip": "t"}')  # spans/relevance/alternatives missing


def test_parse_raises_when_a_span_has_an_unknown_status():
    with pytest.raises(ValidationError):
        parse_feedback(
            _payload(
                spans=[
                    {
                        "text": "go",
                        "occurrence": 1,
                        "status": "error",
                        "correction": "went",
                        "reason": "Past tense.",
                    }
                ]
            )
        )


def test_parse_raises_when_occurrence_is_not_positive():
    with pytest.raises(ValidationError):
        parse_feedback(
            _payload(
                spans=[
                    {
                        "text": "go",
                        "occurrence": 0,
                        "status": "incorrect",
                        "correction": "went",
                        "reason": "Past tense.",
                    }
                ]
            )
        )


def test_models_derive_a_valid_strict_schema():
    # Regression guard for the structured-output path. The SDK derives the
    # schema from the pydantic models, and strict mode has two traps that are
    # easy to reintroduce by hand: every property must appear in `required`,
    # and `additionalProperties: false` must be set on nested objects too.
    # TurnFeedback deliberately declares no defaults -- a default would let
    # the model and its own derived schema disagree about what is required.
    # Verified against openai 3.19.2.
    from openai.lib._parsing._completions import type_to_response_format_param

    # The helper returns a union of response-format TypedDicts; the JSON
    # schema variant is the one this test inspects.
    response_format: Any = type_to_response_format_param(TurnFeedback)
    param: dict[str, Any] = response_format["json_schema"]
    schema: dict[str, Any] = param["schema"]

    assert param["strict"] is True
    assert set(schema["required"]) == {
        "corrected",
        "spans",
        "relevance",
        "alternatives",
        "tip",
    }
    assert schema["additionalProperties"] is False
    defs: dict[str, Any] = schema["$defs"]
    for name, definition in defs.items():
        assert definition["additionalProperties"] is False, f"{name} must forbid extras"
    assert "$schema" not in schema
