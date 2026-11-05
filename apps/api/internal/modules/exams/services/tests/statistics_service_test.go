package tests

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/exams/services"
)

func TestExamStatisticsErrors(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"database unavailable", errors.New("database unavailable"), http.StatusInternalServerError},
		{"database permission failure", gorm.ErrInvalidDB, http.StatusInternalServerError},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			attempts.On("GetExamStatistics", mock.Anything, "owner").Return(nil, tt.err)
			svc := services.NewAttemptService(attempts, repomocks.NewMockExamTestRepository(t), repomocks.NewMockTransactionManager(t))
			stats, err := svc.Statistics(t.Context(), "owner")
			require.Nil(t, stats)
			var appErr *common.AppError
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, tt.status, appErr.Status)
		})
	}
}
