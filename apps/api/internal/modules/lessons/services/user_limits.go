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
	fresh := l.hits[userID][:0]
	for _, at := range l.hits[userID] {
		if now.Sub(at) < l.window {
			fresh = append(fresh, at)
		}
	}
	if len(fresh) >= l.limit {
		l.hits[userID] = fresh
		return false
	}
	l.hits[userID] = append(fresh, now)
	return true
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
