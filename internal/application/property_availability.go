package application

import "context"

// PropertyAvailability describes public availability for one property.
type PropertyAvailability struct {
	PropertyID         string
	PropertyPublicName string
	Address            string
	HasVacantRoom      bool
}

// PropertyAvailabilityRepository loads approved property availability rows.
type PropertyAvailabilityRepository interface {
	// ListPropertyAvailability returns approved property availability rows.
	ListPropertyAvailability(ctx context.Context) ([]PropertyAvailability, error)
}

// PropertyAvailabilityService coordinates property availability retrieval.
type PropertyAvailabilityService struct {
	repository PropertyAvailabilityRepository
}

// NewPropertyAvailabilityService builds a PropertyAvailabilityService with the given repository.
func NewPropertyAvailabilityService(repository PropertyAvailabilityRepository) PropertyAvailabilityService {
	return PropertyAvailabilityService{repository: repository}
}

// ListPropertyAvailability returns approved public property availability.
func (service PropertyAvailabilityService) ListPropertyAvailability(ctx context.Context) ([]PropertyAvailability, error) {
	if service.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	return service.repository.ListPropertyAvailability(ctx)
}
