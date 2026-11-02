package callbacks

import "engflex-api/internal/common/enums"

// IngestTurns is the engine's finalize-time turn batch. Every field is typed
// and owned by Go: audio and word metadata are out of scope here, so nothing
// has an unknown shape.
type IngestTurns struct {
	EngineSessionID string       `json:"engineSessionId" binding:"required"`
	Revision        int64        `json:"revision" binding:"required,min=1"`
	Turns           []IngestTurn `json:"turns" binding:"required,max=500,dive"`
}

// IngestTurn is one transcript line from the engine.
type IngestTurn struct {
	Position       int            `json:"position" binding:"required,min=1"`
	Role           enums.TurnRole `json:"role" binding:"required,oneof=user ai"`
	Text           string         `json:"text" binding:"required,max=10000"`
	WasInterrupted bool           `json:"wasInterrupted"`
}
