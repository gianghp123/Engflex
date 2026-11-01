// Package toeic contains the server-only TOEIC import and attempt documents.
package toeic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

type Content struct {
	SchemaVersion string  `json:"schemaVersion"`
	CollectionID  string  `json:"collectionId"`
	TestID        string  `json:"testId"`
	TestNumber    int     `json:"testNumber"`
	Title         string  `json:"title"`
	AudioAssetID  string  `json:"audioAssetId"`
	Assets        []Asset `json:"assets"`
	Parts         []Part  `json:"parts"`
}
type Asset struct {
	SourceID string `json:"sourceId"`
	SHA256   string `json:"sha256"`
}
type Part struct {
	PartNumber    int    `json:"partNumber"`
	Section       string `json:"section"`
	Items         []Item `json:"items"`
	QuestionCount int    `json:"-"`
}
type Item struct {
	ID            string       `json:"id"`
	Type          string       `json:"type"`
	Questions     []Question   `json:"questions"`
	Passages      []Passage    `json:"passages"`
	Transcript    []Transcript `json:"transcript"`
	ItemAudio     *AudioRange  `json:"itemAudio"`
	StimulusAudio *AudioRange  `json:"stimulusAudio"`
	Media         []Block      `json:"media"`
}
type AudioRange struct {
	AssetID   string   `json:"assetId"`
	StartTime *float64 `json:"startTime"`
	EndTime   *float64 `json:"endTime"`
}
type Passage struct {
	ID     string  `json:"id"`
	Format *string `json:"format"`
	Blocks []Block `json:"blocks"`
}
type Block struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`
	Text           *string    `json:"text"`
	Rows           [][]string `json:"rows"`
	Inline         []Inline   `json:"inline"`
	AssetID        *string    `json:"assetId"`
	VietnameseText *string    `json:"vietnameseText"`
}
type Inline struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Text       *string `json:"text"`
	QuestionID *string `json:"questionId"`
}
type Transcript struct {
	ID             string  `json:"id"`
	Role           string  `json:"role"`
	Speaker        *string `json:"speaker"`
	OptionKey      *string `json:"optionKey"`
	Text           string  `json:"text"`
	VietnameseText *string `json:"vietnameseText"`
}
type Question struct {
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
type Option struct {
	Key  string  `json:"key"`
	Text *string `json:"text"`
}
type Vocabulary struct {
	Word    *string `json:"word"`
	POS     *string `json:"pos"`
	Meaning *string `json:"meaning"`
	Raw     *string `json:"raw"`
}

// Result is specific to exam attempts; lesson/video result documents remain untouched.
type Result struct {
	ScoringVersion *string         `json:"scoringVersion"`
	Sections       *ResultSections `json:"sections"`
	QuestionCount  int             `json:"-"`
	Revision       int             `json:"revision"`
	Answers        []Answer        `json:"answers"`
	CorrectCount   *int            `json:"correctCount"`
	AccuracyPct    *float64        `json:"accuracyPct"`
	// Deprecated: the total is stored only in Attempt.Score; this stays null.
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

// ParseResult only handles exam attempt JSONB and does not alter the lesson or video parser.
func ParseResult(raw []byte) (*Result, error) {
	var result Result
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode TOEIC result: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("invalid trailing result")
	}
	if result.Revision < 1 || result.Answers == nil {
		return nil, fmt.Errorf("invalid TOEIC result revision or answers")
	}
	seen := make(map[string]bool, len(result.Answers))
	for _, answer := range result.Answers {
		if !present(answer.QuestionID) || seen[answer.QuestionID] || (answer.SelectedKey != nil && (len(*answer.SelectedKey) != 1 || !strings.Contains("ABCD", *answer.SelectedKey))) {
			return nil, fmt.Errorf("invalid TOEIC result answer")
		}
		seen[answer.QuestionID] = true
	}
	if result.CorrectCount != nil && (*result.CorrectCount < 0 || *result.CorrectCount > len(result.Answers)) {
		return nil, fmt.Errorf("invalid TOEIC correct count")
	}
	if result.AccuracyPct != nil && (math.IsNaN(*result.AccuracyPct) || *result.AccuracyPct < 0 || *result.AccuracyPct > 100) {
		return nil, fmt.Errorf("invalid TOEIC accuracy")
	}
	result.QuestionCount = len(result.Answers)
	if err := validateScoring(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ParseContent decodes JSONB explicitly and rejects incomplete or inconsistent exams.
func ParseContent(raw []byte) (*Content, error) {
	var c Content
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return nil, fmt.Errorf("decode TOEIC content: %w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("invalid trailing content")
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	for i := range c.Parts {
		for j := range c.Parts[i].Items {
			item := &c.Parts[i].Items[j]
			c.Parts[i].QuestionCount += len(item.Questions)
			if item.Passages == nil {
				item.Passages = []Passage{}
			}
			if item.Transcript == nil {
				item.Transcript = []Transcript{}
			}
			if item.Media == nil {
				item.Media = []Block{}
			}
			for k := range item.Questions {
				if item.Questions[k].Vocabulary == nil {
					item.Questions[k].Vocabulary = []Vocabulary{}
				}
				if item.Questions[k].TargetIDs == nil {
					item.Questions[k].TargetIDs = []string{}
				}
			}
			normalizeBlocks(item.Media)
			for k := range item.Passages {
				normalizeBlocks(item.Passages[k].Blocks)
			}
		}
	}
	return &c, nil
}

func normalizeBlocks(blocks []Block) {
	for i := range blocks {
		if blocks[i].Inline == nil {
			blocks[i].Inline = []Inline{}
		}
	}
}

var shaPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

func present(s string) bool      { return strings.TrimSpace(s) != "" }
func textPresent(s *string) bool { return s != nil && present(*s) }

// Validate checks the export graph; media existence and duration require external verification.
func (c *Content) Validate() error {
	fail := func(s string) error { return fmt.Errorf("invalid TOEIC content: %s", s) }
	if c.SchemaVersion != "0.1.0" || !present(c.CollectionID) || !present(c.TestID) || c.TestNumber < 1 || !present(c.Title) || !present(c.AudioAssetID) {
		return fail("identity or version")
	}
	assets := map[string]bool{}
	for _, a := range c.Assets {
		if !present(a.SourceID) || assets[a.SourceID] || !shaPattern.MatchString(a.SHA256) {
			return fail("asset identity or hash")
		}
		assets[a.SourceID] = true
	}
	if !assets[c.AudioAssetID] {
		return fail("missing audio asset")
	}
	ids := map[string]bool{}
	add := func(id string) bool {
		if !present(id) || ids[id] {
			return false
		}
		ids[id] = true
		return true
	}
	bounds := [][2]int{{1, 6}, {7, 31}, {32, 70}, {71, 100}, {101, 130}, {131, 146}, {147, 200}}
	if len(c.Parts) != 7 {
		return fail("expected seven parts")
	}
	seenParts := map[int]bool{}
	numbers := map[int]bool{}
	previousEnd := 0.0
	for index, p := range c.Parts {
		if p.PartNumber != index+1 || seenParts[p.PartNumber] {
			return fail("part number or order")
		}
		seenParts[p.PartNumber] = true
		if (p.PartNumber <= 4 && p.Section != "listening") || (p.PartNumber > 4 && p.Section != "reading") || len(p.Items) == 0 {
			return fail("part section or items")
		}
		for _, item := range p.Items {
			if !add(item.ID) || (item.Type != "single" && item.Type != "group") || len(item.Questions) == 0 {
				return fail("item identity or type")
			}
			if (p.PartNumber == 1 || p.PartNumber == 2 || p.PartNumber == 5) && (item.Type != "single" || len(item.Questions) != 1) {
				return fail("single item")
			}
			if (p.PartNumber == 3 || p.PartNumber == 4) && (item.Type != "group" || len(item.Questions) != 3) {
				return fail("listening group")
			}
			if p.PartNumber == 6 && (item.Type != "group" || len(item.Questions) != 4) {
				return fail("part 6 group")
			}
			if p.PartNumber == 7 && item.Type != "group" {
				return fail("part 7 group")
			}
			if p.PartNumber == 7 {
				first := item.Questions[0].QuestionNumber
				expected := 1
				if first >= 176 && first <= 185 {
					expected = 2
				} else if first >= 186 {
					expected = 3
				}
				if len(item.Questions) < 2 || len(item.Questions) > 5 || len(item.Passages) != expected {
					return fail("part 7 group or passage count")
				}
				for _, q := range item.Questions {
					if (first <= 175 && q.QuestionNumber > 175) || (first >= 176 && first <= 185 && (q.QuestionNumber < 176 || q.QuestionNumber > 185)) || (first >= 186 && q.QuestionNumber < 186) {
						return fail("part 7 group range")
					}
				}
			}
			if p.PartNumber <= 4 {
				a := item.ItemAudio
				if a == nil || a.StartTime == nil || a.EndTime == nil || a.AssetID != c.AudioAssetID || *a.StartTime < previousEnd || *a.StartTime < 0 || *a.EndTime <= *a.StartTime {
					return fail("item audio")
				}
				previousEnd = *a.EndTime
				if item.StimulusAudio != nil {
					s := item.StimulusAudio
					if s.StartTime == nil || s.EndTime == nil || s.AssetID != a.AssetID || *s.StartTime < *a.StartTime || *s.EndTime > *a.EndTime || *s.EndTime <= *s.StartTime {
						return fail("stimulus audio")
					}
				}
			} else if item.ItemAudio != nil || item.StimulusAudio != nil {
				return fail("reading audio")
			}
			questionIDs := map[string]bool{}
			nodes := map[string]Inline{}
			for _, q := range item.Questions {
				if !add(q.ID) || q.QuestionNumber < bounds[p.PartNumber-1][0] || q.QuestionNumber > bounds[p.PartNumber-1][1] || numbers[q.QuestionNumber] {
					return fail("question identity or number")
				}
				numbers[q.QuestionNumber] = true
				questionIDs[q.ID] = true
				keys := "ABCD"
				if p.PartNumber == 2 {
					keys = "ABC"
				}
				if len(q.Options) != len(keys) {
					return fail("option count")
				}
				correct := false
				for i, o := range q.Options {
					if o.Key != string(keys[i]) || (p.PartNumber >= 3 && !textPresent(o.Text)) {
						return fail("option key or text")
					}
					if q.CorrectKey != nil && *q.CorrectKey == o.Key {
						correct = true
					}
				}
				if !correct || (p.PartNumber >= 3 && p.PartNumber != 6 && !textPresent(q.Stem)) {
					return fail("answer or stem")
				}
			}
			checkBlock := func(b Block) bool {
				if !add(b.ID) {
					return false
				}
				switch b.Type {
				case "paragraph", "heading", "emailHeader", "chatMessage":
					if !textPresent(b.Text) && len(b.Inline) == 0 {
						return false
					}
				case "table":
					if len(b.Rows) == 0 && (b.AssetID == nil || !assets[*b.AssetID]) {
						return false
					}
				case "image":
					if b.AssetID == nil || !assets[*b.AssetID] {
						return false
					}
				default:
					return false
				}
				if b.AssetID != nil && !assets[*b.AssetID] {
					return false
				}
				for _, n := range b.Inline {
					if !add(n.ID) {
						return false
					}
					switch n.Type {
					case "text", "markedText":
						if !textPresent(n.Text) {
							return false
						}
					case "blank", "insertionPoint":
						if textPresent(n.Text) {
							return false
						}
					default:
						return false
					}
					if n.QuestionID != nil && !questionIDs[*n.QuestionID] {
						return false
					}
					nodes[n.ID] = n
				}
				return true
			}
			for _, b := range item.Media {
				if !checkBlock(b) {
					return fail("media block")
				}
			}
			if p.PartNumber == 1 {
				hasImage := false
				for _, b := range item.Media {
					if b.Type == "image" {
						hasImage = true
					}
				}
				if !hasImage {
					return fail("part 1 image")
				}
			}
			for _, passage := range item.Passages {
				if !add(passage.ID) || len(passage.Blocks) == 0 {
					return fail("passage")
				}
				for _, b := range passage.Blocks {
					if !checkBlock(b) {
						return fail("passage block")
					}
				}
			}
			if (p.PartNumber == 6 || p.PartNumber == 7) && len(item.Passages) == 0 {
				return fail("missing passage")
			}
			for _, q := range item.Questions {
				seen := map[string]bool{}
				for _, id := range q.TargetIDs {
					n, ok := nodes[id]
					if !ok || seen[id] || (n.QuestionID != nil && *n.QuestionID != q.ID) {
						return fail("question target")
					}
					seen[id] = true
				}
				if p.PartNumber == 6 {
					if len(q.TargetIDs) != 1 || nodes[q.TargetIDs[0]].Type != "blank" || nodes[q.TargetIDs[0]].QuestionID == nil {
						return fail("part 6 blank")
					}
					linked := 0
					for _, n := range nodes {
						if n.Type == "blank" && n.QuestionID != nil && *n.QuestionID == q.ID {
							linked++
						}
					}
					if linked != 1 {
						return fail("part 6 duplicate blank")
					}
				}
			}
			for _, t := range item.Transcript {
				if !add(t.ID) || !present(t.Text) || (t.Role != "prompt" && t.Role != "option" && t.Role != "utterance") {
					return fail("transcript")
				}
				if t.OptionKey != nil && (len(*t.OptionKey) != 1 || !strings.Contains("ABCD", *t.OptionKey) || (p.PartNumber == 2 && *t.OptionKey == "D")) {
					return fail("transcript option")
				}
			}
		}
	}
	if len(numbers) != 200 {
		return fail("question coverage")
	}
	return nil
}
