package http

import "github.com/gin-gonic/gin"

type Dependencies struct {
	HealthChecker HealthChecker
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.New()
	router.GET("/health", handleHealth(deps.HealthChecker))
	return router
}
