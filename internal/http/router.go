package http

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	HealthChecker HealthChecker
	Logger        *slog.Logger
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	router.Use(requestIDMiddleware(), structuredLoggingMiddleware(logger))
	router.GET("/health", handleHealth(deps.HealthChecker))
	router.NoRoute(writeNotFoundError)
	return router
}
