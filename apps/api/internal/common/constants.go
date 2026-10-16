package common

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100

	// Shared wire limits. Binding tags on request DTOs must match these
	// (guarded by a limits_test.go per module).
	NameMaxLength    = 255
	TitleMaxLength   = 255
	TermMaxLength    = 255
	MeaningMaxLength = 1000
	TextMaxLength    = 10000

	// MaxTurnsPerBatch bounds one engine turn batch. A 45-minute session stays
	// well under 500 turns; the cap is a guard, not a limit.
	MaxTurnsPerBatch = 500

	// Videos per category on the video hub. The HTTP query requires previewK
	// in 1..VideoPreviewMax (guarded by videos limits_test.go); the default
	// only covers internal callers that pass nothing.
	VideoPreviewDefault = 4
	VideoPreviewMax     = 32
)
