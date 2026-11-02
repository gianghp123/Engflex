package middleware

import (
	"engflex-api/internal/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimitPrecedesParsingAndPreserves413(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BodyLimits())
	router.POST("/probe", func(c *gin.Context) {
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err != nil {
			common.Fail(c, common.InvalidRequest(err))
			return
		}
		common.OK(c, "ok", nil)
	})
	for _, length := range []int64{-1, 2 << 20} {
		request := httptest.NewRequest(http.MethodPost, "/probe", strings.NewReader(`{"payload":"`+strings.Repeat("x", 2<<20)+`"}`))
		request.ContentLength = length
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code, response.Body.String())
	}
}
