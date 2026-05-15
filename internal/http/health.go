package http

import (
	"context"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const healthCheckTimeout = 2 * time.Second

// HealthChecker verifies whether the service dependency is ready.
type HealthChecker interface {
	// Check returns an error when the service dependency is unavailable.
	Check(ctx context.Context) error
}

func handleHealth(checker HealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
		defer cancel()

		if checker == nil || checker.Check(ctx) != nil {
			writeServiceUnavailableError(c)
			return
		}

		c.JSON(nethttp.StatusOK, gin.H{"status": "ok"})
	}
}
