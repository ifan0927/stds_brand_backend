package application

import (
	"context"
	"errors"
	"testing"
)

type fakePropertyAvailabilityRepository struct {
	items []PropertyAvailability
	err   error
}

// ListPropertyAvailability returns fake property availability repository results.
func (repository fakePropertyAvailabilityRepository) ListPropertyAvailability(context.Context) ([]PropertyAvailability, error) {
	return repository.items, repository.err
}

// TestPropertyAvailabilityServiceReturnsRepositoryItems verifies that the service returns repository items.
func TestPropertyAvailabilityServiceReturnsRepositoryItems(t *testing.T) {
	items := []PropertyAvailability{
		{
			PropertyID:         "00000000-0000-0000-0000-000000000000",
			PropertyPublicName: "北車館",
			Address:            "台北市中正區範例路 1 號",
			HasVacantRoom:      true,
		},
	}

	got, err := NewPropertyAvailabilityService(fakePropertyAvailabilityRepository{items: items}).ListPropertyAvailability(context.Background())
	if err != nil {
		t.Fatalf("expected property availability items, got error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 property availability item, got %d", len(got))
	}
	if got[0].PropertyPublicName != items[0].PropertyPublicName {
		t.Fatalf("expected property public name %q, got %q", items[0].PropertyPublicName, got[0].PropertyPublicName)
	}
}

// TestPropertyAvailabilityServiceReturnsDependencyUnavailableWithoutRepository verifies the nil repository error path.
func TestPropertyAvailabilityServiceReturnsDependencyUnavailableWithoutRepository(t *testing.T) {
	items, err := NewPropertyAvailabilityService(nil).ListPropertyAvailability(context.Background())
	if !errors.Is(err, ErrDependencyUnavailable) {
		t.Fatalf("expected dependency unavailable, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil property availability items, got %#v", items)
	}
}
