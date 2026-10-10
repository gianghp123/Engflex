// Package logger builds the process-wide structured logger. Services log
// through the standard library's slog package-level functions using
// *Context methods; the handler installed here lifts request-scoped values
// out of the context (the pattern recommended by the slog authors: see the
// Go blog, "Structured Logging with slog").
package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"

	"engflex-api/internal/common"
)

// Context keys; kept private so only the helpers below can set them.
type contextKey int

const (
	requestIDKey contextKey = iota
	userIDKey
)

// WithRequestID returns a context whose log records carry request_id.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID returns a context whose log records carry user_id.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// New builds a logger that writes to stdout: text in development, JSON
// elsewhere. level accepts debug, info, warn, or error (default info);
// debug also enables source locations.
func New(level, format string) *slog.Logger {
	return NewWithWriter(os.Stdout, level, format)
}

// NewWithWriter is New with an injectable sink, so tests and alternative
// destinations (buffers, files) can capture records. Text output is
// colorized only when the writer is a terminal.
func NewWithWriter(w io.Writer, level, format string) *slog.Logger {
	parsed := parseLevel(level)
	opts := &slog.HandlerOptions{Level: parsed}
	opts.AddSource = parsed == slog.LevelDebug

	var base slog.Handler
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		base = slog.NewJSONHandler(w, opts)
	} else {
		if isTerminal(w) {
			w = colorWriter{w: w}
		}
		base = slog.NewTextHandler(w, opts)
	}
	return slog.New(contextHandler{Handler: base})
}

// isTerminal reports whether w writes to a character device (a tty), so
// ANSI colors never leak into files, pipes, or log collectors.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// levelColors maps the encoded level token to its ANSI-painted form:
// debug cyan, info green, warn yellow, error red.
var levelColors = []struct{ plain, colored string }{
	{"level=DEBUG", "level=\x1b[36mDEBUG\x1b[0m"},
	{"level=INFO", "level=\x1b[32mINFO\x1b[0m"},
	{"level=WARN", "level=\x1b[33mWARN\x1b[0m"},
	{"level=ERROR", "level=\x1b[31mERROR\x1b[0m"},
}

// colorWriter paints the level token of each text record. slog's
// TextHandler escapes control characters inside attribute values, so a
// colored level cannot come from ReplaceAttr — the rewrite happens on the
// encoded stream instead.
type colorWriter struct {
	w io.Writer
}

func (cw colorWriter) Write(p []byte) (int, error) {
	for _, lc := range levelColors {
		p = bytes.ReplaceAll(p, []byte(lc.plain), []byte(lc.colored))
	}
	return cw.w.Write(p)
}

// Report logs a failed service operation, choosing the level from the
// error's HTTP stance: expected client-facing failures (4xx) log at Warn,
// server faults and unknown errors at Error. Extra attrs are key/value
// pairs appended to the record.
func Report(ctx context.Context, msg string, err error, attrs ...any) {
	level := slog.LevelError
	var appErr *common.AppError
	if errors.As(err, &appErr) && appErr.Status < http.StatusInternalServerError {
		level = slog.LevelWarn
	}
	slog.Log(ctx, level, msg, append([]any{"error", err}, attrs...)...)
	// Server faults (5xx/unknown) also go to Sentry when it is enabled.
	// Client errors stay in logs only. Panics never reach here: the Sentry
	// gin middleware captures them and re-panics into Recovery, which logs
	// via slog — exactly one event per failure either way.
	if level == slog.LevelError && sentry.CurrentHub().Client() != nil {
		sentry.CaptureException(err)
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// contextHandler wraps a handler and prepends request-scoped attributes
// found in the context, so every record produced with a request context
// carries request_id and user_id without callers repeating them.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if id, ok := ctx.Value(requestIDKey).(string); ok && id != "" {
		record.AddAttrs(slog.String("request_id", id))
	}
	if id, ok := ctx.Value(userIDKey).(string); ok && id != "" {
		record.AddAttrs(slog.String("user_id", id))
	}
	return h.Handler.Handle(ctx, record)
}

// WithAttrs and WithGroup must keep the wrapper so derived loggers still
// read the context; the embedded handler would otherwise be returned.
func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}
