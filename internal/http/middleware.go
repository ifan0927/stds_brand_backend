package http

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	requestIDHeader    = "X-Request-ID"
	requestIDContext   = "request_id"
	maxRequestIDLength = 128
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := sanitizeRequestID(c.GetHeader(requestIDHeader))
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(requestIDContext, requestID)
		c.Header(requestIDHeader, requestID)
		c.Next()
	}
}

func structuredLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(requestIDContext)
		logger.Info("request completed",
			slog.Any("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency", time.Since(start).Milliseconds()),
		)
	}
}

func sanitizeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxRequestIDLength || !utf8.ValidString(value) {
		return ""
	}

	for _, char := range value {
		if char < 0x21 || char > 0x7e {
			return ""
		}
	}

	return value
}

func generateRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		fallback := sha256.Sum256([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
		return hex.EncodeToString(fallback[:16])
	}
	return hex.EncodeToString(bytes[:])
}
