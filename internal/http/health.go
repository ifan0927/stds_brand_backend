package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const healthCheckTimeout = 2 * time.Second

type HealthChecker interface {
	Check(ctx context.Context) error
}

type publicErrorResponse struct {
	Error publicError `json:"error"`
}

type publicError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleHealth(checker HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
		defer cancel()

		if checker == nil || checker.Check(ctx) != nil {
			c.JSON(nethttp.StatusServiceUnavailable, publicErrorResponse{
				Error: publicError{
					Code:    "SERVICE_UNAVAILABLE",
					Message: "Service unavailable.",
				},
			})
			return
		}

		c.JSON(nethttp.StatusOK, gin.H{"status": "ok"})
	}
}
