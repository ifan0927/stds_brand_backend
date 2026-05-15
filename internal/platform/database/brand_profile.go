package database

import (
	"context"
	"database/sql"

	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const brandProfileQuery = `SELECT brand_name, contact_phone, contact_email, contact_address, updated_at FROM approved_brand_profile_v1 LIMIT 1`

// BrandProfileRepository reads the approved brand profile from the readonly database.
type BrandProfileRepository struct {
	db *sql.DB
}

// NewBrandProfileRepository builds a BrandProfileRepository with the given database handle.
func NewBrandProfileRepository(db *sql.DB) BrandProfileRepository {
	return BrandProfileRepository{db: db}
}

// GetBrandProfile returns the approved brand profile from the readonly database.
func (repository BrandProfileRepository) GetBrandProfile(ctx context.Context) (*application.BrandProfile, error) {
	if repository.db == nil {
		return nil, errReadonlyDatabaseNotConfigured
	}

	var profile application.BrandProfile
	if err := repository.db.QueryRowContext(ctx, brandProfileQuery).Scan(
		&profile.BrandName,
		&profile.ContactPhone,
		&profile.ContactEmail,
		&profile.ContactAddress,
		&profile.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &profile, nil
}
