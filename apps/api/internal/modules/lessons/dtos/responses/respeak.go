package responses

// RespeakResponse is one targeted re-speak score. Nothing is persisted: the
// panel shows the gated list plus the pre-gate count, and the top error's
// coaching tip travels at the response level because the card does not render
// a full word-by-word panel.
type RespeakResponse struct {
	Score       float64             `json:"score"`
	Errors      []MispronouncedWord `json:"errors"`
	TotalErrors int                 `json:"totalErrors"`
	Tip         string              `json:"tip"`
}
