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

	"github.com/gin-gonic/gin"
)

type fakeHealthChecker struct {
	err error
}

func (checker fakeHealthChecker) Check(context.Context) error {
	return checker.err
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
	if containsAny(recorder.Body.String(), []string{"postgres://", "password", "localhost/db"}) {
		t.Fatalf("response leaked database details: %s", recorder.Body.String())
	}
}

func TestHealthzIsNotRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Dependencies{HealthChecker: fakeHealthChecker{}})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
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

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
