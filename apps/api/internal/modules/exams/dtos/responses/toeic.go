// Package responses contains wire-only exam DTOs. Taking types intentionally
// exclude all answer-bearing and translated fields.
package responses

type TestDetail struct {
	ID            string           `json:"id"`
	CategoryID    string           `json:"categoryId"`
	Slug          string           `json:"slug"`
	Title         string           `json:"title"`
	ParsedContent *ContentMetadata `json:"content"`
	Category      *Category        `json:"category"`
}

type TestMetadata struct {
	ID            string           `json:"id"`
	CategoryID    string           `json:"categoryId"`
	Slug          string           `json:"slug"`
	Title         string           `json:"title"`
	ParsedContent *ContentMetadata `json:"content"`
}
type ContentMetadata struct {
	Parts []PartMetadata `json:"parts"`
}
type PartMetadata struct {
	PartNumber    int    `json:"partNumber"`
	Section       string `json:"section"`
	QuestionCount int    `json:"questionCount"`
}

type TakingContent struct {
	SchemaVersion string       `json:"schemaVersion"`
	CollectionID  string       `json:"collectionId"`
	TestID        string       `json:"testId"`
	TestNumber    int          `json:"testNumber"`
	Title         string       `json:"title"`
	AudioAssetID  string       `json:"audioAssetId"`
	Assets        []Asset      `json:"assets"`
	Parts         []TakingPart `json:"parts"`
}
type Asset struct {
	SourceID string `json:"sourceId"`
	SHA256   string `json:"sha256"`
}
type TakingPart struct {
	PartNumber int          `json:"partNumber"`
	Section    string       `json:"section"`
	Items      []TakingItem `json:"items"`
}
type TakingItem struct {
	ID            string           `json:"id"`
	Type          string           `json:"type"`
	Questions     []TakingQuestion `json:"questions"`
	Passages      []TakingPassage  `json:"passages"`
	ItemAudio     *AudioRange      `json:"itemAudio"`
	StimulusAudio *AudioRange      `json:"stimulusAudio"`
	Media         []TakingBlock    `json:"media"`
}
type AudioRange struct {
	AssetID   string   `json:"assetId"`
	StartTime *float64 `json:"startTime"`
	EndTime   *float64 `json:"endTime"`
}
type TakingPassage struct {
	ID     string        `json:"id"`
	Format *string       `json:"format"`
	Blocks []TakingBlock `json:"blocks"`
}
type TakingBlock struct {
	ID      string     `json:"id"`
	Type    string     `json:"type"`
	Text    *string    `json:"text"`
	Rows    [][]string `json:"rows" tstype:"string[][] | null"`
	Inline  []Inline   `json:"inline"`
	AssetID *string    `json:"assetId"`
}
type Inline struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Text       *string `json:"text"`
	QuestionID *string `json:"questionId"`
}
type TakingQuestion struct {
	ID             string   `json:"id"`
	QuestionNumber int      `json:"questionNumber"`
	Stem           *string  `json:"stem"`
	Options        []Option `json:"options"`
	TargetIDs      []string `json:"targetIds"`
}
type Option struct {
	Key  string  `json:"key"`
	Text *string `json:"text"`
}

// ReviewContent must only be mapped after authorization and completion.
type ReviewContent struct {
	SchemaVersion string       `json:"schemaVersion"`
	CollectionID  string       `json:"collectionId"`
	TestID        string       `json:"testId"`
	TestNumber    int          `json:"testNumber"`
	Title         string       `json:"title"`
	AudioAssetID  string       `json:"audioAssetId"`
	Assets        []Asset      `json:"assets"`
	Parts         []ReviewPart `json:"parts"`
}
type ReviewPart struct {
	PartNumber int          `json:"partNumber"`
	Section    string       `json:"section"`
	Items      []ReviewItem `json:"items"`
}
type ReviewItem struct {
	ID            string           `json:"id"`
	Type          string           `json:"type"`
	Questions     []ReviewQuestion `json:"questions"`
	Passages      []ReviewPassage  `json:"passages"`
	Transcript    []Transcript     `json:"transcript"`
	ItemAudio     *AudioRange      `json:"itemAudio"`
	StimulusAudio *AudioRange      `json:"stimulusAudio"`
	Media         []ReviewBlock    `json:"media"`
}
type ReviewPassage struct {
	ID     string        `json:"id"`
	Format *string       `json:"format"`
	Blocks []ReviewBlock `json:"blocks"`
}
type ReviewBlock struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Text           *string    `json:"text"`
	Rows           [][]string `json:"rows" tstype:"string[][] | null"`
	Inline         []Inline   `json:"inline"`
	AssetID        *string    `json:"assetId"`
	VietnameseText *string    `json:"vietnameseText"`
}
type Transcript struct {
	ID             string  `json:"id"`
	Role           string  `json:"role"`
	Speaker        *string `json:"speaker"`
	OptionKey      *string `json:"optionKey"`
	Text           string  `json:"text"`
	VietnameseText *string `json:"vietnameseText"`
}
type ReviewQuestion struct {
	ID             string       `json:"id"`
	QuestionNumber int          `json:"questionNumber"`
	Stem           *string      `json:"stem"`
	Options        []Option     `json:"options"`
	CorrectKey     *string      `json:"correctKey"`
	Explanation    *string      `json:"explanation"`
	VietnameseText *string      `json:"vietnameseText"`
	Vocabulary     []Vocabulary `json:"vocabulary"`
	TargetIDs      []string     `json:"targetIds"`
}
type Vocabulary struct {
	Word    *string `json:"word"`
	POS     *string `json:"pos"`
	Meaning *string `json:"meaning"`
	Raw     *string `json:"raw"`
}
type Result struct {
	ScoringVersion *string         `json:"scoringVersion"`
	Sections       *ResultSections `json:"sections"`
	Revision       int             `json:"revision"`
	Answers        []Answer        `json:"answers"`
	CorrectCount   *int            `json:"correctCount"`
	AccuracyPct    *float64        `json:"accuracyPct"`
	// Deprecated: read the total from the parent attempt's score instead.
	ScaledScore *float64 `json:"scaledScore"`
}
type ResultSections struct {
	Listening *SectionResult `json:"listening"`
	Reading   *SectionResult `json:"reading"`
}
type SectionResult struct {
	QuestionCount int      `json:"questionCount"`
	CorrectCount  int      `json:"correctCount"`
	ScaledScore   *float64 `json:"scaledScore"`
}
type Answer struct {
	QuestionID  string  `json:"questionId"`
	SelectedKey *string `json:"selectedKey"`
	Flagged     bool    `json:"flagged"`
}
