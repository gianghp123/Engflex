// Paid exam AI contract (spec §8.3): cache read free, first call charges
// exactly once, LLM failure releases and allows retry, output and charge
// commit together, non-completed attempts are 422s, and shadow mode
// generates without metering.
package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	repomocks "engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	accessservices "engflex-api/internal/modules/access/services"
	accessmocks "engflex-api/internal/modules/access/services/mocks"
)

type fakeOutputs struct {
	existing *models.AIOutput
	created  []*models.AIOutput
	// missFirst models the generate race: the first cache read misses
	// (a concurrent request has not committed yet) and later reads hit.
	missFirst bool
	getCalls  int
}

func (f *fakeOutputs) GetBySubject(ctx context.Context, subjectType enums.UsageRefType, subjectID string, kind enums.AIOutputKind) (*models.AIOutput, error) {
	f.getCalls++
	if f.missFirst && f.getCalls == 1 {
		return nil, gorm.ErrRecordNotFound
	}
	if f.existing != nil {
		return f.existing, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeOutputs) Create(ctx context.Context, m *models.AIOutput) error {
	f.created = append(f.created, m)
	return nil
}

func (f *fakeOutputs) WithTx(tx *gorm.DB) repositories.AIOutputRepository { return f }

func completedAttempt() *models.Attempt {
	return &models.Attempt{
		ID:     "00000000-0000-0000-0000-000000000201",
		UserID: "user_1",
		Status: enums.AttemptStatusCompleted,
		Result: datatypes.JSON([]byte(`{}`)),
	}
}

func newAIR(t *testing.T, outputs *fakeOutputs, enforce bool) (*accessmocks.MockPolicyDecider, *accessmocks.MockUsageMeter, *accessservices.AIReviewService) {
	t.Helper()
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	tests := repomocks.NewMockExamTestRepository(t)
	svc := accessservices.NewAIReviewService(attempts, tests, outputs, llmmocks.NewMockLLMClient(t), policy, usage, enforce, fakeTxManager{})
	return policy, usage, svc
}

func TestGetExamReviewCachedReadIsFree(t *testing.T) {
	cached := &models.AIOutput{UserID: "user_1", Kind: enums.AIOutputKindExamReview, Payload: datatypes.JSON([]byte(`{"summary":"ok"}`))}
	policy, usage, svc := newAIR(t, &fakeOutputs{existing: cached}, true)

	got, err := svc.GetExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, cached, got)
	policy.AssertNotCalled(t, "Decide", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	usage.AssertNotCalled(t, "Reserve", mock.Anything)
}

func TestGetStudyPlanCachedReadIsFree(t *testing.T) {
	cached := &models.AIOutput{UserID: "user_1", Kind: enums.AIOutputKindStudyPlan, Payload: datatypes.JSON([]byte(`{"summary":"plan"}`))}
	policy, usage, svc := newAIR(t, &fakeOutputs{existing: cached}, true)

	got, err := svc.GetStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, cached, got)
	policy.AssertNotCalled(t, "Decide", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	usage.AssertNotCalled(t, "Reserve", mock.Anything)
}

func TestGenerateExamReviewChargesOnceAndCaches(t *testing.T) {
	outputs := &fakeOutputs{}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	testsRepo := repomocks.NewMockExamTestRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKey1M}, accessservices.TrialInfo{}, nil).Once()
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionExamReview).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKey1M,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	policy.EXPECT().Location().Return(time.UTC).Maybe()
	usage.EXPECT().Reserve(mock.Anything, "user_1", enums.PlanKey1M, mock.Anything,
		"exam_review:00000000-0000-0000-0000-000000000201", enums.UsageRefTypeExamAttempt,
		"00000000-0000-0000-0000-000000000201", common.ReviewLeaseTTL, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-r1", Status: accessservices.ReservationStatusReserved}, nil).Once()
	// The attempt's exam content load for prompt building:
	testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
	llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	usage.EXPECT().CommitTx(mock.Anything, mock.Anything, "res-r1").Return(nil).Once()

	svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, true, fakeTxManager{})
	out, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, outputs.created, 1)
	assert.Equal(t, 1, outputs.created[0].CreditsCharged)
	assert.Equal(t, enums.AIOutputKindExamReview, outputs.created[0].Kind)
}

