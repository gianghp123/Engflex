// Package observability bootstraps process-wide observability sinks.
// Sentry is optional: an empty DSN leaves the process running exactly as
// before, with no network traffic.
package observability

import (
	"github.com/getsentry/sentry-go"

	"engflex-api/config"
)

// Init installs the Sentry client when cfg carries a DSN and reports
// whether events will leave the process. An empty DSN is a normal
// disabled state, not an error. A malformed DSN returns an error and the
// caller keeps the server running without Sentry: observability must
// never block boot (unlike required keys such as LLM_API_KEY, which fail
// fast because scoring cannot work without them).
func Init(cfg config.SentryConfig) (bool, error) {
	if cfg.DSN == "" {
		return false, nil
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Debug:            cfg.Debug,
		SendDefaultPII:   false,
		EnableTracing:    true,
		TracesSampleRate: cfg.TracesSampleRate,
		BeforeSend:       scrubEvent,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// scrubEvent strips credentials before any event leaves the process. The
// API authenticates with Clerk Bearer tokens, which must never reach a
// third party even though SendDefaultPII is off.
func scrubEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if event.Request != nil {
		delete(event.Request.Headers, "Authorization")
	}
	return event
}
