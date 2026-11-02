package server

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"engflex-api/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Runtime owns admission and shutdown; optional AI providers do not gate readiness.
type Runtime struct {
	HTTP     *http.Server
	draining atomic.Bool
	cleanup  func()
}

func NewRuntime(db *gorm.DB, cfg config.Config) *Runtime {
	runtime := &Runtime{}
	router, cleanup := newRouter(db, cfg)
	runtime.cleanup = cleanup
	router.GET("/readyz", func(c *gin.Context) {
		if runtime.draining.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "draining"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		pool, err := db.DB()
		if err != nil || pool.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	runtime.HTTP = &http.Server{Addr: cfg.Server.Addr(), Handler: router,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 150 * time.Second, IdleTimeout: 60 * time.Second}
	return runtime
}

func (r *Runtime) Run(ctx context.Context) error {
	defer r.cleanup()
	failed := make(chan error, 1)
	go func() { failed <- r.HTTP.ListenAndServe() }()
	select {
	case err := <-failed:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		r.draining.Store(true)
		shutdown, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()
		if err := r.HTTP.Shutdown(shutdown); err != nil {
			_ = r.HTTP.Close()
			return err
		}
		return nil
	}
}
