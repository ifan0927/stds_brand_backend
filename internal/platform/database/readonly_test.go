package database

import (
	"context"
	"testing"
)

// TestReadonlyHealthCheckerWithoutDatabaseURLIsUnavailable verifies missing database configuration.
func TestReadonlyHealthCheckerWithoutDatabaseURLIsUnavailable(t *testing.T) {
	checker, err := NewReadonlyHealthChecker(context.Background(), "")
	if err != nil {
		t.Fatalf("expected empty database URL to create unavailable checker, got %v", err)
	}
	defer checker.Close()

	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("expected health check to fail when readonly database is not configured")
	}
}
