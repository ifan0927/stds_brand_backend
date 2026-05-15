package http

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifan0927/stds_brand_backend/internal/application"
)

type fakeHealthChecker struct {
	err error
}

// Check returns the configured fake health check error.
func (checker fakeHealthChecker) Check(context.Context) error {
	return checker.err
}

type fakeBrandProfileService struct {
	profile *application.BrandProfile
	err     error
}

// GetBrandProfile returns fake brand profile service results.
func (service fakeBrandProfileService) GetBrandProfile(context.Context) (*application.BrandProfile, error) {
	return service.profile, service.err
}

type fakeFAQService struct {
	items []application.FAQItem
	err   error
}

// ListFAQItems returns fake FAQ service results.
func (service fakeFAQService) ListFAQItems(context.Context) ([]application.FAQItem, error) {
	return service.items, service.err
}

type fakePropertyAvailabilityService struct {
	items []application.PropertyAvailability
	err   error
}

// ListPropertyAvailability returns fake property availability service results.
func (service fakePropertyAvailabilityService) ListPropertyAvailability(context.Context) ([]application.PropertyAvailability, error) {
	return service.items, service.err
}

// TestHealthReturnsOKWhenDatabaseIsReady verifies the healthy response path.
func TestHealthReturnsOKWhenDatabaseIsReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

// TestHealthReturnsSafeUnavailableWhenDatabaseFails verifies the safe unavailable health response.
func TestHealthReturnsSafeUnavailableWhenDatabaseFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{err: errors.New("postgres://user:password@localhost/db failed")}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %q", body.Error.Code)
	}
	if body.Error.Message != "Service unavailable." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
	if containsAny(recorder.Body.String(), []string{"postgres://", "password", "localhost/db"}) {
		t.Fatalf("response leaked database details: %s", recorder.Body.String())
	}
}

// TestUnmatchedRouteReturnsPublicNotFoundError verifies the public not-found error shape.
func TestUnmatchedRouteReturnsPublicNotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", body.Error.Code)
	}
	if body.Error.Message != "Not found." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
}

// TestPublicErrorHelperWritesValidationErrorShape verifies the validation error response shape.
func TestPublicErrorHelperWritesValidationErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})
	router.GET("/test-validation", func(c *gin.Context) {
		writeValidationError(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test-validation", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("expected VALIDATION_FAILED, got %q", body.Error.Code)
	}
	if body.Error.Message != "Invalid request." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
}

// TestPublicErrorHelperWritesSafeInternalErrorShape verifies the internal error response shape.
func TestPublicErrorHelperWritesSafeInternalErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})
	router.GET("/test-internal", func(c *gin.Context) {
		_ = errors.New("sql: password=secret failed")
		writeInternalError(c)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test-internal", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", body.Error.Code)
	}
	if body.Error.Message != "Internal server error." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
	if containsAny(recorder.Body.String(), []string{"sql:", "password=secret"}) {
		t.Fatalf("response leaked internal details: %s", recorder.Body.String())
	}
}

// TestRequestIDMiddlewarePreservesIncomingRequestID verifies accepted request ID propagation.
func TestRequestIDMiddlewarePreservesIncomingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(requestIDHeader); got != "brand-request-123" {
		t.Fatalf("expected response request ID %q, got %q", "brand-request-123", got)
	}
}

// TestRequestIDMiddlewareGeneratesRequestIDWhenMissing verifies generated request IDs.
func TestRequestIDMiddlewareGeneratesRequestIDWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, request)

	requestID := recorder.Header().Get(requestIDHeader)
	if len(requestID) != 32 {
		t.Fatalf("expected generated request ID length 32, got %d: %q", len(requestID), requestID)
	}
	if _, err := hex.DecodeString(requestID); err != nil {
		t.Fatalf("expected generated request ID to be hex: %v", err)
	}
}

// TestRequestIDMiddlewareRegeneratesUnsafeIncomingRequestID verifies unsafe request ID replacement.
func TestRequestIDMiddlewareRegeneratesUnsafeIncomingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set(requestIDHeader, "unsafe\nrequest")

	router.ServeHTTP(recorder, request)

	requestID := recorder.Header().Get(requestIDHeader)
	if requestID == "" {
		t.Fatal("expected generated request ID")
	}
	if requestID == "unsafe\nrequest" {
		t.Fatal("expected unsafe incoming request ID to be replaced")
	}
}

