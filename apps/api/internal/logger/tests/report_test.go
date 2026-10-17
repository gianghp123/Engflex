package tests

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/logger"
)

// TestReport_LogsUnderlyingDbCause is the end-to-end guard for the Sentry
// issue: a hidden 500 (common.Internal()) must still name the database error
// it wrapped, because slog renders an error through Error(). Before the cause
// was preserved, the record said only "Internal Server Error" and the fault
// was undiagnosable.
func TestReport_LogsUnderlyingDbCause(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantLevel string
		wantCause string // substring the record must carry; "" = must not
	}{
		{
			"server fault logs cause at ERROR",
			common.FromDBError(errors.New("column part_of_speech does not exist"), "vocabulary item"),
			`"level":"ERROR"`,
			"part_of_speech",
		},
		{
			"client fault logs at WARN without internals",
			common.BadRequest("term is required"),
			`"level":"WARN"`,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(logger.NewWithWriter(&buf, "info", "json"))
			t.Cleanup(func() { slog.SetDefault(prev) })

			logger.Report(context.Background(), "list saved vocabulary failed", tt.err)
			out := buf.String()

			require.Contains(t, out, tt.wantLevel)
			if tt.wantCause == "" {
				require.NotContains(t, out, "part_of_speech")
				return
			}
			require.Contains(t, out, tt.wantCause)
		})
	}
}