func TestGenerateStudyPlanChargesTwoCredits(t *testing.T) {
	outputs := &fakeOutputs{}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	testsRepo := repomocks.NewMockExamTestRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKey1M}, accessservices.TrialInfo{}, nil).Once()
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionStudyPlan).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKey1M,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 2}}}, nil).Once()
	policy.EXPECT().Location().Return(time.UTC).Maybe()
	usage.EXPECT().Reserve(mock.Anything, "user_1", enums.PlanKey1M,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 2}},
		"study_plan:00000000-0000-0000-0000-000000000201", enums.UsageRefTypeStudyPlan,
		"00000000-0000-0000-0000-000000000201", common.ReviewLeaseTTL, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-p1", Status: accessservices.ReservationStatusReserved}, nil).Once()
	testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
	llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	usage.EXPECT().CommitTx(mock.Anything, mock.Anything, "res-p1").Return(nil).Once()

	svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, true, fakeTxManager{})
	out, err := svc.GenerateStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	require.NoError(t, err)
	require.NotNil(t, out)
	require.Len(t, outputs.created, 1)
	assert.Equal(t, 2, outputs.created[0].CreditsCharged, "study plan charges exactly 2 credits")
	assert.Equal(t, enums.AIOutputKindStudyPlan, outputs.created[0].Kind)
}

func TestGenerateExamReviewLLMFailureReleases(t *testing.T) {
	outputs := &fakeOutputs{}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	testsRepo := repomocks.NewMockExamTestRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKey1M}, accessservices.TrialInfo{}, nil).Once()
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionExamReview).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKey1M,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	policy.EXPECT().Location().Return(time.UTC).Maybe()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-r2", Status: accessservices.ReservationStatusReserved}, nil).Once()
	testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
	llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(gorm.ErrInvalidDB).Once()
	usage.EXPECT().Release(mock.Anything, "res-r2").Return(nil).Once()

	svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, true, fakeTxManager{})
	_, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status)
	require.Equal(t, common.CodeAIUnavailable, appErr.Code)
	assert.Empty(t, outputs.created, "no output without a charge, no charge without an output")
}

func TestGenerateRequiresCompletedAttempt(t *testing.T) {
	tests := []struct {
		name   string
		action func(svc *accessservices.AIReviewService) error
	}{
		{"review", func(svc *accessservices.AIReviewService) error {
			_, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			return err
		}},
		{"study plan", func(svc *accessservices.AIReviewService) error {
			_, err := svc.GenerateStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inProgress := completedAttempt()
			inProgress.Status = enums.AttemptStatusInProgress
			attempts := repomocks.NewMockAttemptRepository(t)
			attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(inProgress, nil).Once()
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)
			svc := accessservices.NewAIReviewService(attempts, repomocks.NewMockExamTestRepository(t), &fakeOutputs{}, llmmocks.NewMockLLMClient(t), policy, usage, true, fakeTxManager{})

			err := tt.action(svc)
			appErr := &common.AppError{}
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, 422, appErr.Status, "an unsubmitted attempt has nothing to review")
		})
	}
}

func TestGenerateExamReviewReturnsExistingCacheWithoutMetering(t *testing.T) {
	cached := &models.AIOutput{UserID: "user_1", Kind: enums.AIOutputKindExamReview, Payload: datatypes.JSON([]byte(`{"summary":"ok"}`))}
	outputs := &fakeOutputs{existing: cached}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()

	svc := accessservices.NewAIReviewService(attempts, repomocks.NewMockExamTestRepository(t), outputs, llm, policy, usage, true, fakeTxManager{})
	got, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	require.NoError(t, err)
	assert.Equal(t, cached, got, "re-POST on a generated review replays free without metering or a second LLM call")
	assert.Empty(t, outputs.created)
	policy.AssertNotCalled(t, "Decide", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	usage.AssertNotCalled(t, "Reserve", mock.Anything)
	llm.AssertNotCalled(t, "CompleteStructured", mock.Anything)
}

func TestGenerateExamReviewCommittedReplayReturnsCache(t *testing.T) {
	cached := &models.AIOutput{UserID: "user_1", Kind: enums.AIOutputKindExamReview, Payload: datatypes.JSON([]byte(`{}`))}
	// The cache lands between our first read and our Reserve (a concurrent
	// generate committed first): the in_progress replay re-reads it.
	outputs := &fakeOutputs{existing: cached, missFirst: true}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKey1M}, accessservices.TrialInfo{}, nil).Once()
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionExamReview).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKey1M,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	policy.EXPECT().Location().Return(time.UTC).Maybe()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, common.Conflict("already committed").WithCode("in_progress"))

	svc := accessservices.NewAIReviewService(attempts, repomocks.NewMockExamTestRepository(t), outputs, llm, policy, usage, true, fakeTxManager{})
	out, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	require.NoError(t, err)
	assert.Equal(t, cached, out, "a committed reservation replays the cache without a second charge or LLM call")
	llm.AssertNotCalled(t, "CompleteStructured", mock.Anything)
}

