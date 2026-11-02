package middleware

import (
	"engflex-api/internal/common"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// BodyLimits applies before JSON or multipart parsing, including chunked bodies.
// Provider audio caps are rechecked after parsing by their existing services.
func BodyLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := int64(1 << 20)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/internal/conversations/") {
			limit = 4 << 20
		} else if strings.HasSuffix(path, "/pronounce") || strings.HasSuffix(path, "/transcribe") {
			limit = (2 << 20) + (256 << 10)
		}
		if c.Request.ContentLength > limit {
			common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "request body too large"))
			c.Abort()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
