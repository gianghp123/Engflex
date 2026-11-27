from pydantic import BaseModel, Field


class PersonaBody(BaseModel):
    """Structured persona input; the engine owns the template."""

    name: str = ""
    roleTitle: str = ""
    personality: str = ""
    style: str = ""
    objective: str = ""
    gender: str = ""


class ObjectiveBody(BaseModel):
    """One roleplay task. ID is the 1-based position, local to the scenario."""

    id: int = 0
    label: str = ""
    hint: str = ""


class ScenarioBody(BaseModel):
    title: str = ""
    objective: str = ""
    cefrLevel: str = "B1"
    objectives: list[ObjectiveBody] = Field(default_factory=list)


class LearnerBody(BaseModel):
    level: str = "B1"
    goals: list[str] = Field(default_factory=list)


class RunnerBody(BaseModel):
    """Session payload Go sends in the /start body."""

    userId: str
    conversationId: str
    maxDuration: int = Field(default=300, ge=30, le=3600)
    persona: PersonaBody | None = None
    scenario: ScenarioBody | None = None
    learner: LearnerBody | None = None


class TranscriptCommandBody(BaseModel):
    """One transcript command.

    Three actions on the reviewed turn, so they are one endpoint:

    - ``review``  hand back the latest learner turn and mark it as the one
      under review; the reply carries the engine's own copy of that text, so
      the modal edits exactly what will be rewritten.
    - ``send``    rewrite the marked turn with ``text`` and regenerate.
    - ``dismiss`` forget the mark. Records nothing.

    The client never names a turn: its row ordinal and the engine's collector
    position are separate bookkeeping, so the engine resolves the target.
    """

    conversationId: str
    action: str
    text: str | None = None
