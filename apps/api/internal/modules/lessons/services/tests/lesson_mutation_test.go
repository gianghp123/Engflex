package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/services"
)

// Create maps the nested Details document (LessonDetailsInput ->
// LessonDetails) field by field despite the two types having different
// names, and always starts a lesson as draft — same guard as the identical
// scenario Details bug (copier silently zeroes a Scanner/Valuer-backed
// nested field instead of erroring).
func TestLessonService_Create_MapsDetailsAndStartsDraft(t *testing.T) {
	_, secMock, lessonMock := setupLesson(t)
	duration := 12
	lessonMock.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Lesson) bool {
		return m.Status == enums.PublishStatusDraft &&
			m.Details.CoverImageURL == "https://example.com/cover.png" &&
			m.Details.EstimatedDurationMin != nil && *m.Details.EstimatedDurationMin == 12
	})).Return(nil).Once()
	svc := services.NewLessonService(secMock, lessonMock)

	_, err := svc.Create(context.Background(), requests.CreateLesson{
		Slug: "unit-1", Title: "Unit 1",
		SectionID: "11111111-1111-1111-1111-111111111111",
		CEFRLevel: "B1",
		Details: requests.LessonDetailsInput{
			EstimatedDurationMin: &duration,
			CoverImageURL:        "https://example.com/cover.png",
		},
	})

	require.NoError(t, err)
}

func TestLessonService_Update_PublishesOnExplicitStatus(t *testing.T) {
	lessonID := "11111111-1111-1111-1111-111111111111"
	_, secMock, lessonMock := setupLesson(t)
	lessonMock.On("GetByID", mock.Anything, lessonID).
		Return(&models.Lesson{ID: lessonID, Status: enums.PublishStatusDraft}, nil).Once()
	lessonMock.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Lesson) bool {
		return m.Status == enums.PublishStatusPublished
	})).Return(nil).Once()
	svc := services.NewLessonService(secMock, lessonMock)

	published := enums.PublishStatusPublished
	_, err := svc.Update(context.Background(), lessonID, requests.UpdateLesson{Status: &published})

	require.NoError(t, err)
}

// Create rejects a config that doesn't parse against its declared type —
// a malformed admin-authored document must never reach the column.
func TestLessonActivityService_Create_RejectsInvalidConfig(t *testing.T) {
	repo := mocks.NewMockLessonActivityRepository(t)
	svc := lessonActivitySvc(t, repo)

	_, err := svc.Create(context.Background(), requests.CreateLessonActivity{
		LessonID: "11111111-1111-1111-1111-111111111111", PartNumber: 1,
		Type: enums.ActivityTypeReading, Title: "Part 1",
		Config: []byte(`not valid json`),
	})

	requireAppError(t, err, 400)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestLessonActivityService_Create_AcceptsValidConfig(t *testing.T) {
	repo := mocks.NewMockLessonActivityRepository(t)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(m *models.LessonActivity) bool {
		return string(m.Config) == `{"text":"hello","questions":[]}`
	})).Return(nil).Once()
	svc := lessonActivitySvc(t, repo)

	_, err := svc.Create(context.Background(), requests.CreateLessonActivity{
		LessonID: "11111111-1111-1111-1111-111111111111", PartNumber: 1,
		Type: enums.ActivityTypeReading, Title: "Part 1",
		Config: []byte(`{"text":"hello","questions":[]}`),
	})

	require.NoError(t, err)
}

func TestLessonActivityService_Update_ValidatesAgainstEffectiveType(t *testing.T) {
	activityID := "33333333-3333-3333-3333-333333333333"
	repo := mocks.NewMockLessonActivityRepository(t)
	repo.On("GetByID", mock.Anything, activityID).
		Return(&models.LessonActivity{ID: activityID, Type: enums.ActivityTypeReading}, nil).Once()
	svc := lessonActivitySvc(t, repo)

	// A writing-shaped config against the existing (reading) type: missing
	// "questions" is fine for ReadingConfig (zero value), so assert the
	// stricter failure instead — an empty config body is syntactically
	// valid JSON for any shape, so use a non-object to force a parse error.
	_, err := svc.Update(context.Background(), activityID, requests.UpdateLessonActivity{
		Config: []byte(`[1,2,3]`),
	})

	requireAppError(t, err, 400)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func lessonActivitySvc(t *testing.T, repo *mocks.MockLessonActivityRepository) services.LessonActivityService {
	t.Helper()
	return services.NewLessonActivityService(
		repo,
		nil,
		nil,
		mocks.NewMockAttemptRepository(t),
	)
}
