package services

import (
	"context"
	"encoding/json"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/exams/dtos/requests"
	"errors"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

func (s *AttemptService) Patch(ctx context.Context, userID, id string, req requests.PatchAttempt) (*models.Attempt, error) {
	if req.Revision < 1 || len(req.Answers) > 200 {
		return nil, common.BadRequest("invalid revision or batch size")
	}
	return s.change(ctx, userID, id, &req, false)
}

func (s *AttemptService) Cancel(ctx context.Context, userID, id string) (*models.Attempt, error) {
	return s.change(ctx, userID, id, nil, true)
}

func (s *AttemptService) change(ctx context.Context, userID, id string, req *requests.PatchAttempt, cancel bool) (*models.Attempt, error) {
	var m *models.Attempt
	var afterCommit error
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		repo := s.attempts.WithTx(tx)
		var err error
		m, err = repo.GetExamAttemptForUpdate(ctx, userID, id)
		if err != nil {
			return common.FromDBError(err, "attempt")
		}
		now := time.Now().UTC()
		if m.ExamTest == nil || m.ExamTest.Category == nil || m.ExamTest.Category.Type != enums.ExamTypeTOEIC || m.StartedAt == nil {
			return common.Internal()
		}
		m.ExamTest.ParsedContent, err = toeic.ParseContent(m.ExamTest.Content)
		if err != nil {
			return common.Internal()
		}
		m.ParsedResult, err = toeic.ParseResult(m.Result)
		if err != nil {
			return common.Internal()
		}
		filterSelectedParts(m.ExamTest.ParsedContent, m.SelectedParts)
		if m.Status == enums.AttemptStatusCompleted {
			if req != nil && !req.Submit {
				return common.Conflict("attempt is completed")
			}
			return nil
		}
		if cancel && m.Status == enums.AttemptStatusAbandoned {
			return nil
		}
		if m.Status != enums.AttemptStatusInProgress {
			return common.Conflict("attempt is closed")
		}
		previous := m.ParsedResult.Revision
		expired := m.DeadlineAt != nil && !now.Before(*m.DeadlineAt)
		if cancel && !expired {
			duration := max(0, int(now.Sub(*m.StartedAt).Seconds()))
			m.Status = enums.AttemptStatusAbandoned
			m.DurationSec = &duration
			m.ParsedResult.Revision++
			m.Result, err = json.Marshal(m.ParsedResult)
			if err != nil {
				return common.Internal()
			}
			return repo.SaveExamAttempt(ctx, m, previous)
		}
		if !expired && req == nil {
			return nil
		}
		questions := map[string]toeic.Question{}
		for _, p := range m.ExamTest.ParsedContent.Parts {
			for _, item := range p.Items {
				for _, q := range item.Questions {
					questions[q.ID] = q
				}
			}
		}
		indices := map[string]int{}
		for i, a := range m.ParsedResult.Answers {
			if _, ok := questions[a.QuestionID]; !ok {
				return common.Internal()
			}
			indices[a.QuestionID] = i
		}
		if len(indices) != len(questions) {
			return common.Internal()
		}
		if !expired {
			if req.Revision != previous {
				return common.Conflict("stale revision; GET the attempt again")
			}
			seen := map[string]bool{}
			for _, answer := range req.Answers {
				index, ok := indices[answer.QuestionID]
				if !ok || seen[answer.QuestionID] {
					return common.BadRequest("unknown or duplicate question")
				}
				seen[answer.QuestionID] = true
				if answer.SelectedKey != nil {
					valid := false
					for _, option := range questions[answer.QuestionID].Options {
						if option.Key == *answer.SelectedKey {
							valid = true
						}
					}
					if !valid {
						return common.BadRequest("invalid option for question")
					}
				}
				m.ParsedResult.Answers[index].SelectedKey = answer.SelectedKey
				m.ParsedResult.Answers[index].Flagged = answer.Flagged
			}
		}
		if expired || req.Submit {
			finalize(m, questions, now)
		}
		m.ParsedResult.Revision++
		m.Result, err = json.Marshal(m.ParsedResult)
		if err != nil {
			return common.Internal()
		}
		if err := repo.SaveExamAttempt(ctx, m, previous); err != nil {
			return err
		}
		if expired && req != nil && !req.Submit {
			afterCommit = common.Conflict("deadline passed; attempt finalized")
		}
		return nil
	})
	if err != nil {
		var appErr *common.AppError
		if !errors.As(err, &appErr) {
			appErr = common.FromDBError(err, "attempt")
		}
		logger.Report(ctx, "exam attempt operation failed", appErr, "attemptID", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "exam attempt operation done", "attemptID", id, "status", m.Status)
	if afterCommit != nil {
		return nil, afterCommit
	}
	return m, nil
}

func finalize(m *models.Attempt, questions map[string]toeic.Question, now time.Time) {
	end := now
	if m.DeadlineAt != nil && end.After(*m.DeadlineAt) {
		end = *m.DeadlineAt
	}
	correct := 0
	correctByID := make(map[string]bool, len(m.ParsedResult.Answers))
	for _, a := range m.ParsedResult.Answers {
		q := questions[a.QuestionID]
		if a.SelectedKey != nil && q.CorrectKey != nil && *a.SelectedKey == *q.CorrectKey {
			correct++
			correctByID[a.QuestionID] = true
		}
	}
	accuracy := 100 * float64(correct) / float64(len(questions))
	duration := int(end.Sub(*m.StartedAt).Seconds())
	if duration < 0 {
		duration = 0
	}
	m.Status = enums.AttemptStatusCompleted
	m.SubmittedAt = &end
	m.DurationSec = &duration
	m.Score = nil
	m.ParsedResult.CorrectCount = &correct
	m.ParsedResult.AccuracyPct = &accuracy
	m.ParsedResult.ScaledScore = nil
	version := toeic.ScoringVersion
	m.ParsedResult.ScoringVersion = &version
	sections := &toeic.ResultSections{}
	for _, part := range m.ExamTest.ParsedContent.Parts {
		section := &sections.Reading
		if part.Section == "listening" {
			section = &sections.Listening
		}
		if *section == nil {
			*section = &toeic.SectionResult{}
		}
		for _, item := range part.Items {
			for _, question := range item.Questions {
				(*section).QuestionCount++
				if correctByID[question.ID] {
					(*section).CorrectCount++
				}
			}
		}
	}
	for _, section := range []*toeic.SectionResult{sections.Listening, sections.Reading} {
		if section != nil && section.QuestionCount == 100 {
			score, _ := toeic.EstimateScore(section.CorrectCount)
			section.ScaledScore = &score
		}
	}
	m.ParsedResult.Sections = sections
	if sections.Listening != nil && sections.Reading != nil && sections.Listening.ScaledScore != nil && sections.Reading.ScaledScore != nil {
		score := *sections.Listening.ScaledScore + *sections.Reading.ScaledScore
		m.Score = &score
	}
}
