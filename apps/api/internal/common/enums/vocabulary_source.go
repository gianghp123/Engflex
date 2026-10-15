package enums

// VocabularySource is how a term entered the user's vocabulary set
// (user_vocabulary.source_type).
type VocabularySource string

const (
	VocabularySourceLesson       VocabularySource = "lesson"
	VocabularySourceConversation VocabularySource = "conversation"
	VocabularySourceVideo        VocabularySource = "video"
	VocabularySourceManual       VocabularySource = "manual"
)
