package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A panic travelling through sentrygin (Repanic: true) into our Recovery
// must still produce the shared {message} 500 envelope — never Gin's
// default HTML error page.
func TestRecoveryAfterSentryRepanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recovery(), sentrygin.New(sentrygin.Options{Repanic: true, WaitForDelivery: false}))
	r.GET("/boom", func(c *gin.Context) {
		panic("y tho")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), `"message"`)
	require.NotContains(t, w.Body.String(), "<html>")
}

// The repanicked panic must ALSO reach Sentry: exactly one event alongside
// the envelope. The client below never touches the network — BeforeSend
// records and drops.
func TestSentryCapturesRepanickedPanic(t *testing.T) {
	var mu sync.Mutex
	var events []*sentry.Event
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://test@test.ingest.sentry.io/1",
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			mu.Lock()
			events = append(events, event)
			mu.Unlock()
			return nil
		},
	})
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recovery(), sentrygin.New(sentrygin.Options{Repanic: true, WaitForDelivery: false}))
	r.GET("/boom", func(c *gin.Context) {
		panic("y tho")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	r.ServeHTTP(w, req)
	require.True(t, sentry.Flush(2*time.Second))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, events, 1, "repanicked panic must produce exactly one Sentry event")
}
