package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/studioauth"
	"github.com/gin-gonic/gin"
)

const defaultStudioAuthBodyLimit int64 = 1 << 20

func StudioAuth(verifier *studioauth.Verifier, maxBodyBytes int64) gin.HandlerFunc {
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultStudioAuthBodyLimit
	}
	return func(c *gin.Context) {
		if verifier == nil || c.Request == nil || c.Request.Body == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "service_unavailable"}})
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes+1))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_request"}})
			return
		}
		if int64(len(body)) > maxBodyBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": gin.H{"code": "request_too_large"}})
			return
		}
		request := studioauth.SignedRequest{
			Method:  c.Request.Method,
			Path:    c.Request.URL.Path,
			Headers: c.Request.Header,
			Body:    body,
		}
		if err := verifier.Verify(c.Request.Context(), request); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "unauthorized_request"}})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
