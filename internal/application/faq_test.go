package application

import (
	"context"
	"errors"
	"testing"
)

type fakeFAQRepository struct {
	items []FAQItem
	err   error
}

func (repository fakeFAQRepository) ListFAQItems(context.Context) ([]FAQItem, error) {
	return repository.items, repository.err
}

func TestFAQServiceReturnsRepositoryItems(t *testing.T) {
	items := []FAQItem{
		{
			Question:  "如何預約看房？",
			Answer:    "請透過公開聯絡方式與我們確認可預約時段。",
			SortOrder: 10,
		},
	}

	got, err := NewFAQService(fakeFAQRepository{items: items}).ListFAQItems(context.Background())
	if err != nil {
		t.Fatalf("expected FAQ items, got error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 FAQ item, got %d", len(got))
	}
	if got[0].Question != items[0].Question {
		t.Fatalf("expected question %q, got %q", items[0].Question, got[0].Question)
	}
}

func TestFAQServiceReturnsDependencyUnavailableWithoutRepository(t *testing.T) {
	items, err := NewFAQService(nil).ListFAQItems(context.Background())
	if !errors.Is(err, ErrDependencyUnavailable) {
		t.Fatalf("expected dependency unavailable, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil FAQ items, got %#v", items)
	}
}
