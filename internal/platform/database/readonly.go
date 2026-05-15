package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var errReadonlyDatabaseNotConfigured = errors.New("readonly database is not configured")

type ReadonlyHealthChecker struct {
	pool *pgxpool.Pool
}

func NewReadonlyHealthChecker(ctx context.Context, databaseURL string) (*ReadonlyHealthChecker, error) {
	if databaseURL == "" {
		return &ReadonlyHealthChecker{}, nil
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	return &ReadonlyHealthChecker{pool: pool}, nil
}

func (checker *ReadonlyHealthChecker) Check(ctx context.Context) error {
	if checker == nil || checker.pool == nil {
		return errReadonlyDatabaseNotConfigured
	}

	for _, query := range readonlyHealthQueries {
		if _, err := checker.pool.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (checker *ReadonlyHealthChecker) Close() {
	if checker == nil || checker.pool == nil {
		return
	}
	checker.pool.Close()
}

var readonlyHealthQueries = []string{
	"SELECT 1 FROM approved_brand_profile_v1 LIMIT 0",
	"SELECT 1 FROM approved_brand_faq_items_v1 LIMIT 0",
	"SELECT 1 FROM approved_brand_property_availability_v1 LIMIT 0",
}