func TestGenerateShadowModeSkipsMetering(t *testing.T) {
	tests := []struct {
		name     string
		action   func(svc *accessservices.AIReviewService) (*models.AIOutput, error)
		decision accessservices.Action
	}{
		{"review", func(svc *accessservices.AIReviewService) (*models.AIOutput, error) {
			return svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
		}, accessservices.ActionExamReview},
		{"study plan", func(svc *accessservices.AIReviewService) (*models.AIOutput, error) {
			return svc.GenerateStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
		}, accessservices.ActionStudyPlan},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &fakeOutputs{}
			llm := llmmocks.NewMockLLMClient(t)
			attempts := repomocks.NewMockAttemptRepository(t)
			testsRepo := repomocks.NewMockExamTestRepository(t)
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)

			attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
			policy.EXPECT().Principal(mock.Anything, "user_1", false).
				Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyFree}, accessservices.TrialInfo{}, nil).Once()
			policy.EXPECT().Decide(mock.Anything, "user_1", false, tt.decision).
				Return(accessservices.Decision{Allow: false, Code: common.CodeInsufficientCredits, Plan: enums.PlanKeyFree}, nil).Once()
			testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
			llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, false, fakeTxManager{})
			out, err := tt.action(svc)
			require.NoError(t, err, "shadow mode generates despite the denial")
			require.NotNil(t, out)
			require.Len(t, outputs.created, 1)
			usage.AssertNotCalled(t, "Reserve", mock.Anything)
			usage.AssertNotCalled(t, "CommitTx", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestGenerateDeniedWhenEnforced(t *testing.T) {
	tests := []struct {
		name     string
		action   func(svc *accessservices.AIReviewService) error
		decision accessservices.Action
	}{
		{"review", func(svc *accessservices.AIReviewService) error {
			_, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			return err
		}, accessservices.ActionExamReview},
		{"study plan", func(svc *accessservices.AIReviewService) error {
			_, err := svc.GenerateStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			return err
		}, accessservices.ActionStudyPlan},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)
			policy.EXPECT().Principal(mock.Anything, "user_1", false).
				Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyFree}, accessservices.TrialInfo{}, nil).Once()
			policy.EXPECT().Decide(mock.Anything, "user_1", false, tt.decision).
				Return(accessservices.Decision{Allow: false, Code: common.CodeInsufficientCredits, Plan: enums.PlanKeyFree}, nil).Once()

			svc := accessservices.NewAIReviewService(attempts, repomocks.NewMockExamTestRepository(t), &fakeOutputs{}, llmmocks.NewMockLLMClient(t), policy, usage, true, fakeTxManager{})
			err := tt.action(svc)
			appErr := &common.AppError{}
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, 429, appErr.Status)
			require.Equal(t, common.CodeInsufficientCredits, appErr.Code)
			usage.AssertNotCalled(t, "Reserve", mock.Anything)
		})
	}
}

