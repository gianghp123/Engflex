package tests

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
)

// TestFromDBError_PreservesCauseOnServerFault pins the diagnosis contract: a
// server fault (the 500 fallthrough) keeps the underlying database error so
// logs and Sentry can name it, while every client-facing branch keeps Message
// alone so internals never reach the wire.
func TestFromDBError_PreservesCauseOnServerFault(t *testing.T) {
	dbErr := errors.New("ERROR: column vocabulary_items.part_of_speech does not exist (SQLSTATE 42703)")

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
		wantCause  string // substring Error() must carry; "" = must equal wantMsg
	}{
		{"unknown db error -> 500 keeps cause", dbErr, http.StatusInternalServerError, "Internal Server Error", "part_of_speech"},
		{"record not found -> 404, no cause", gorm.ErrRecordNotFound, http.StatusNotFound, "vocabulary item not found", ""},
		{"unique violation -> 409, no cause", &pgconn.PgError{Code: "23505"}, http.StatusConflict, "vocabulary item already exists", ""},
		{"invalid id -> 400, no cause", &common.InvalidIDError{Name: "id"}, http.StatusBadRequest, "invalid id", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := common.FromDBError(tt.err, "vocabulary item")
			require.NotNil(t, appErr)
			assert.Equal(t, tt.wantStatus, appErr.Status)
			assert.Equal(t, tt.wantMsg, appErr.Message)
			if tt.wantCause == "" {
				assert.Equal(t, tt.wantMsg, appErr.Error())
			} else {
				assert.Contains(t, appErr.Error(), tt.wantCause)
			}
		})
	}
}

// TestFromDBError_CauseIsUnwrappable keeps the cause reachable through the
// standard error chain without disturbing errors.As on the AppError itself.
func TestFromDBError_CauseIsUnwrappable(t *testing.T) {
	dbErr := errors.New("boom")
	appErr := common.FromDBError(dbErr, "vocabulary item")

	require.ErrorIs(t, appErr, dbErr)

	var target *common.AppError
	require.ErrorAs(t, appErr, &target)
	assert.Equal(t, http.StatusInternalServerError, target.Status)
}

// TestAppError_NoCauseKeepsMessageAlone guards the constructors that attach no
// cause: Error() stays exactly Message, so nothing changes for them.
func TestAppError_NoCauseKeepsMessageAlone(t *testing.T) {
	assert.Equal(t, "Internal Server Error", common.Internal().Error())
	assert.Equal(t, "term is required", common.BadRequest("term is required").Error())
}
