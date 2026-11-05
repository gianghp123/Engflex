package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/ielts"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/exams/dtos/requests"
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
		if m.ExamTest == nil || m.ExamTest.Category == nil || m.StartedAt == nil {
			return common.Internal()
		}

		switch m.ExamTest.Category.Type {
		case enums.ExamTypeTOEIC:
			afterCommit, err = s.changeToeic(ctx, repo, m, req, cancel, now)
			return err
		case enums.ExamTypeIELTS:
			afterCommit, err = s.changeIelts(ctx, repo, m, req, cancel, now)
			return err
		default:
			return common.Internal()
		}
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

func (s *AttemptService) changeToeic(ctx context.Context, repo repositories.AttemptRepository, m *models.Attempt, req *requests.PatchAttempt, cancel bool, now time.Time) (error, error) {
	var err error
	m.ExamTest.ParsedContent, err = toeic.ParseContent(m.ExamTest.Content)
	if err != nil {
		return nil, common.Internal()
	}
	m.ParsedResult, err = toeic.ParseResult(m.Result)
	if err != nil {
		return nil, common.Internal()
	}
	filterSelectedParts(m.ExamTest.ParsedContent, m.SelectedParts)
	if m.Status == enums.AttemptStatusCompleted {
		if req != nil && !req.Submit {
			return nil, common.Conflict("attempt is completed")
		}
		return nil, nil
	}
	if cancel && m.Status == enums.AttemptStatusAbandoned {
		return nil, nil
	}
	if m.Status != enums.AttemptStatusInProgress {
		return nil, common.Conflict("attempt is closed")
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
			return nil, common.Internal()
		}
		return nil, repo.SaveExamAttempt(ctx, m, previous)
	}
	if !expired && req == nil {
		return nil, nil
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
			return nil, common.Internal()
		}
		indices[a.QuestionID] = i
	}
	if len(indices) != len(questions) {
		return nil, common.Internal()
	}
	if !expired {
		if req.Revision != previous {
			return nil, common.Conflict("stale revision; GET the attempt again")
		}
		seen := map[string]bool{}
		for _, answer := range req.Answers {
			index, ok := indices[answer.QuestionID]
			if !ok || seen[answer.QuestionID] {
				return nil, common.BadRequest("unknown or duplicate question")
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
					return nil, common.BadRequest("invalid option for question")
				}
			}
			m.ParsedResult.Answers[index].SelectedKey = answer.SelectedKey
			m.ParsedResult.Answers[index].Flagged = answer.Flagged
		}
	}
	if expired || req.Submit {
		finalizeToeic(m, questions, now)
	}
	m.ParsedResult.Revision++
	m.Result, err = json.Marshal(m.ParsedResult)
	if err != nil {
		return nil, common.Internal()
	}
	if err := repo.SaveExamAttempt(ctx, m, previous); err != nil {
		return nil, err
	}
	var afterCommit error
	if expired && req != nil && !req.Submit {
		afterCommit = common.Conflict("deadline passed; attempt finalized")
	}
	return afterCommit, nil
}

