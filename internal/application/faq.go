package application

import "context"

type FAQItem struct {
	Question  string
	Answer    string
	SortOrder int
}

type FAQRepository interface {
	ListFAQItems(ctx context.Context) ([]FAQItem, error)
}

type FAQService struct {
	repository FAQRepository
}

func NewFAQService(repository FAQRepository) FAQService {
	return FAQService{repository: repository}
}

func (service FAQService) ListFAQItems(ctx context.Context) ([]FAQItem, error) {
	if service.repository == nil {
		return nil, ErrDependencyUnavailable
	}
	return service.repository.ListFAQItems(ctx)
}
