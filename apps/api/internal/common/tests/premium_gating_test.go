// Black-boxes the gating switch: Option A disables per-item premium checks
// process-wide, and the checker must never even be called while off — the
// flag also removes the subscription lookup's failure mode from content reads.
package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
)

type countingChecker struct{ calls int }

func (c *countingChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	c.calls++
	return false, nil
}

type failingChecker struct{}

func (failingChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	return false, errors.New("subscription store down")
}

func TestRequirePremiumAccessGatingSwitch(t *testing.T) {
	tests := []struct {
		name        string
		gatingOn    bool
		isAdmin     bool
		contentPaid bool
		checker     common.PremiumChecker
		wantErrCode int // 0 = nil
		wantCalls   int
	}{
		{name: "gating off never calls the checker", gatingOn: false, contentPaid: true, checker: &countingChecker{}, wantCalls: 0},
		{name: "gating on blocks a free user", gatingOn: true, contentPaid: true, checker: &countingChecker{}, wantErrCode: 403, wantCalls: 1},
		{name: "gating on passes an admin without lookup", gatingOn: true, isAdmin: true, contentPaid: true, checker: &countingChecker{}, wantCalls: 0},
		{name: "non premium content is always free", gatingOn: true, contentPaid: false, checker: &countingChecker{}, wantCalls: 0},
		{name: "gating on surfaces checker failures", gatingOn: true, contentPaid: true, checker: failingChecker{}, wantErrCode: 500, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			err := common.RequirePremiumAccess(context.Background(), tt.checker, "user_1", tt.isAdmin, tt.contentPaid)

			if tt.wantErrCode == 0 {
				require.NoError(t, err)
			} else {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantErrCode, appErr.Status)
			}
			if cc, ok := tt.checker.(*countingChecker); ok {
				require.Equal(t, tt.wantCalls, cc.calls)
			}
		})
	}
}
