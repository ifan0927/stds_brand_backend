package application

import "context"

// FAQItem describes one public brand FAQ entry.
type FAQItem struct {
	Question  string
	Answer    string
	SortOrder int
}

// FAQRepository loads approved FAQ items for the brand API.
type FAQRepository interface {
	// ListFAQItems returns approved FAQ items.
	ListFAQItems(ctx context.Context) ([]FAQItem, error)
}

// FAQService coordinates brand FAQ retrieval.
type FAQService struct {
	repository FAQRepository
}

// NewFAQService builds an FAQService with the given repository.
func NewFAQService(repository FAQRepository) FAQService {
	return FAQService{repository: repository}
}

// ListFAQItems returns approved FAQ items in display order.
func (service FAQService) ListFAQItems(ctx context.Context) ([]FAQItem, error) {
	if service.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	return service.repository.ListFAQItems(ctx)
}
