package services

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"engflex-api/internal/common"
	"engflex-api/internal/llm"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/prompts"
	"engflex-api/internal/modules/vocabulary/schemas"
	"engflex-api/internal/utils"
)

// VocabularyWordDraftService generates a draft word entry with the LLM. It
// writes nothing: the dialog shows the draft, the learner edits it, and the
// existing add-word endpoint persists the reviewed result.
type VocabularyWordDraftService interface {
	Generate(ctx context.Context, req requests.GenerateWordDraft) (*responses.WordDraft, error)
}

type vocabularyWordDraftService struct {
	llm llm.LLMClient
}

// NewVocabularyWordDraftService builds the service over the shared LLM client.
func NewVocabularyWordDraftService(client llm.LLMClient) VocabularyWordDraftService {
	return &vocabularyWordDraftService{llm: client}
}

func (s *vocabularyWordDraftService) Generate(ctx context.Context, req requests.GenerateWordDraft) (*responses.WordDraft, error) {
	term := utils.TidyTerm(req.Term)
	if term == "" {
		return nil, common.BadRequest("term is required")
	}
	languageName, ok := explanationLanguage(req.OutputLanguage)
	if !ok {
		return nil, common.BadRequest("unsupported output language")
	}
	prompt := prompts.BuildWordDraftPrompt(term, strings.TrimSpace(req.Context), languageName)
	var out schemas.WordDraft
	if err := s.llm.CompleteStructured(ctx, prompt, &out, schemas.WordDraftSchemaName, schemas.WordDraftSchema); err != nil {
		appErr := common.ServiceUnavailable("generation unavailable")
		logger.Report(ctx, "word draft engine call failed", appErr, "term", term)
		return nil, appErr
	}
	normalized, nErr := schemas.NormalizeWordDraft(out)
	if nErr != nil {
		appErr := common.ServiceUnavailable("generation unavailable")
		logger.Report(ctx, "word draft rejected", appErr, "term", term, "reason", nErr.Error())
		return nil, appErr
	}
	var dto responses.WordDraft
	if mErr := utils.Map(&dto, normalized); mErr != nil {
		logger.Report(ctx, "word draft map failed", mErr, "term", term)
		return nil, common.Internal()
	}
	slog.InfoContext(ctx, "word draft generated", "term", term, "outputLanguage", languageName)
	return &dto, nil
}

// explanationLanguage resolves the request's BCP-47 code to the English name
// of the language the explanation fields are written in ("" = English, the
// default brief). Any assigned language is accepted — the web picker loads
// its options from the app's configured locales, so adding a locale needs no
// change here. Unassigned tags ("zz"), malformed tags ("not a tag") and the
// non-language codes are refused: better a 400 than a brief the model cannot
// interpret.
func explanationLanguage(code string) (string, bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", true
	}
	tag, err := language.Parse(code)
	if err != nil || slices.Contains(nonLanguageTags, tag) {
		return "", false
	}
	name := display.English.Languages().Name(tag)
	if name == "" || strings.EqualFold(name, code) {
		return "", false
	}
	return name, true
}

// nonLanguageTags are well-formed codes that name no single language; the
// model cannot write explanations in them.
var nonLanguageTags = []language.Tag{
	language.MustParse("und"), // Unknown language
	language.MustParse("mul"), // Multiple languages
	language.MustParse("zxx"), // No linguistic content
}
