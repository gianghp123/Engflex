package enums

// ScenarioDifficulty is the scenario difficulty value (scenarios.cefr_level).
// Separate from CEFR because "B1+" is a product half-step, not a CEFR level.
type ScenarioDifficulty string

const (
	ScenarioDifficultyA1     ScenarioDifficulty = "A1"
	ScenarioDifficultyA2     ScenarioDifficulty = "A2"
	ScenarioDifficultyB1     ScenarioDifficulty = "B1"
	ScenarioDifficultyB1Plus ScenarioDifficulty = "B1+"
	ScenarioDifficultyB2     ScenarioDifficulty = "B2"
	ScenarioDifficultyC1     ScenarioDifficulty = "C1"
)
