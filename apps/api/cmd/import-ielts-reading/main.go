// Command import-ielts-reading validates a reviewed bundle and emits a
// transaction that can be applied with psql. Test versions are immutable.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"engflex-api/internal/exams/ielts"
)

func main() {
	path := flag.String("bundle", "", "path to the normalized IELTS Reading JSON")
	reviewer := flag.String("reviewed-by", "", "person who verified the passages and answer key")
	version := flag.Int("version", 1, "immutable revision of this test")
	flag.Parse()
	if err := run(*path, *reviewer, *version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path, reviewer string, version int) error {
	if path == "" || strings.TrimSpace(reviewer) == "" {
		return fmt.Errorf("both -bundle and -reviewed-by are required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := ielts.ParseContent(raw)
	if err != nil {
		return err
	}
	if !regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`).MatchString(content.TestID) {
		return fmt.Errorf("bundle ID must be a lowercase slug")
	}
	if version < 1 || content.Skill != "reading" || content.SubType != "general_training" || len(content.Sections) != 3 || strings.TrimSpace(content.Title) == "" {
		return fmt.Errorf("expected a titled General Training Reading test with three sections and a positive version")
	}
	number := 0
	for i, section := range content.Sections {
		if section.SectionNumber != i+1 || len(section.Passages) == 0 || len(section.QuestionGroups) == 0 {
			return fmt.Errorf("incomplete section %d", i+1)
		}
		for _, passage := range section.Passages {
			if len(passage.Content) == 0 {
				return fmt.Errorf("empty passage %q", passage.ID)
			}
			for _, paragraph := range passage.Content {
				if strings.TrimSpace(paragraph.Text) == "" {
					return fmt.Errorf("empty paragraph in passage %q", passage.ID)
				}
			}
		}
		for _, group := range section.QuestionGroups {
			if len(group.Questions) == 0 {
				return fmt.Errorf("empty question group %q", group.ID)
			}
			for _, question := range group.Questions {
				number++
				if question.QuestionNumber != number || question.Stem == nil || strings.TrimSpace(*question.Stem) == "" {
					return fmt.Errorf("invalid question %d", number)
				}
			}
		}
	}
	if number != 40 {
		return fmt.Errorf("expected 40 questions, got %d", number)
	}
	key := quote(fmt.Sprintf("ielts-reading-%s-v%d", content.TestID, version))
	body := quote(string(raw))
	category := "'ielts-general-training-reading'"
	_, err = fmt.Fprintf(os.Stdout, `BEGIN;
SET LOCAL standard_conforming_strings = on;
INSERT INTO exam_categories (slug, title, type, sub_type)
VALUES (%[1]s, 'IELTS General Training Reading', 'ielts', 'general_training')
ON CONFLICT (slug) DO NOTHING;
INSERT INTO exam_tests (category_id, source_key, slug, title, content, skill, sub_type)
SELECT id, %[2]s, %[2]s, %[3]s, %[4]s::jsonb, 'reading', 'general_training'
FROM exam_categories WHERE slug = %[1]s AND type = 'ielts' AND sub_type = 'general_training'
ON CONFLICT (source_key) DO NOTHING;
CREATE TEMP TABLE ielts_import_check (ok boolean CHECK (ok)) ON COMMIT DROP;
INSERT INTO ielts_import_check VALUES (EXISTS (
    SELECT 1 FROM exam_tests t JOIN exam_categories c ON c.id = t.category_id
    WHERE c.slug = %[1]s AND c.type = 'ielts' AND c.sub_type = 'general_training'
      AND t.source_key = %[2]s AND t.title = %[3]s AND t.content = %[4]s::jsonb
));
COMMIT;
`, category, key, quote(content.Title), body)
	return err
}

func quote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