func (s *AttemptService) changeIelts(ctx context.Context, repo repositories.AttemptRepository, m *models.Attempt, req *requests.PatchAttempt, cancel bool, now time.Time) (error, error) {
	var err error
	m.ExamTest.ParsedIeltsContent, err = ielts.ParseContent(m.ExamTest.Content)
	if err != nil {
		return nil, common.Internal()
	}
	m.ParsedIeltsResult, err = ielts.ParseResult(m.Result)
	if err != nil {
		return nil, common.Internal()
	}
	filterIeltsSections(m.ExamTest.ParsedIeltsContent, m.SelectedParts)

	if m.Status == enums.AttemptStatusCompleted {
		if req != nil && !req.Submit {
			return nil, common.Conflict("attempt is completed")
		}
		return nil, nil
	}
	if cancel && m.Status == enums.AttemptStatusAbandoned {
		return nil, nil
	}
	if m.Status != enums.AttemptStatusInProgress {
		return nil, common.Conflict("attempt is closed")
	}

	previous := m.ParsedIeltsResult.Revision
	expired := m.DeadlineAt != nil && !now.Before(*m.DeadlineAt)

	if cancel && !expired {
		duration := max(0, int(now.Sub(*m.StartedAt).Seconds()))
		m.Status = enums.AttemptStatusAbandoned
		m.DurationSec = &duration
		m.ParsedIeltsResult.Revision++
		m.Result, err = json.Marshal(m.ParsedIeltsResult)
		if err != nil {
			return nil, common.Internal()
		}
		return nil, repo.SaveExamAttempt(ctx, m, previous)
	}

	if !expired && req == nil {
		return nil, nil
	}

	questions := map[string]ielts.Question{}
	for _, sec := range m.ExamTest.ParsedIeltsContent.Sections {
		for _, qg := range sec.QuestionGroups {
			for _, q := range qg.Questions {
				questions[q.ID] = q
			}
		}
	}

	indices := map[string]int{}
	for i, a := range m.ParsedIeltsResult.Answers {
		if _, ok := questions[a.QuestionID]; !ok {
			return nil, common.Internal()
		}
		indices[a.QuestionID] = i
	}
	if len(indices) != len(questions) {
		return nil, common.Internal()
	}

	if !expired {
		if req.Revision != previous {
			return nil, common.Conflict("stale revision; GET the attempt again")
		}
		seen := map[string]bool{}
		for _, answer := range req.Answers {
			index, ok := indices[answer.QuestionID]
			if !ok || seen[answer.QuestionID] {
				return nil, common.BadRequest("unknown or duplicate question")
			}
			seen[answer.QuestionID] = true
			q := questions[answer.QuestionID]
			if answer.SelectedKey != nil && len(q.Options) > 0 {
				valid := false
				for _, option := range q.Options {
					if strings.EqualFold(option.Key, *answer.SelectedKey) {
						valid = true
						break
					}
				}
				if !valid {
					return nil, common.BadRequest("invalid option for question")
				}
			}
			m.ParsedIeltsResult.Answers[index].SelectedKey = answer.SelectedKey
			m.ParsedIeltsResult.Answers[index].SelectedKeys = answer.SelectedKeys
			m.ParsedIeltsResult.Answers[index].TextValue = answer.TextValue
			m.ParsedIeltsResult.Answers[index].Flagged = answer.Flagged
		}
	}

	if expired || req.Submit {
		graded, err := ielts.GradeIeltsAttempt(m.ExamTest.ParsedIeltsContent, m.ParsedIeltsResult.Answers)
		if err != nil {
			return nil, common.Internal()
		}
		end := now
		if m.DeadlineAt != nil && end.After(*m.DeadlineAt) {
			end = *m.DeadlineAt
		}
		duration := max(0, int(end.Sub(*m.StartedAt).Seconds()))
		m.Status = enums.AttemptStatusCompleted
		m.SubmittedAt = &end
		m.DurationSec = &duration
		m.Score = &graded.BandScore
		graded.Revision = previous + 1
		m.ParsedIeltsResult = graded
	} else {
		m.ParsedIeltsResult.Revision++
	}

	m.Result, err = json.Marshal(m.ParsedIeltsResult)
	if err != nil {
		return nil, common.Internal()
	}

	if err := repo.SaveExamAttempt(ctx, m, previous); err != nil {
		return nil, err
	}

	var afterCommit error
	if expired && req != nil && !req.Submit {
		afterCommit = common.Conflict("deadline passed; attempt finalized")
	}
	return afterCommit, nil
}

func finalizeToeic(m *models.Attempt, questions map[string]toeic.Question, now time.Time) {
	end := now
	if m.DeadlineAt != nil && end.After(*m.DeadlineAt) {
		end = *m.DeadlineAt
	}
	correct := 0
	for _, a := range m.ParsedResult.Answers {
		q := questions[a.QuestionID]
		if a.SelectedKey != nil && q.CorrectKey != nil && *a.SelectedKey == *q.CorrectKey {
			correct++
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
}
