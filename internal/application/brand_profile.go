package application

import (
	"context"
	"time"
)

// BrandProfile describes the public brand contact profile.
type BrandProfile struct {
	BrandName      string
	ContactPhone   *string
	ContactEmail   *string
	ContactAddress *string
	UpdatedAt      time.Time
}

// BrandProfileRepository loads the approved brand profile.
type BrandProfileRepository interface {
	// GetBrandProfile returns the approved brand profile.
	GetBrandProfile(ctx context.Context) (*BrandProfile, error)
}

// BrandProfileService coordinates brand profile retrieval.
type BrandProfileService struct {
	repository BrandProfileRepository
}

// NewBrandProfileService builds a BrandProfileService with the given repository.
func NewBrandProfileService(repository BrandProfileRepository) BrandProfileService {
	return BrandProfileService{repository: repository}
}

// GetBrandProfile returns the approved public brand profile.
func (service BrandProfileService) GetBrandProfile(ctx context.Context) (*BrandProfile, error) {
	if service.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	return service.repository.GetBrandProfile(ctx)
}
