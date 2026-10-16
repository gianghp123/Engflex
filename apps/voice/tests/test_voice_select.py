from app.schemas import PersonaBody
from pipeline.services import build_tts


def _voice_of(persona):
    return build_tts(persona)._settings.voice


def test_male_persona_selects_male_voice(monkeypatch):
    from config import settings

    monkeypatch.setattr(settings, "male_voice_id", "male-1")
    monkeypatch.setattr(settings, "female_voice_id", "female-1")
    monkeypatch.setattr(settings, "voice_id", "default-1")
    assert _voice_of(PersonaBody(name="Marcus", gender="male")) == "male-1"


def test_female_persona_selects_female_voice(monkeypatch):
    from config import settings

    monkeypatch.setattr(settings, "male_voice_id", "male-1")
    monkeypatch.setattr(settings, "female_voice_id", "female-1")
    monkeypatch.setattr(settings, "voice_id", "default-1")
    assert _voice_of(PersonaBody(name="Amelia", gender="female")) == "female-1"


def test_unknown_gender_defaults_to_female_voice(monkeypatch):
    from config import settings

    monkeypatch.setattr(settings, "male_voice_id", "male-1")
    monkeypatch.setattr(settings, "female_voice_id", "female-1")
    monkeypatch.setattr(settings, "voice_id", "default-1")
    assert _voice_of(PersonaBody(name="X", gender="")) == "female-1"
    assert _voice_of(PersonaBody(name="Y", gender="nonbinary")) == "female-1"
    assert _voice_of(None) == "female-1"


def test_empty_female_id_falls_back_to_voice_id(monkeypatch):
    from config import settings

    monkeypatch.setattr(settings, "male_voice_id", "male-1")
    monkeypatch.setattr(settings, "female_voice_id", "")
    monkeypatch.setattr(settings, "voice_id", "default-1")
    assert _voice_of(None) == "default-1"


def test_empty_gender_ids_fall_back_to_provider_default(monkeypatch):
    from config import settings

    monkeypatch.setattr(settings, "male_voice_id", "")
    monkeypatch.setattr(settings, "female_voice_id", "")
    monkeypatch.setattr(settings, "voice_id", "")
    assert _voice_of(PersonaBody(name="Marcus", gender="male")) is None
