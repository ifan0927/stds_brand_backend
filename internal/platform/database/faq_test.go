package database

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// TestFAQRepositoryScansApprovedItems verifies FAQ row scanning.
func TestFAQRepositoryScansApprovedItems(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"question",
		"answer",
		"sort_order",
	}).
		AddRow("如何預約看房？", "請透過公開聯絡方式與我們確認可預約時段。", 10).
		AddRow("是否可以線上詢問？", "可以，請使用公開聯絡方式。", 20)
	mock.ExpectQuery(regexp.QuoteMeta(faqItemsQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewFAQRepository(db).ListFAQItems(context.Background())
	if err != nil {
		t.Fatalf("expected FAQ items, got error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 FAQ items, got %d", len(items))
	}
	if items[0].Question != "如何預約看房？" {
		t.Fatalf("expected first question, got %q", items[0].Question)
	}
	if items[0].Answer != "請透過公開聯絡方式與我們確認可預約時段。" {
		t.Fatalf("expected first answer, got %q", items[0].Answer)
	}
	if items[0].SortOrder != 10 {
		t.Fatalf("expected first sort order 10, got %d", items[0].SortOrder)
	}
}

// TestFAQRepositoryReturnsEmptyListWhenNoApprovedItemsExist verifies empty FAQ results.
func TestFAQRepositoryReturnsEmptyListWhenNoApprovedItemsExist(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"question", "answer", "sort_order"})
	mock.ExpectQuery(regexp.QuoteMeta(faqItemsQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewFAQRepository(db).ListFAQItems(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for empty FAQ list, got %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil empty FAQ list")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty FAQ list, got %#v", items)
	}
}

// TestFAQRepositoryReturnsScanError verifies FAQ scan error propagation.
func TestFAQRepositoryReturnsScanError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{
		"question",
		"answer",
		"sort_order",
	}).AddRow("如何預約看房？", "請透過公開聯絡方式與我們確認可預約時段。", "not-an-int")
	mock.ExpectQuery(regexp.QuoteMeta(faqItemsQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewFAQRepository(db).ListFAQItems(context.Background())
	if err == nil {
		t.Fatal("expected scan error")
	}
	if items != nil {
		t.Fatalf("expected nil FAQ items, got %#v", items)
	}
}

// TestFAQRepositoryReturnsRowsError verifies FAQ row iteration error propagation.
func TestFAQRepositoryReturnsRowsError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	expectedErr := errors.New("row failed")
	rows := sqlmock.NewRows([]string{
		"question",
		"answer",
		"sort_order",
	}).
		AddRow("如何預約看房？", "請透過公開聯絡方式與我們確認可預約時段。", 10).
		RowError(0, expectedErr)
	mock.ExpectQuery(regexp.QuoteMeta(faqItemsQuery)).WithArgs().WillReturnRows(rows)

	items, err := NewFAQRepository(db).ListFAQItems(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected rows error, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil FAQ items, got %#v", items)
	}
}

// TestFAQRepositoryReturnsDatabaseError verifies FAQ query error propagation.
func TestFAQRepositoryReturnsDatabaseError(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	expectedErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta(faqItemsQuery)).WithArgs().WillReturnError(expectedErr)

	items, err := NewFAQRepository(db).ListFAQItems(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil FAQ items, got %#v", items)
	}
}
