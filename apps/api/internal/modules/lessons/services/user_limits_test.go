package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The fixed-window limiter with explicit timestamps: the window slides,
// stale hits never accumulate, and idle users leave no empty slice behind.
func TestUserRateLimiterWindow(t *testing.T) {
	window := time.Minute
	base := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		limit int
	}{
		{name: "production budget of sixty", limit: 60},
		{name: "small budget of three", limit: 3},
		{name: "budget of one", limit: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewUserRateLimiterForTest(tt.limit, window)
			for i := 0; i < tt.limit; i++ {
				require.True(t, l.allow("user_1", base), "hit %d of %d is allowed", i+1, tt.limit)
			}
			require.False(t, l.allow("user_1", base), "one over budget is denied inside the window")

			// Past the window every hit is allowed again and only the fresh
			// hit is retained: stale entries are evicted, not accumulated.
			later := base.Add(window + time.Second)
			require.True(t, l.allow("user_1", later), "after the window the budget resets")
			l.mu.Lock()
			kept := len(l.hits["user_1"])
			l.mu.Unlock()
			assert.Equal(t, 1, kept, "only the fresh hit is retained after the window slides")
		})
	}
}

// An empty window leaves no entry: the key is removed instead of retaining
// an empty slice for an idle user.
func TestUserRateLimiterRemovesEmptyKeys(t *testing.T) {
	l := NewUserRateLimiterForTest(0, time.Minute)
	require.False(t, l.allow("user_1", time.Now()), "a zero budget denies everything")
	l.mu.Lock()
	_, ok := l.hits["user_1"]
	l.mu.Unlock()
	assert.False(t, ok, "a denied empty window leaves no map entry behind")
}
