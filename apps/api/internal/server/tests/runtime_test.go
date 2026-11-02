package tests

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"engflex-api/config"
	"engflex-api/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func runtimeForTest(t *testing.T) (*server.Runtime, *gorm.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to migrated PostgreSQL")
	}
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	return server.NewRuntime(db, config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost"}}}), db
}

func TestReadinessDistinguishesDatabaseFailureFromLiveness(t *testing.T) {
	runtime, db := runtimeForTest(t)
	pool, err := db.DB()
	require.NoError(t, err)
	for _, tt := range []struct {
		name, path string
		closeDB    bool
		status     int
	}{
		{"database ready", "/readyz", false, http.StatusOK},
		{"database unavailable", "/readyz", true, http.StatusServiceUnavailable},
		{"liveness despite database failure", "/health", false, http.StatusOK},
		{"legacy liveness despite database failure", "/healthz", false, http.StatusOK},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.closeDB {
				require.NoError(t, pool.Close())
			}
			response := httptest.NewRecorder()
			runtime.HTTP.Handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
			require.Equal(t, tt.status, response.Code, response.Body.String())
		})
	}
}

func TestShutdownDrainsRequestsBeforeStopping(t *testing.T) {
	for _, tt := range []struct {
		name     string
		inflight bool
	}{{"idle server", false}, {"active request completes during drain", true}} {
		t.Run(tt.name, func(t *testing.T) {
			runtime, _ := runtimeForTest(t)
			runtime.HTTP.Addr = "127.0.0.1:0"
			address := make(chan string, 1)
			runtime.HTTP.BaseContext = func(listener net.Listener) context.Context {
				address <- listener.Addr().String()
				return context.Background()
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var releaseOnce sync.Once
			finishRequest := func() { releaseOnce.Do(func() { close(release) }) }
			t.Cleanup(finishRequest)
			routes := runtime.HTTP.Handler
			runtime.HTTP.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/probe" {
					routes.ServeHTTP(w, r)
					return
				}
				close(entered)
				<-release
				w.WriteHeader(http.StatusOK)
			})
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			stopped := make(chan error, 1)
			go func() { stopped <- runtime.Run(ctx) }()
			var origin string
			select {
			case addr := <-address:
				origin = "http://" + addr
			case <-time.After(5 * time.Second):
				t.Fatal("server did not listen")
			}
			transport := &http.Transport{}
			t.Cleanup(transport.CloseIdleConnections)
			client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
			responseStatus := make(chan int, 1)
			if tt.inflight {
				go func() {
					response, err := client.Get(origin + "/probe")
					if err != nil {
						responseStatus <- 0
						return
					}
					defer response.Body.Close()
					responseStatus <- response.StatusCode
				}()
				select {
				case <-entered:
				case <-time.After(5 * time.Second):
					t.Fatal("request did not enter handler")
				}
			}
			cancel()
			require.Eventually(t, func() bool {
				response := httptest.NewRecorder()
				routes.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
				return response.Code == http.StatusServiceUnavailable
			}, 5*time.Second, 10*time.Millisecond)
			if tt.inflight {
				select {
				case <-stopped:
					t.Fatal("server stopped before the active request finished")
				default:
				}
				finishRequest()
				require.Equal(t, http.StatusOK, <-responseStatus)
			}
			require.NoError(t, <-stopped)
			_, err := client.Get(origin + "/health")
			require.Error(t, err, "new requests must not be admitted after shutdown")
		})
	}
}
