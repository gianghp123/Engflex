package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type sqlTrace struct {
	gormlog.Interface
	statements []string
}

func (l *sqlTrace) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	l.statements = append(l.statements, sql)
}

// Opt-in integration coverage. All fixtures, including completed snapshots, roll back.
func TestPostgresExamScoringAndStats(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	trace := &sqlTrace{Interface: gormlog.Default.LogMode(gormlog.Silent)}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: trace})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	tx := db.Begin()
	require.NoError(t, tx.Error)
	t.Cleanup(func() { require.NoError(t, tx.Rollback().Error) })
	ctx := context.Background()
	owner := "stats-" + uuid.NewString()
	cat := &models.ExamCategory{Slug: owner, Title: "Statistics probe", Type: enums.ExamTypeTOEIC}
	require.NoError(t, tx.Create(cat).Error)
	raw, err := os.ReadFile("../../../modules/exams/toeic/tests/testdata/synthetic.json")
	require.NoError(t, err)
	test := &models.ExamTest{CategoryID: cat.ID, SourceKey: owner, Slug: owner, Title: "Statistics probe", Content: datatypes.JSON(raw)}
	require.NoError(t, tx.Create(test).Error)
	now := time.Now().UTC().Truncate(time.Second)
	start := now.Add(-24 * time.Hour)
	mode := enums.ExamModeExam
	ptr := func(v float64) *float64 { return &v }
	insert := func(user string, status enums.AttemptStatus, runMode enums.ExamMode, score *float64, version string, submitted time.Time, id string) {
		result, err := json.Marshal(map[string]any{"revision": 1, "answers": []any{}, "scoringVersion": version})
		require.NoError(t, err)
		a := &models.Attempt{ID: id, UserID: user, ExamTestID: &test.ID, Type: enums.AttemptTypeExam, Status: status, Mode: &runMode, SelectedParts: []int64{1, 2, 3, 4, 5, 6, 7}, Score: score, Result: datatypes.JSON(result), StartedAt: &start, CreatedAt: now}
		if status == enums.AttemptStatusCompleted {
			a.SubmittedAt = &submitted
		}
		require.NoError(t, tx.Create(a).Error)
	}
	insert(owner, enums.AttemptStatusCompleted, mode, ptr(615), toeic.ScoringVersion, now.Add(-time.Hour), uuid.NewString())
	// A late-created record with an older submission must not become latest.
	require.NoError(t, tx.Model(&models.Attempt{}).Where("user_id = ? AND score = ?", owner, 615).Update("created_at", now.Add(24*time.Hour)).Error)
	insert(owner, enums.AttemptStatusCompleted, mode, ptr(10), toeic.ScoringVersion, now, "00000000-0000-4000-8000-000000000001")
	insert(owner, enums.AttemptStatusCompleted, mode, ptr(990), toeic.ScoringVersion, now, "ffffffff-ffff-4fff-8fff-ffffffffffff")
	insert(owner, enums.AttemptStatusCompleted, mode, ptr(500), "other-policy", now.Add(time.Hour), uuid.NewString())
	for i := 0; i < 7; i++ {
		insert(owner, enums.AttemptStatusCompleted, mode, nil, "", now, uuid.NewString())
	}
	insert(owner, enums.AttemptStatusCompleted, enums.ExamModePractice, ptr(990), toeic.ScoringVersion, now, uuid.NewString())
	insert(owner, enums.AttemptStatusAbandoned, mode, ptr(990), toeic.ScoringVersion, now, uuid.NewString())
	insert(owner, enums.AttemptStatusInProgress, mode, nil, "", now, uuid.NewString())
	insert(owner+"-other", enums.AttemptStatusCompleted, mode, ptr(990), toeic.ScoringVersion, now, uuid.NewString())
	insert(owner+"-legacy", enums.AttemptStatusCompleted, mode, nil, "", now, uuid.NewString())
	repo := repositories.NewAttemptRepository(tx)
	tests := []struct {
		name, user string
		completed  int64
		scored     bool
	}{{"all pages and filters", owner, 11, true}, {"new user", owner + "-new", 0, false}, {"legacy only", owner + "-legacy", 1, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := repo.GetExamStats(ctx, tt.user)
			require.NoError(t, err)
			require.Equal(t, tt.completed, stats.CompletedTests)
			if tt.scored {
				require.Equal(t, 990.0, *stats.HighestScore)
				require.InDelta(t, 1615.0/3, *stats.AverageScore, 0.0001)
				require.Equal(t, 990.0, *stats.LatestScore)
			} else {
				require.Nil(t, stats.HighestScore)
				require.Nil(t, stats.AverageScore)
				require.Nil(t, stats.LatestScore)
			}
		})
	}
	query := ""
	for _, statement := range trace.statements {
		if strings.Contains(statement, "WITH completed AS") {
			query = statement
			break
		}
	}
	for _, fragment := range []string{"a.user_id =", "a.mode = 'exam'", "a.status = 'completed'", "c.type = 'toeic'", "score IS NOT NULL", "scoring_version = 'toeic-estimate-v1'", "submitted_at DESC, id DESC", "COUNT(*)"} {
		require.Contains(t, query, fragment)
	}
	// Verify actual finalization writes score and JSON in the same transaction.
	for _, correct := range []bool{false, true} {
		t.Run(fmt.Sprintf("atomic score %v", correct), func(t *testing.T) {
			svc := services.NewAttemptService(repo, repositories.NewExamTestRepository(tx), repositories.NewTransactionManager(tx))
			a, err := svc.Start(ctx, owner+"-grading", test.ID, requests.StartAttempt{Mode: mode, SelectedParts: []int{1, 2, 3, 4, 5, 6, 7}})
			require.NoError(t, err)
			answers := []requests.Answer{}
			if correct {
				for _, part := range a.ExamTest.ParsedContent.Parts {
					for _, item := range part.Items {
						for _, q := range item.Questions {
							answers = append(answers, requests.Answer{QuestionID: q.ID, SelectedKey: q.CorrectKey})
						}
					}
				}
			}
			graded, err := svc.Patch(ctx, owner+"-grading", a.ID, requests.PatchAttempt{Revision: 1, Answers: answers, Submit: true})
			require.NoError(t, err)
			stored, err := repo.GetExamAttempt(ctx, owner+"-grading", a.ID)
			require.NoError(t, err)
			want := 10.0
			if correct {
				want = 990
			}
			require.Equal(t, want, *stored.Score)
			require.JSONEq(t, string(graded.Result), string(stored.Result))
			parsed, err := toeic.ParseResult(stored.Result)
			require.NoError(t, err)
			require.Equal(t, want/2, *parsed.Sections.Listening.ScaledScore)
			retried, err := svc.Patch(ctx, owner+"-grading", a.ID, requests.PatchAttempt{Revision: 1, Submit: true})
			require.NoError(t, err)
			require.Equal(t, parsed.Revision, retried.ParsedResult.Revision)
		})
	}
}
