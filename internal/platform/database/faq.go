package database

import (
	"context"
	"database/sql"

	"github.com/ifan0927/stds_brand_backend/internal/application"
)

const faqItemsQuery = `SELECT question, answer, sort_order FROM approved_brand_faq_items_v1 ORDER BY sort_order`

type FAQRepository struct {
	db *sql.DB
}

func NewFAQRepository(db *sql.DB) FAQRepository {
	return FAQRepository{db: db}
}

func (repository FAQRepository) ListFAQItems(ctx context.Context) ([]application.FAQItem, error) {
	if repository.db == nil {
		return nil, errReadonlyDatabaseNotConfigured
	}

	rows, err := repository.db.QueryContext(ctx, faqItemsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]application.FAQItem, 0)
	for rows.Next() {
		var item application.FAQItem
		if err := rows.Scan(&item.Question, &item.Answer, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