// TestStructuredLoggingMiddlewareWritesSafeRequestFields verifies structured request log fields.
func TestStructuredLoggingMiddlewareWritesSafeRequestFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		Logger:        logger,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health?token=secret", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if entry["request_id"] != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %#v", entry["request_id"])
	}
	if entry["method"] != http.MethodGet {
		t.Fatalf("expected method GET, got %#v", entry["method"])
	}
	if entry["path"] != "/health" {
		t.Fatalf("expected safe path /health, got %#v", entry["path"])
	}
	if entry["status"] != float64(http.StatusOK) {
		t.Fatalf("expected status %d, got %#v", http.StatusOK, entry["status"])
	}
	if _, ok := entry["latency"].(float64); !ok {
		t.Fatalf("expected numeric latency, got %#v", entry["latency"])
	}
	if containsAny(logs.String(), []string{"token=secret", "postgres://", "password"}) {
		t.Fatalf("log leaked sensitive data: %s", logs.String())
	}
}

// TestBrandProfileReturnsApprovedProfile verifies the approved brand profile response.
func TestBrandProfileReturnsApprovedProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	phone := "02-1234-5678"
	email := "hello@example.com"
	address := "台北市中正區範例路 1 號"
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		BrandProfileService: fakeBrandProfileService{profile: &application.BrandProfile{
			BrandName:      "STDS",
			ContactPhone:   &phone,
			ContactEmail:   &email,
			ContactAddress: &address,
			UpdatedAt:      time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC),
		}},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/profile", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body brandProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Profile == nil {
		t.Fatal("expected profile response")
	}
	if body.Profile.BrandName != "STDS" {
		t.Fatalf("expected brand name STDS, got %q", body.Profile.BrandName)
	}
	if body.Profile.ContactPhone == nil || *body.Profile.ContactPhone != phone {
		t.Fatalf("expected contact phone %q, got %#v", phone, body.Profile.ContactPhone)
	}
	if body.Profile.ContactEmail == nil || *body.Profile.ContactEmail != email {
		t.Fatalf("expected contact email %q, got %#v", email, body.Profile.ContactEmail)
	}
	if body.Profile.ContactAddress == nil || *body.Profile.ContactAddress != address {
		t.Fatalf("expected contact address %q, got %#v", address, body.Profile.ContactAddress)
	}
	if body.Profile.UpdatedAt != "2026-05-15T12:00:00Z" {
		t.Fatalf("expected updated_at RFC3339, got %q", body.Profile.UpdatedAt)
	}
}

// TestBrandProfileReturnsNullWhenNoApprovedProfileExists verifies the empty brand profile response.
func TestBrandProfileReturnsNullWhenNoApprovedProfileExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker:       fakeHealthChecker{},
		BrandProfileService: fakeBrandProfileService{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/profile", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body brandProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Profile != nil {
		t.Fatalf("expected null profile, got %#v", body.Profile)
	}
}

// TestBrandProfileReturnsSafeUnavailableWhenRepositoryFails verifies the brand profile failure response.
func TestBrandProfileReturnsSafeUnavailableWhenRepositoryFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker:       fakeHealthChecker{},
		BrandProfileService: fakeBrandProfileService{err: errors.New("sql: password=secret failed")},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/profile", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %q", body.Error.Code)
	}
	if body.Error.Message != "Service unavailable." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
	if containsAny(recorder.Body.String(), []string{"sql:", "password=secret"}) {
		t.Fatalf("response leaked internal details: %s", recorder.Body.String())
	}
}

// TestFAQsReturnApprovedItems verifies the approved FAQ response.
func TestFAQsReturnApprovedItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		FAQService: fakeFAQService{items: []application.FAQItem{
			{
				Question:  "如何預約看房？",
				Answer:    "請透過公開聯絡方式與我們確認可預約時段。",
				SortOrder: 10,
			},
			{
				Question:  "是否可以線上詢問？",
				Answer:    "可以，請使用公開聯絡方式。",
				SortOrder: 20,
			},
		}},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/faqs", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body faqResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("expected 2 FAQ items, got %d", len(body.Items))
	}
	if body.Items[0].Question != "如何預約看房？" {
		t.Fatalf("expected first question, got %q", body.Items[0].Question)
	}
	if body.Items[0].Answer != "請透過公開聯絡方式與我們確認可預約時段。" {
		t.Fatalf("expected first answer, got %q", body.Items[0].Answer)
	}
	if body.Items[0].SortOrder != 10 {
		t.Fatalf("expected first sort order 10, got %d", body.Items[0].SortOrder)
	}
}

