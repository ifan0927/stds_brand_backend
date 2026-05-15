package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifan0927/stds_brand_backend/internal/application"
	"github.com/ifan0927/stds_brand_backend/internal/config"
	brandhttp "github.com/ifan0927/stds_brand_backend/internal/http"
	"github.com/ifan0927/stds_brand_backend/internal/platform/database"
)

func main() {
	if err := config.LoadDotenv(); err != nil {
		log.Fatalf("failed to load dotenv file: %v", err)
	}

	cfg := config.Load()
	gin.SetMode(gin.ReleaseMode)

	checker, err := database.NewReadonlyHealthChecker(context.Background(), cfg.BrandReadonlyDatabaseURL)
	if err != nil {
		log.Fatalf("failed to configure readonly database: %v", err)
	}
	defer checker.Close()

	brandProfileRepository := database.NewBrandProfileRepository(checker.DB())
	brandProfileService := application.NewBrandProfileService(brandProfileRepository)

	server := &http.Server{
		Addr: ":" + cfg.AppPort,
		Handler: brandhttp.NewRouter(brandhttp.Dependencies{
			HealthChecker:       checker,
			BrandProfileService: brandProfileService,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting brand backend on port %s", cfg.AppPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server stopped: %v", err)
	}
}
