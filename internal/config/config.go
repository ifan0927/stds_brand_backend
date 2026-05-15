package config

import (
	"os"
)

const (
	defaultAppPort = "8080"
)

// Config contains runtime configuration for the brand backend.
type Config struct {
	AppPort                  string
	BrandReadonlyDatabaseURL string
}

// Load reads runtime configuration from the environment.
func Load() Config {
	return Config{
		AppPort:                  stringWithDefault(os.Getenv("APP_PORT"), defaultAppPort),
		BrandReadonlyDatabaseURL: os.Getenv("BRAND_READONLY_DATABASE_URL"),
	}
}

func stringWithDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
