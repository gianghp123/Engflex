package enums

// PublishStatus gates admin-curated vocabulary content (shared dictionary
// entries, system decks) from learner-facing reads and SRS until reviewed.
type PublishStatus string

const (
	PublishStatusDraft     PublishStatus = "draft"
	PublishStatusPublished PublishStatus = "published"
)
