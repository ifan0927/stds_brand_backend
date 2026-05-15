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
