package tests

import (
	"context"
	"testing"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/lessons/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePremiumChecker defaults to "not premium" — most tests exercise
// free-content behavior and don't care about the subscription lookup.
type fakePremiumChecker struct {
	isPremium bool
}

func (f fakePremiumChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	return f.isPremium, nil
}

func setupLesson(t *testing.T) (services.LessonService, *mocks.MockLessonSectionRepository, *mocks.MockLessonRepository) {
	t.Helper()
	return setupLessonWithPremium(t, false)
}

func setupLessonWithPremium(t *testing.T, isPremium bool) (services.LessonService, *mocks.MockLessonSectionRepository, *mocks.MockLessonRepository) {
	t.Helper()
	secMock := mocks.NewMockLessonSectionRepository(t)
	lessonMock := mocks.NewMockLessonRepository(t)
	svc := services.NewLessonService(secMock, lessonMock, fakePremiumChecker{isPremium: isPremium})
	return svc, secMock, lessonMock
}

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}
