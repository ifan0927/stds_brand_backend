package database

import (
	"context"
	"database/sql"

	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const propertyAvailabilityQuery = `SELECT property_id, property_public_name, address, has_vacant_room FROM approved_brand_property_availability_v1`

// PropertyAvailabilityRepository reads approved property availability from the readonly database.
type PropertyAvailabilityRepository struct {
	db *sql.DB
}

// NewPropertyAvailabilityRepository builds a PropertyAvailabilityRepository with the given database handle.
func NewPropertyAvailabilityRepository(db *sql.DB) PropertyAvailabilityRepository {
	return PropertyAvailabilityRepository{db: db}
}

// ListPropertyAvailability returns approved property availability from the readonly database.
func (repository PropertyAvailabilityRepository) ListPropertyAvailability(ctx context.Context) ([]application.PropertyAvailability, error) {
	if repository.db == nil {
		return nil, errReadonlyDatabaseNotConfigured
	}

	rows, err := repository.db.QueryContext(ctx, propertyAvailabilityQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]application.PropertyAvailability, 0)
	for rows.Next() {
		var item application.PropertyAvailability
		if err := rows.Scan(&item.PropertyID, &item.PropertyPublicName, &item.Address, &item.HasVacantRoom); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
