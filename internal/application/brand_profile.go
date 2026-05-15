package application

import (
	"context"
	"time"
)

type BrandProfile struct {
	BrandName      string
	ContactPhone   *string
	ContactEmail   *string
	ContactAddress *string
	UpdatedAt      time.Time
}

type BrandProfileRepository interface {
	GetBrandProfile(ctx context.Context) (*BrandProfile, error)
}

type BrandProfileService struct {
	repository BrandProfileRepository
}

func NewBrandProfileService(repository BrandProfileRepository) BrandProfileService {
	return BrandProfileService{repository: repository}
}

func (service BrandProfileService) GetBrandProfile(ctx context.Context) (*BrandProfile, error) {
	if service.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	return service.repository.GetBrandProfile(ctx)
}
