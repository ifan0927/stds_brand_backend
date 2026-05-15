package database

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestPropertyAvailabilityRepositoryScansApprovedItems verifies property availability row scanning.
func TestPropertyAvailabilityRepositoryScansApprovedItems(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"property_id",
		"property_public_name",
		"address",
		"has_vacant_room",
	}).
		AddRow("00000000-0000-0000-0000-000000000000", "北車館", "台北市中正區範例路 1 號", true).
		AddRow("11111111-1111-1111-1111-111111111111", "南港館", "台北市南港區範例路 2 號", false)
	mock.ExpectQuery(regexp.QuoteMeta(propertyAvailabilityQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewPropertyAvailabilityRepository(db).ListPropertyAvailability(context.Background())
	if err != nil {
		t.Fatalf("expected property availability items, got error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 property availability items, got %d", len(items))
	}
	if items[0].PropertyID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("expected first property ID, got %q", items[0].PropertyID)
	}
	if items[0].PropertyPublicName != "北車館" {
		t.Fatalf("expected first property public name, got %q", items[0].PropertyPublicName)
	}
	if items[0].Address != "台北市中正區範例路 1 號" {
		t.Fatalf("expected first address, got %q", items[0].Address)
	}
	if !items[0].HasVacantRoom {
		t.Fatal("expected first property to have vacant room")
	}
}

// TestPropertyAvailabilityRepositoryReturnsEmptyListWhenNoApprovedItemsExist verifies empty property availability results.
func TestPropertyAvailabilityRepositoryReturnsEmptyListWhenNoApprovedItemsExist(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"property_id", "property_public_name", "address", "has_vacant_room"})
	mock.ExpectQuery(regexp.QuoteMeta(propertyAvailabilityQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewPropertyAvailabilityRepository(db).ListPropertyAvailability(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for empty property availability list, got %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil empty property availability list")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty property availability list, got %#v", items)
	}
}

// TestPropertyAvailabilityRepositoryReturnsScanError verifies property availability scan error propagation.
func TestPropertyAvailabilityRepositoryReturnsScanError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"property_id",
		"property_public_name",
		"address",
		"has_vacant_room",
	}).AddRow("00000000-0000-0000-0000-000000000000", "北車館", "台北市中正區範例路 1 號", "not-a-bool")
	mock.ExpectQuery(regexp.QuoteMeta(propertyAvailabilityQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewPropertyAvailabilityRepository(db).ListPropertyAvailability(context.Background())
	if err == nil {
		t.Fatal("expected scan error")
	}
	if items != nil {
		t.Fatalf("expected nil property availability items, got %#v", items)
	}
}

// TestPropertyAvailabilityRepositoryReturnsRowsError verifies property availability row iteration error propagation.
func TestPropertyAvailabilityRepositoryReturnsRowsError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	expectedErr := errors.New("row failed")
	rows := sqlmock.NewRows([]string{
		"property_id",
		"property_public_name",
		"address",
		"has_vacant_room",
	}).
		AddRow("00000000-0000-0000-0000-000000000000", "北車館", "台北市中正區範例路 1 號", true).
		RowError(0, expectedErr)
	mock.ExpectQuery(regexp.QuoteMeta(propertyAvailabilityQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewPropertyAvailabilityRepository(db).ListPropertyAvailability(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rows error, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil property availability items, got %#v", items)
	}
}

// TestPropertyAvailabilityRepositoryReturnsDatabaseError verifies property availability query error propagation.
func TestPropertyAvailabilityRepositoryReturnsDatabaseError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	expectedErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta(propertyAvailabilityQuery)).WithArgs().WillReturnError(expectedErr)

	items, err := NewPropertyAvailabilityRepository(db).ListPropertyAvailability(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil property availability items, got %#v", items)
	}
}
