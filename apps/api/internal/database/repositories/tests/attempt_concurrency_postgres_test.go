package tests

import (
	"context"
	"net/http"
	"os"
	"sync"
	"testing"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	llmmocks "engflex-api/internal/llm/mocks"
	lessons "engflex-api/internal/modules/lessons/services"
	videoRequests "engflex-api/internal/modules/videos/dtos/requests"
	videos "engflex-api/internal/modules/videos/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

func TestPostgresAttemptConcurrency(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	ctx := context.Background()
	userID := "concurrency-" + uuid.NewString()
	section := &models.LessonSection{Slug: uuid.NewString(), Title: "Concurrency", CEFRBand: enums.CEFR("A1")}
	require.NoError(t, db.Create(section).Error)
	lesson := &models.Lesson{Slug: uuid.NewString(), Title: "Concurrency", SectionID: section.ID, CEFRLevel: enums.CEFR("A1")}
	require.NoError(t, db.Create(lesson).Error)
	activity := &models.LessonActivity{LessonID: lesson.ID, Type: enums.ActivityTypeReading, Config: []byte(`{"text":"test","questions":[{"stem":"one","instruction":"pick","options":[],"answerKey":"A","explanation":"one"},{"stem":"two","instruction":"pick","options":[],"answerKey":"B","explanation":"two"}]}`)}
	require.NoError(t, db.Create(activity).Error)
	video := &models.VideoExercise{Title: "Concurrency", VideoURL: "test"}
	require.NoError(t, db.Create(video).Error)
	t.Cleanup(func() {
		require.NoError(t, db.Where("user_id = ?", userID).Delete(&models.Attempt{}).Error)
		require.NoError(t, db.Delete(activity).Error)
		require.NoError(t, db.Delete(lesson).Error)
		require.NoError(t, db.Delete(section).Error)
		require.NoError(t, db.Delete(video).Error)
	})
	repo := repositories.NewAttemptRepository(db)
	svc := lessons.NewLessonActivityService(repositories.NewLessonActivityRepository(db), pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), repo, repositories.NewTransactionManager(db), allowLearningAccess{})
	const workers = 8
	ids := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			m, e := svc.StartAttempt(ctx, userID, lesson.ID)
			errs <- e
			if e == nil {
				ids <- m.ID
			}
		})
	}
	wg.Wait()
	close(ids)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	var attemptID string
	for id := range ids {
		if attemptID == "" {
			attemptID = id
		}
		require.Equal(t, attemptID, id)
	}
	_, err = svc.CheckAnswer(ctx, userID, activity.ID, attemptID, 0, "A")
	require.NoError(t, err)
	first, err := repo.GetByID(ctx, attemptID)
	require.NoError(t, err)
	require.Equal(t, enums.AttemptStatusInProgress, first.Status)
	// Concurrent answers must merge the latest locked row, not overwrite each other.
	answers := make(chan error, 2)
	for index, key := range []string{"A", "B"} {
		wg.Go(func() { _, e := svc.CheckAnswer(ctx, userID, activity.ID, attemptID, index, key); answers <- e })
	}
	wg.Wait()
	close(answers)
	// The first answer can race the completing answer and see a closed attempt;
	// neither may undo completion or erase an already acknowledged answer.
	for e := range answers {
		if e != nil {
			var app *common.AppError
			require.ErrorAs(t, e, &app)
			require.Equal(t, http.StatusConflict, app.Status)
		}
	}
	completed, err := repo.GetByID(ctx, attemptID)
	require.NoError(t, err)
	require.Equal(t, enums.AttemptStatusCompleted, completed.Status)
	require.NotNil(t, completed.Score)
	require.Equal(t, 100.0, *completed.Score)
	require.NotNil(t, completed.StartedAt)
	require.NotNil(t, completed.SubmittedAt)
	vsvc := videos.NewVideoAttemptService(repo, repositories.NewVideoExerciseRepository(db), allowLearningAccess{})
	open, err := vsvc.StartAttempt(ctx, userID, video.ID, enums.AttemptTypeDictation, videoRequests.StartVideoAttempt{})
	require.NoError(t, err)
	saved, err := vsvc.SaveProgress(ctx, userID, video.ID, open.ID, videoRequests.SaveVideoProgress{Revision: open.Revision})
	require.NoError(t, err)
	require.Equal(t, open.Revision+1, saved.Revision)
	_, err = vsvc.SaveProgress(ctx, userID, video.ID, open.ID, videoRequests.SaveVideoProgress{Revision: open.Revision})
	var app *common.AppError
	require.ErrorAs(t, err, &app)
	require.Equal(t, http.StatusConflict, app.Status)
	score := 90.0
	req := videoRequests.CompleteVideoAttempt{Score: &score}
	_, err = vsvc.CompleteAttempt(ctx, userID, video.ID, open.ID, req)
	require.NoError(t, err)
	_, err = vsvc.CompleteAttempt(ctx, userID, video.ID, open.ID, req)
	require.NoError(t, err)
	_, err = vsvc.SaveProgress(ctx, userID, video.ID, open.ID, videoRequests.SaveVideoProgress{Revision: saved.Revision})
	require.ErrorAs(t, err, &app)
	require.Equal(t, http.StatusConflict, app.Status)
}

type allowLearningAccess struct{}

func (allowLearningAccess) CheckAccess(context.Context, string, string) error { return nil }
