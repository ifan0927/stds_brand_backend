package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBrandProfileRepositoryScansApprovedProfile(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	updatedAt := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"brand_name",
		"contact_phone",
		"contact_email",
		"contact_address",
		"updated_at",
	}).AddRow("STDS", "02-1234-5678", "hello@example.com", "台北市中正區範例路 1 號", updatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(brandProfileQuery)).WithArgs().WillReturnRows(rows)

	profile, err := NewBrandProfileRepository(db).GetBrandProfile(context.Background())
	if err != nil {
		t.Fatalf("expected profile, got error: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile")
	}
	if profile.BrandName != "STDS" {
		t.Fatalf("expected brand name STDS, got %q", profile.BrandName)
	}
	if profile.ContactPhone == nil || *profile.ContactPhone != "02-1234-5678" {
		t.Fatalf("expected contact phone, got %#v", profile.ContactPhone)
	}
	if profile.ContactEmail == nil || *profile.ContactEmail != "hello@example.com" {
		t.Fatalf("expected contact email, got %#v", profile.ContactEmail)
	}
	if profile.ContactAddress == nil || *profile.ContactAddress != "台北市中正區範例路 1 號" {
		t.Fatalf("expected contact address, got %#v", profile.ContactAddress)
	}
	if !profile.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("expected updated_at %s, got %s", updatedAt, profile.UpdatedAt)
	}
}

func TestBrandProfileRepositoryReturnsNilWhenNoApprovedProfileExists(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(brandProfileQuery)).WithArgs().WillReturnError(sql.ErrNoRows)

	profile, err := NewBrandProfileRepository(db).GetBrandProfile(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for missing singleton, got %v", err)
	}
	if profile != nil {
		t.Fatalf("expected nil profile, got %#v", profile)
	}
}

func TestBrandProfileRepositoryReturnsScanError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"brand_name",
		"contact_phone",
		"contact_email",
		"contact_address",
		"updated_at",
	}).AddRow("STDS", nil, nil, nil, "not-a-time")
	mock.ExpectQuery(regexp.QuoteMeta(brandProfileQuery)).WithArgs().WillReturnRows(rows)

	profile, err := NewBrandProfileRepository(db).GetBrandProfile(context.Background())
	if err == nil {
		t.Fatal("expected scan error")
	}
	if profile != nil {
		t.Fatalf("expected nil profile, got %#v", profile)
	}
}

func TestBrandProfileRepositoryReturnsDatabaseError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	expectedErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta(brandProfileQuery)).WithArgs().WillReturnError(expectedErr)

	profile, err := NewBrandProfileRepository(db).GetBrandProfile(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if profile != nil {
		t.Fatalf("expected nil profile, got %#v", profile)
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return db, mock, func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet SQL expectations: %v", err)
		}
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close mock database: %v", err)
		}
	}
}
