package http

import (
	"log"
	"net/http"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// respondWithSuccess sends a successful JSON response
func respondWithSuccess(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, dto.NewSuccessResponse(data))
}

// respondWithError sends an error JSON response
func respondWithError(c *gin.Context, statusCode int, code, message string, details map[string]string) {
	c.JSON(statusCode, dto.NewErrorResponse[any](code, message, details))
}

// handleDomainError converts domain errors to HTTP responses
func handleDomainError(c *gin.Context, err error) {
	domainErr := errors.GetDomainError(err)
	if domainErr == nil {
		// Log full error details for unexpected errors
		requestID, _ := c.Get(RequestIDKey)
		log.Printf("[ERROR] [%v] Unexpected error: %+v", requestID, err)

		// Return generic message to client
		respondWithError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", nil)
		return
	}

	statusCode := mapErrorToHTTPStatus(domainErr.Code)

	// Log internal errors with full details
	if statusCode == http.StatusInternalServerError {
		requestID, _ := c.Get(RequestIDKey)
		log.Printf("[ERROR] [%v] Domain error: code=%s, message=%s, cause=%+v",
			requestID, domainErr.Code, domainErr.Message, domainErr.Cause)
	}

	respondWithError(c, statusCode, domainErr.Code, domainErr.Message, nil)
}

// mapErrorToHTTPStatus maps domain error codes to HTTP status codes
func mapErrorToHTTPStatus(code string) int {
	switch code {
	// Not Found errors (404)
	case "SESSION_NOT_FOUND", "MESSAGE_NOT_FOUND", "NOT_FOUND", "CONTACT_NOT_FOUND", "CHAT_NOT_FOUND":
		return http.StatusNotFound

	// Conflict errors (409)
	case "SESSION_EXISTS", "DUPLICATE":
		return http.StatusConflict

	// Bad Request errors (400)
	case "INVALID_PHONE", "VALIDATION_FAILED", "INVALID_INPUT", "EMPTY_CONTENT", "INVALID_MESSAGE_TYPE",
		"INVALID_MEDIA_SIZE", "MEDIA_TOO_LARGE", "UNSUPPORTED_MEDIA_TYPE", "INVALID_MIME_TYPE", "UNSUPPORTED_MIME_TYPE",
		"DISCONNECTED", "SESSION_INVALID", "INVALID_EMOJI", "INVALID_REACTION", "INVALID_RECEIPT_TYPE",
		"INVALID_PRESENCE_STATE", "INVALID_JID", "INVALID_STATUS":
		return http.StatusBadRequest

	// Timeout errors (408)
	case "QR_TIMEOUT", "AUTH_TIMEOUT":
		return http.StatusRequestTimeout

	// Service Unavailable errors (503)
	case "CONNECTION_FAILED", "RECONNECT_FAILED", "CIRCUIT_OPEN", "WHATSAPP_UNAVAILABLE":
		return http.StatusServiceUnavailable

	// Internal Server errors (500)
	case "DATABASE_ERROR", "INTERNAL_ERROR", "QR_GENERATION_FAILED", "AUTH_FAILED",
		"MESSAGE_SEND_FAILED", "MEDIA_DOWNLOAD_FAILED", "MEDIA_UPLOAD_FAILED",
		"REACTION_SEND_FAILED", "RECEIPT_SEND_FAILED", "PRESENCE_SEND_FAILED",
		"CONFIG_MISSING", "CONFIG_INVALID", "WHATSAPP_ERROR":
		return http.StatusInternalServerError

	// Unauthorized errors (401)
	case "UNAUTHORIZED", "MISSING_API_KEY", "INVALID_API_KEY":
		return http.StatusUnauthorized

	// Forbidden errors (403)
	case "FORBIDDEN", "INSUFFICIENT_PERMISSIONS":
		return http.StatusForbidden

	// Rate Limit errors (429)
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests

	// Default to Internal Server Error for unknown codes
	default:
		return http.StatusInternalServerError
	}
}
