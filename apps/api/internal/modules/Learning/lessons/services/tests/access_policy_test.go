package tests

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"net/http"
	"testing"
)

type premiumState struct {
	active bool
	err    error
}

func (p premiumState) IsPremium(context.Context, string) (bool, error) { return p.active, p.err }

type lessonPolicyFunc func(context.Context, string, string) error

func (f lessonPolicyFunc) CheckAccess(ctx context.Context, userID, id string) error {
	return f(ctx, userID, id)
}
func TestLessonAccessMatrix(t *testing.T) {
	for _, tc := range []struct {
		name      string
		admin     bool
		published bool
		premium   bool
		active    bool
		outage    bool
		status    int
	}{
		{"published free", false, true, false, false, false, 0}, {"draft hidden", false, false, false, true, false, http.StatusNotFound}, {"draft admin preview", true, false, true, false, false, 0}, {"premium active", false, true, true, true, false, 0}, {"premium expired", false, true, true, false, false, http.StatusForbidden}, {"premium revoked", false, true, true, false, false, http.StatusForbidden}, {"billing unavailable", false, true, true, false, true, http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewMockLessonRepository(t)
			status := enums.PublishStatusDraft
			if tc.published {
				status = enums.PublishStatusPublished
			}
			repo.On("GetByID", mock.Anything, "lesson").Return(&models.Lesson{ID: "lesson", Status: status, IsPremium: tc.premium}, nil).Once()
			state := premiumState{active: tc.active}
			if tc.outage {
				state.err = common.Internal()
			}
			policy := services.NewLessonService(mocks.NewMockLessonSectionRepository(t), repo, state)
			ctx := common.WithActor(context.Background(), common.Actor{UserID: "learner", Admin: tc.admin})
			err := policy.CheckAccess(ctx, "learner", "lesson")
			if tc.status == 0 {
				require.NoError(t, err)
			} else {
				requireAppError(t, err, tc.status)
			}
		})
	}
}
func TestLessonActionAdmissionDeniesBeforeInference(t *testing.T) {
	for _, action := range []string{"list", "detail", "start", "resume", "check", "pronounce", "writing"} {
		t.Run(action, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			attempts := mocks.NewMockAttemptRepository(t)
			engine := voicemocks.NewMockPronounceEngine(t)
			llm := llmmocks.NewMockLLMClient(t)
			manager := mocks.NewMockTransactionManager(t)
			policy := lessonPolicyFunc(func(context.Context, string, string) error { return common.Forbidden("premium required") })
			svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(engine), llm, attempts, manager, policy)
			activity := attachCheckFixture()
			if action == "detail" || action == "check" || action == "pronounce" || action == "writing" {
				repo.On("GetByID", mock.Anything, activity.ID).Return(activity, nil).Once()
			}
			var err error
			switch action {
			case "list":
				_, err = svc.ListByLesson(context.Background(), attachLessonID)
			case "detail":
				_, err = svc.GetByID(context.Background(), activity.ID)
			case "start":
				_, err = svc.StartAttempt(context.Background(), attachUserID, attachLessonID)
			case "resume":
				_, err = svc.GetOpenProgress(context.Background(), attachUserID, attachLessonID)
			case "check":
				_, err = svc.CheckAnswer(context.Background(), attachUserID, activity.ID, attachOpenID, 0, "A")
			case "pronounce":
				_, err = svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, []byte("audio"), "audio/webm")
			case "writing":
				_, err = svc.ScoreWriting(context.Background(), attachUserID, activity.ID, attachOpenID, "draft")
			}
			requireAppError(t, err, http.StatusForbidden)
			engine.AssertNotCalled(t, "Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
func TestPremiumLostDuringSpeakingDoesNotPersist(t *testing.T) {
	repo := mocks.NewMockLessonActivityRepository(t)
	attempts := mocks.NewMockAttemptRepository(t)
	engine := voicemocks.NewMockPronounceEngine(t)
	manager := mocks.NewMockTransactionManager(t)
	calls := 0
	policy := lessonPolicyFunc(func(context.Context, string, string) error {
		calls++
		if calls > 1 {
			return common.Forbidden()
		}
		return nil
	})
	activity := pronounceSpeakingFixture()
	repo.On("GetByID", mock.Anything, activity.ID).Return(activity, nil).Once()
	attempts.On("GetForUpdate", mock.Anything, attachOpenID).Return(attachOpenAttempt(), nil).Twice()
	engine.On("Pronounce", mock.Anything, mock.Anything, "en", mock.Anything, "audio/webm").Return(&voice.PronounceResult{Score: 90}, nil).Once()
	manager.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(*gorm.DB) error) error { return fn(&gorm.DB{}) }).Once()
	repo.On("WithTx", mock.Anything).Return(repo).Once()
	attempts.On("WithTx", mock.Anything).Return(attempts).Once()
	svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(engine), llmmocks.NewMockLLMClient(t), attempts, manager, policy)
	_, err := svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, []byte("audio"), "audio/webm")
	requireAppError(t, err, http.StatusForbidden)
	attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
