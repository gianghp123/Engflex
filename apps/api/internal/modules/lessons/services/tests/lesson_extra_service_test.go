package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func TestLessonService_ListSections_RepoError(t *testing.T) {
	svc, secMock, _ := setupLesson(t)
	secMock.On("ListWithUnits", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError)

	res, err := svc.ListSections(context.Background(), requests.ListLessons{}, false)

	requireAppError(t, err, http.StatusInternalServerError)
	assert.Nil(t, res)
}

func TestLessonActivityService_ListByLesson_EmptyReturns200(t *testing.T) {
	lessonID := "11111111-1111-1111-1111-111111111111"
	repo := mocks.NewMockLessonActivityRepository(t)
	repo.On("ListByLesson", mock.Anything, lessonID).
		Return(make([]*models.LessonActivity, 0), nil)
	svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), mocks.NewMockAttemptRepository(t))

	res, err := svc.ListByLesson(context.Background(), lessonID)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Empty(t, res)
}

func TestLessonBookmarkService_Save_Idempotent(t *testing.T) {
	userID := "user_1"
	lessonID := "11111111-1111-1111-1111-111111111111"
	row := &models.LessonBookmark{UserID: userID, LessonID: lessonID}
	repo := mocks.NewMockLessonBookmarkRepository(t)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)
	repo.On("Get", mock.Anything, userID, lessonID).Return(row, nil)
	svc := services.NewLessonBookmarkService(repo)

	for i := 0; i < 2; i++ {
		got, err := svc.Save(context.Background(), userID, lessonID)
		require.NoError(t, err)
		assert.Equal(t, lessonID, got.LessonID)
	}
}
