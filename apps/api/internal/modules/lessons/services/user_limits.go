package services

import (
	"sync"
	"time"
)

// writingRequestsPerMinute caps free writing scores per user (spec §8.2).
// Completion-critical AI is never quota-gated; this is an abuse/cost guard.
const writingRequestsPerMinute = 60

// userRateLimiter is a fixed-window per-user counter. In-process by design:
// the guard is cheap, best-effort protection for free AI, and the Postgres
// counters (access module) own everything that money depends on.
type userRateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string][]time.Time
}

func newUserRateLimiter(limit int, window time.Duration) *userRateLimiter {
	return &userRateLimiter{limit: limit, window: window, hits: map[string][]time.Time{}}
}

func (l *userRateLimiter) allow(userID string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Build a fresh slice: reusing l.hits[userID][:0] aliases the stored
	// backing array, so the filter's appends can clobber not-yet-visited
	// entries and the map retains memory for idle users.
	fresh := make([]time.Time, 0, len(l.hits[userID]))
	for _, at := range l.hits[userID] {
		if now.Sub(at) < l.window {
			fresh = append(fresh, at)
		}
	}
	if len(fresh) == 0 {
		delete(l.hits, userID)
	} else {
		l.hits[userID] = fresh
	}
	if len(fresh) >= l.limit {
		return false
	}
	l.hits[userID] = append(fresh, now)
	return true
}

// NewUserRateLimiterForTest exposes the limiter for white-box rate tests
// with explicit timestamps. Never pass nil; the zero limit denies all.
func NewUserRateLimiterForTest(limit int, window time.Duration) *userRateLimiter {
	return newUserRateLimiter(limit, window)
}

// inflightGuard allows at most max concurrent entries per key. Pronunciation
// scores one upload at a time per user (spec §8.2).
type inflightGuard struct {
	mu  sync.Mutex
	max int
	now map[string]int
}

func newInflightGuard(max int) *inflightGuard {
	return &inflightGuard{max: max, now: map[string]int{}}
}

func (g *inflightGuard) acquire(userID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.now[userID] >= g.max {
		return false
	}
	g.now[userID]++
	return true
}

func (g *inflightGuard) release(userID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.now[userID] > 0 {
		g.now[userID]--
	}
}
