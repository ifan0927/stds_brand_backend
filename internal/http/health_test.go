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

func (checker fakeHealthChecker) Check(context.Context) error {
	return checker.err
}

type fakeBrandProfileService struct {
	profile *application.BrandProfile
	err     error
}

func (service fakeBrandProfileService) GetBrandProfile(context.Context) (*application.BrandProfile, error) {
	return service.profile, service.err
}

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

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
