package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
)

type publicErrorCode string

const (
	errorCodeValidationFailed   publicErrorCode = "VALIDATION_FAILED"
	errorCodeNotFound           publicErrorCode = "NOT_FOUND"
	errorCodeServiceUnavailable publicErrorCode = "SERVICE_UNAVAILABLE"
	errorCodeInternalError      publicErrorCode = "INTERNAL_ERROR"
)

type publicErrorResponse struct {
	Error     publicError `json:"error"`
	RequestID string      `json:"request_id"`
}

type publicError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeValidationError(c *gin.Context) {
	writePublicError(c, nethttp.StatusBadRequest, errorCodeValidationFailed)
}

func writeNotFoundError(c *gin.Context) {
	writePublicError(c, nethttp.StatusNotFound, errorCodeNotFound)
}

func writeServiceUnavailableError(c *gin.Context) {
	writePublicError(c, nethttp.StatusServiceUnavailable, errorCodeServiceUnavailable)
}

func writeInternalError(c *gin.Context) {
	writePublicError(c, nethttp.StatusInternalServerError, errorCodeInternalError)
}

func writePublicError(c *gin.Context, status int, code publicErrorCode) {
	c.JSON(status, publicErrorResponse{
		Error: publicError{
			Code:    string(code),
			Message: publicErrorMessage(code),
		},
		RequestID: requestIDFromContext(c),
	})
}

func publicErrorMessage(code publicErrorCode) string {
	switch code {
	case errorCodeValidationFailed:
		return "Invalid request."
	case errorCodeNotFound:
		return "Not found."
	case errorCodeServiceUnavailable:
		return "Service unavailable."
	case errorCodeInternalError:
		return "Internal server error."
	default:
		return "Internal server error."
	}
}

func requestIDFromContext(c *gin.Context) string {
	value, ok := c.Get(requestIDContext)
	if !ok {
		return ""
	}

	requestID, ok := value.(string)
	if !ok {
		return ""
	}
	return requestID
}
