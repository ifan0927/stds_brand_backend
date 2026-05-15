package database

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var errReadonlyDatabaseNotConfigured = errors.New("readonly database is not configured")

type ReadonlyHealthChecker struct {
	db *sql.DB
}

func NewReadonlyHealthChecker(_ context.Context, databaseURL string) (*ReadonlyHealthChecker, error) {
	if databaseURL == "" {
		return &ReadonlyHealthChecker{}, nil
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	return &ReadonlyHealthChecker{db: db}, nil
}

func (checker *ReadonlyHealthChecker) Check(ctx context.Context) error {
	if checker == nil || checker.db == nil {
		return errReadonlyDatabaseNotConfigured
	}

	for _, query := range readonlyHealthQueries {
		if _, err := checker.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (checker *ReadonlyHealthChecker) DB() *sql.DB {
	if checker == nil {
		return nil
	}
	return checker.db
}

func (checker *ReadonlyHealthChecker) Close() {
	if checker == nil || checker.db == nil {
		return
	}
	_ = checker.db.Close()
}

var readonlyHealthQueries = []string{
	"SELECT 1 FROM approved_brand_profile_v1 LIMIT 0",
	"SELECT 1 FROM approved_brand_faq_items_v1 LIMIT 0",
	"SELECT 1 FROM approved_brand_property_availability_v1 LIMIT 0",
}
