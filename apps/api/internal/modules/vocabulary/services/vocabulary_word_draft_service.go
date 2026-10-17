package services

import (
	"context"
	"log/slog"
	"strings"

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
	languageName, ok := llm.ExplanationLanguage(req.Language)
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
	slog.InfoContext(ctx, "word draft generated", "term", term, "language", languageName)
	return &dto, nil
}