// Admin bypass on paid exam AI (spec §5.5): after Principal reports an admin,
// a denial-shaped Decide is ignored and no hold is ever reserved or
// committed — the LLM still runs and the output is stored unmetered. A free
// user with the same denial keeps the 429.
func TestGenerateAdminBypassSkipsMetering(t *testing.T) {
	tests := []struct {
		name       string
		action     func(svc *accessservices.AIReviewService) (*models.AIOutput, error)
		decision   accessservices.Action
		charges    []accessservices.Charge
		isAdmin    bool
		principal  accessservices.Principal
		wantStatus int
	}{
		{
			name: "admin review succeeds despite a denial-shaped decide",
			action: func(svc *accessservices.AIReviewService) (*models.AIOutput, error) {
				return svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", true)
			}, decision: accessservices.ActionExamReview,
			charges:   []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
			isAdmin:   true,
			principal: accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyAdmin, IsAdmin: true},
		},
		{
			name: "admin study plan succeeds despite a denial-shaped decide",
			action: func(svc *accessservices.AIReviewService) (*models.AIOutput, error) {
				return svc.GenerateStudyPlan(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", true)
			}, decision: accessservices.ActionStudyPlan,
			charges:   []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 2}},
			isAdmin:   true,
			principal: accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyAdmin, IsAdmin: true},
		},
		{
			name: "free user with the same denial is still denied",
			action: func(svc *accessservices.AIReviewService) (*models.AIOutput, error) {
				return svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			}, decision: accessservices.ActionExamReview,
			charges:    []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
			principal:  accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyFree},
			wantStatus: 429,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &fakeOutputs{}
			llm := llmmocks.NewMockLLMClient(t)
			attempts := repomocks.NewMockAttemptRepository(t)
			testsRepo := repomocks.NewMockExamTestRepository(t)
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)

			attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
			policy.EXPECT().Principal(mock.Anything, "user_1", tt.isAdmin).
				Return(tt.principal, accessservices.TrialInfo{}, nil).Once()
			policy.EXPECT().Decide(mock.Anything, "user_1", tt.isAdmin, tt.decision).
				Return(accessservices.Decision{Allow: false, Code: common.CodeInsufficientCredits, Plan: tt.principal.Plan,
					Charges: tt.charges}, nil).Once()
			if tt.wantStatus == 0 {
				testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
				llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			}

			svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, true, fakeTxManager{})
			out, err := tt.action(svc)
			if tt.wantStatus != 0 {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantStatus, appErr.Status)
				require.Nil(t, out)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, out)
			usage.AssertNotCalled(t, "Reserve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			usage.AssertNotCalled(t, "CommitTx", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// Shadow mode never blocks: a Principal or Decide store error is logged and
// the review generates unmetered instead of 503ing.
func TestGenerateShadowModeToleratesStoreErrors(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*accessmocks.MockPolicyDecider)
	}{
		{
			name: "principal store error proceeds unmetered",
			setup: func(policy *accessmocks.MockPolicyDecider) {
				policy.EXPECT().Principal(mock.Anything, "user_1", false).
					Return(accessservices.Principal{}, accessservices.TrialInfo{}, errors.New("billing down")).Once()
			},
		},
		{
			name: "decide store error proceeds unmetered",
			setup: func(policy *accessmocks.MockPolicyDecider) {
				policy.EXPECT().Principal(mock.Anything, "user_1", false).
					Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyFree}, accessservices.TrialInfo{}, nil).Once()
				policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionExamReview).
					Return(accessservices.Decision{}, errors.New("usage store down")).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputs := &fakeOutputs{}
			llm := llmmocks.NewMockLLMClient(t)
			attempts := repomocks.NewMockAttemptRepository(t)
			testsRepo := repomocks.NewMockExamTestRepository(t)
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)

			attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
			tt.setup(policy)
			testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
			llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, false, fakeTxManager{})
			out, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
			require.NoError(t, err, "shadow mode never fails on a store error")
			require.NotNil(t, out)
			usage.AssertNotCalled(t, "Reserve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// Settlement and output commit atomically: when the tx covering the
// ai_outputs insert and the charge fails, the hold is released and the
// request fails closed (503) — no charge without an output.
func TestGenerateSettleFailureRollsBackAndReleases(t *testing.T) {
	outputs := &fakeOutputs{}
	llm := llmmocks.NewMockLLMClient(t)
	attempts := repomocks.NewMockAttemptRepository(t)
	testsRepo := repomocks.NewMockExamTestRepository(t)
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)

	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000201").Return(completedAttempt(), nil).Once()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKey1M}, accessservices.TrialInfo{}, nil).Once()
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionExamReview).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKey1M,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	policy.EXPECT().Location().Return(time.UTC).Maybe()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-fail", Status: accessservices.ReservationStatusReserved}, nil).Once()
	testsRepo.EXPECT().GetByID(mock.Anything, mock.Anything).Return(&models.ExamTestCatalog{}, nil).Maybe()
	llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	usage.EXPECT().CommitTx(mock.Anything, mock.Anything, "res-fail").Return(errors.New("settle down")).Once()
	usage.EXPECT().Release(mock.Anything, "res-fail").Return(nil).Once()

	svc := accessservices.NewAIReviewService(attempts, testsRepo, outputs, llm, policy, usage, true, fakeTxManager{})
	_, err := svc.GenerateExamReview(context.Background(), "user_1", "00000000-0000-0000-0000-000000000201", false)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status)
	require.Equal(t, common.CodePolicyUnavailable, appErr.Code)
}
