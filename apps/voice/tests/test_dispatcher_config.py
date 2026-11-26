"""Dispatcher settings behavior beyond env passthrough: the derived
MAX_PENDING default and the startup floor on MAX_SESSIONS (conversions and
fail-closed branches, per the repo's test rules)."""

import pytest
from pydantic import ValidationError

from config import Settings


def test_max_pending_derives_from_max_sessions():
    assert Settings.model_validate({"max_sessions": 7}).max_pending == 1
    assert Settings.model_validate({"max_sessions": 10}).max_pending == 2


def test_explicit_max_pending_wins():
    cfg = Settings.model_validate({"max_sessions": 10, "max_pending": 5})
    assert cfg.max_pending == 5


def test_max_sessions_must_be_positive():
    # pydantic may surface the model_post_init error raw (ValueError) or
    # wrapped (ValidationError); both are acceptable startup failures.
    with pytest.raises((ValueError, ValidationError)):
        Settings.model_validate({"max_sessions": 0})


def test_max_pending_must_be_positive():
    with pytest.raises((ValueError, ValidationError)):
        Settings.model_validate({"max_pending": 0})
