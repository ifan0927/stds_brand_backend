package config

import "testing"

// TestLoadUsesDefaultsAndEnvironment verifies environment-backed configuration loading.
func TestLoadUsesDefaultsAndEnvironment(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("BRAND_READONLY_DATABASE_URL", "postgres://brand_readonly:brand_readonly@localhost:5432/stds_backend?sslmode=disable")

	cfg := Load()

	if cfg.AppPort != defaultAppPort {
		t.Fatalf("expected default app port %q, got %q", defaultAppPort, cfg.AppPort)
	}
	if cfg.BrandReadonlyDatabaseURL != "postgres://brand_readonly:brand_readonly@localhost:5432/stds_backend?sslmode=disable" {
		t.Fatalf("unexpected readonly database URL: %q", cfg.BrandReadonlyDatabaseURL)
	}
}