// TestFAQsReturnEmptyListWhenNoApprovedItemsExist verifies the empty FAQ response.
func TestFAQsReturnEmptyListWhenNoApprovedItemsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		FAQService:    fakeFAQService{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/faqs", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"items":[]`) {
		t.Fatalf("expected empty items list, got %s", recorder.Body.String())
	}

	var body faqResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Items) != 0 {
		t.Fatalf("expected empty FAQ list, got %#v", body.Items)
	}
}

// TestFAQsReturnSafeUnavailableWhenRepositoryFails verifies the FAQ failure response.
func TestFAQsReturnSafeUnavailableWhenRepositoryFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		FAQService:    fakeFAQService{err: errors.New("sql: password=secret failed")},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/brand/faqs", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %q", body.Error.Code)
	}
	if body.Error.Message != "Service unavailable." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
	if containsAny(recorder.Body.String(), []string{"sql:", "password=secret"}) {
		t.Fatalf("response leaked internal details: %s", recorder.Body.String())
	}
}

// TestPropertyAvailabilityReturnsApprovedItems verifies the approved property availability response.
func TestPropertyAvailabilityReturnsApprovedItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker: fakeHealthChecker{},
		PropertyAvailabilityService: fakePropertyAvailabilityService{items: []application.PropertyAvailability{
			{
				PropertyID:         "00000000-0000-0000-0000-000000000000",
				PropertyPublicName: "北車館",
				Address:            "台北市中正區範例路 1 號",
				HasVacantRoom:      true,
			},
			{
				PropertyID:         "11111111-1111-1111-1111-111111111111",
				PropertyPublicName: "南港館",
				Address:            "台北市南港區範例路 2 號",
				HasVacantRoom:      false,
			},
		}},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/properties/availability", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var body propertyAvailabilityResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("expected 2 property availability items, got %d", len(body.Items))
	}
	if body.Items[0].PropertyID != "00000000-0000-0000-0000-000000000000" {
		t.Fatalf("expected first property ID, got %q", body.Items[0].PropertyID)
	}
	if body.Items[0].PropertyPublicName != "北車館" {
		t.Fatalf("expected first property public name, got %q", body.Items[0].PropertyPublicName)
	}
	if body.Items[0].Address != "台北市中正區範例路 1 號" {
		t.Fatalf("expected first address, got %q", body.Items[0].Address)
	}
	if !body.Items[0].HasVacantRoom {
		t.Fatal("expected first property to have vacant room")
	}
}

// TestPropertyAvailabilityReturnsEmptyListWhenNoApprovedItemsExist verifies the empty property availability response.
func TestPropertyAvailabilityReturnsEmptyListWhenNoApprovedItemsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker:               fakeHealthChecker{},
		PropertyAvailabilityService: fakePropertyAvailabilityService{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/properties/availability", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"items":[]`) {
		t.Fatalf("expected empty items list, got %s", recorder.Body.String())
	}

	var body propertyAvailabilityResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(body.Items) != 0 {
		t.Fatalf("expected empty property availability list, got %#v", body.Items)
	}
}

// TestPropertyAvailabilityReturnsSafeUnavailableWhenRepositoryFails verifies the property availability failure response.
func TestPropertyAvailabilityReturnsSafeUnavailableWhenRepositoryFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{
		HealthChecker:               fakeHealthChecker{},
		PropertyAvailabilityService: fakePropertyAvailabilityService{err: errors.New("sql: password=secret failed")},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/properties/availability", nil)
	request.Header.Set(requestIDHeader, "brand-request-123")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var body publicErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Error.Code != "SERVICE_UNAVAILABLE" {
		t.Fatalf("expected SERVICE_UNAVAILABLE, got %q", body.Error.Code)
	}
	if body.Error.Message != "Service unavailable." {
		t.Fatalf("expected safe message, got %q", body.Error.Message)
	}
	if body.RequestID != "brand-request-123" {
		t.Fatalf("expected request_id brand-request-123, got %q", body.RequestID)
	}
	if containsAny(recorder.Body.String(), []string{"sql:", "password=secret"}) {
		t.Fatalf("response leaked internal details: %s", recorder.Body.String())
	}
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
