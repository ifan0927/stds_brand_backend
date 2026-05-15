package http

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	HealthChecker       HealthChecker
	BrandProfileService BrandProfileService
	FAQService          FAQService
	Logger              *slog.Logger
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	router.Use(requestIDMiddleware(), structuredLoggingMiddleware(logger))
	router.GET("/health", handleHealth(deps.HealthChecker))
	router.GET("/api/v1/brand/profile", handleBrandProfile(deps.BrandProfileService))
	router.GET("/api/v1/brand/faqs", handleFAQs(deps.FAQService))
	router.NoRoute(writeNotFoundError)
	return router
}
