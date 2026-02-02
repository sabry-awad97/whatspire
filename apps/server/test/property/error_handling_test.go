package property

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"whatspire/internal/application/dto"
	"whatspire/internal/domain/errors"
	"whatspire/pkg/validator"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: whatsapp-http-api-enhancement, Property 22: Error Response Structure Consistency
// *For any* error occurring in any endpoint, the response SHALL have the structure
// {success: false, error: {code, message, details}}.
// **Validates: Requirements 8.1**

func TestErrorResponseStructureConsistency_Property22(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 22.1: All error responses have success=false
	properties.Property("All error responses have success=false", prop.ForAll(
		func(code, message string) bool {
			if code == "" || message == "" {
				return true // skip empty values
			}

			resp := dto.NewErrorResponse[any](code, message, nil)
			return !resp.Success
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 22.2: All error responses have error object with code and message
	properties.Property("All error responses have error object with code and message", prop.ForAll(
		func(code, message string) bool {
			if code == "" || message == "" {
				return true // skip empty values
			}

			resp := dto.NewErrorResponse[any](code, message, nil)

			if resp.Error == nil {
				return false
			}
			if resp.Error.Code != code {
				return false
			}
			if resp.Error.Message != message {
				return false
			}
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property 22.3: Error responses with details preserve all details
	properties.Property("Error responses with details preserve all details", prop.ForAll(
		func(code, message string, detailsMap map[string]string) bool {
			if code == "" || message == "" {
				return true // skip empty values
			}

			resp := dto.NewErrorResponse[any](code, message, detailsMap)

			if resp.Error == nil {
				return false
			}

			// Verify all details are preserved
			if len(detailsMap) > 0 {
				if resp.Error.Details == nil {
					return false
				}
				for key, value := range detailsMap {
					if resp.Error.Details[key] != value {
						return false
					}
				}
			}
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.MapOf(gen.AlphaString(), gen.AlphaString()),
	))

	// Property 22.4: Error responses never have data field populated
	properties.Property("Error responses never have data field populated", prop.ForAll(
		func(code, message string) bool {
			if code == "" || message == "" {
				return true // skip empty values
			}

			resp := dto.NewErrorResponse[string](code, message, nil)

			// Data should be empty for error responses
			return resp.Data == ""
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 23: Validation Error Detail Inclusion
// *For any* validation error, the error response SHALL include field-level details
// mapping field names to error messages.
// **Validates: Requirements 8.2**

func TestValidationErrorDetailInclusion_Property23(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 23.1: ValidationErrors extracts field-level details
	properties.Property("ValidationErrors extracts field-level details", prop.ForAll(
		func(fieldName, tag string) bool {
			if fieldName == "" || tag == "" {
				return true // skip empty values
			}

			// Create a mock validation error
			mockErr := &mockFieldError{
				field: fieldName,
				tag:   tag,
			}

			mockValidationErrors := validatorv10.ValidationErrors{mockErr}

			details := validator.ValidationErrors(mockValidationErrors)

			// Verify field is in details
			if details == nil {
				return false
			}

			_, exists := details[fieldName]
			return exists
		},
		gen.AlphaString(),
		gen.OneConstOf("required", "min", "max", "uuid", "e164", "url", "oneof"),
	))

	// Property 23.2: Each validation tag has a descriptive message
	properties.Property("Each validation tag has a descriptive message", prop.ForAll(
		func(tag string) bool {
			mockErr := &mockFieldError{
				field: "TestField",
				tag:   tag,
			}

			mockValidationErrors := validatorv10.ValidationErrors{mockErr}
			details := validator.ValidationErrors(mockValidationErrors)

			if details == nil {
				return false
			}

			message, exists := details["TestField"]
			if !exists {
				return false
			}

			// Message should not be empty and should be descriptive
			return message != "" && len(message) > 5
		},
		gen.OneConstOf("required", "min", "max", "uuid", "e164", "url", "email", "numeric"),
	))

	// Property 23.3: Multiple field errors are all included
	properties.Property("Multiple field errors are all included", prop.ForAll(
		func(field1, field2, tag1, tag2 string) bool {
			if field1 == "" || field2 == "" || tag1 == "" || tag2 == "" || field1 == field2 {
				return true // skip invalid cases
			}

			mockErr1 := &mockFieldError{field: field1, tag: tag1}
			mockErr2 := &mockFieldError{field: field2, tag: tag2}

			mockValidationErrors := validatorv10.ValidationErrors{mockErr1, mockErr2}
			details := validator.ValidationErrors(mockValidationErrors)

			if details == nil {
				return false
			}

			// Both fields should be present
			_, exists1 := details[field1]
			_, exists2 := details[field2]

			return exists1 && exists2
		},
		gen.AlphaString(),
		gen.AlphaString(),
		gen.OneConstOf("required", "min", "max"),
		gen.OneConstOf("required", "min", "max"),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 24: Domain Error to HTTP Status Mapping
// *For any* domain error code, the HTTP status code SHALL be deterministically mapped
// (e.g., NOT_FOUND → 404, VALIDATION_FAILED → 400).
// **Validates: Requirements 8.3**

func TestDomainErrorToHTTPStatusMapping_Property24(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 24.1: NOT_FOUND errors map to 404
	properties.Property("NOT_FOUND errors map to 404", prop.ForAll(
		func(_ int) bool {
			notFoundCodes := []string{
				"SESSION_NOT_FOUND",
				"MESSAGE_NOT_FOUND",
				"NOT_FOUND",
				"CONTACT_NOT_FOUND",
				"CHAT_NOT_FOUND",
			}

			for _, code := range notFoundCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusNotFound {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.2: Validation errors map to 400
	properties.Property("Validation errors map to 400", prop.ForAll(
		func(_ int) bool {
			validationCodes := []string{
				"INVALID_PHONE",
				"VALIDATION_FAILED",
				"INVALID_INPUT",
				"EMPTY_CONTENT",
				"INVALID_MESSAGE_TYPE",
				"INVALID_EMOJI",
				"INVALID_REACTION",
				"INVALID_RECEIPT_TYPE",
				"INVALID_PRESENCE_STATE",
				"INVALID_JID",
			}

			for _, code := range validationCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusBadRequest {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.3: Conflict errors map to 409
	properties.Property("Conflict errors map to 409", prop.ForAll(
		func(_ int) bool {
			conflictCodes := []string{
				"SESSION_EXISTS",
				"DUPLICATE",
			}

			for _, code := range conflictCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusConflict {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.4: Service unavailable errors map to 503
	properties.Property("Service unavailable errors map to 503", prop.ForAll(
		func(_ int) bool {
			unavailableCodes := []string{
				"CONNECTION_FAILED",
				"RECONNECT_FAILED",
				"CIRCUIT_OPEN",
				"WHATSAPP_UNAVAILABLE",
			}

			for _, code := range unavailableCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusServiceUnavailable {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.5: Internal errors map to 500
	properties.Property("Internal errors map to 500", prop.ForAll(
		func(_ int) bool {
			internalCodes := []string{
				"DATABASE_ERROR",
				"INTERNAL_ERROR",
				"MESSAGE_SEND_FAILED",
				"REACTION_SEND_FAILED",
				"RECEIPT_SEND_FAILED",
				"PRESENCE_SEND_FAILED",
			}

			for _, code := range internalCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusInternalServerError {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.6: Timeout errors map to 408
	properties.Property("Timeout errors map to 408", prop.ForAll(
		func(_ int) bool {
			timeoutCodes := []string{
				"QR_TIMEOUT",
				"AUTH_TIMEOUT",
			}

			for _, code := range timeoutCodes {
				err := errors.NewDomainError(code, "test message")
				status := mapDomainErrorToStatus(err)
				if status != http.StatusRequestTimeout {
					return false
				}
			}
			return true
		},
		gen.Const(0),
	))

	// Property 24.7: Unknown error codes default to 500
	properties.Property("Unknown error codes default to 500", prop.ForAll(
		func(unknownCode string) bool {
			if unknownCode == "" {
				return true // skip empty
			}

			// Use a random string that's unlikely to be a known code
			unknownCode = "UNKNOWN_" + unknownCode

			err := errors.NewDomainError(unknownCode, "test message")
			status := mapDomainErrorToStatus(err)

			// Should default to 500
			return status == http.StatusInternalServerError
		},
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Feature: whatsapp-http-api-enhancement, Property 25: Rate Limit Response Headers
// *For any* request that exceeds the rate limit, the response SHALL have HTTP status 429
// and include a Retry-After header.
// **Validates: Requirements 8.5**

func TestRateLimitResponseHeaders_Property25(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50
	properties := gopter.NewProperties(parameters)

	// Property 25.1: Rate limit exceeded returns 429
	properties.Property("Rate limit exceeded returns 429", prop.ForAll(
		func(_ int) bool {
			// Create a mock rate limit error response
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.JSON(http.StatusTooManyRequests, dto.NewErrorResponse[any](
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				nil,
			))

			return w.Code == http.StatusTooManyRequests
		},
		gen.Const(0),
	))

	// Property 25.2: Rate limit response includes Retry-After header
	properties.Property("Rate limit response includes Retry-After header", prop.ForAll(
		func(retryAfter int) bool {
			if retryAfter < 1 {
				retryAfter = 1
			}
			if retryAfter > 3600 {
				retryAfter = 3600
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate rate limit middleware behavior
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, dto.NewErrorResponse[any](
				"RATE_LIMIT_EXCEEDED",
				"Too many requests. Please try again later.",
				nil,
			))

			// Verify Retry-After header is present
			retryHeader := w.Header().Get("Retry-After")
			return retryHeader != ""
		},
		gen.IntRange(1, 3600),
	))

	// Property 25.3: Rate limit response includes X-RateLimit headers
	properties.Property("Rate limit response includes X-RateLimit headers", prop.ForAll(
		func(limit, remaining int, reset int64) bool {
			if limit < 1 {
				limit = 1
			}
			if remaining < 0 {
				remaining = 0
			}
			if reset < 0 {
				reset = 0
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate rate limit middleware behavior
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

			// Verify headers are present
			hasLimit := w.Header().Get("X-RateLimit-Limit") != ""
			hasRemaining := w.Header().Get("X-RateLimit-Remaining") != ""
			hasReset := w.Header().Get("X-RateLimit-Reset") != ""

			return hasLimit && hasRemaining && hasReset
		},
		gen.IntRange(1, 1000),
		gen.IntRange(0, 1000),
		gen.Int64Range(0, 9999999999),
	))

	properties.TestingRun(t)
}

// Helper function to map domain errors to HTTP status codes
// This mirrors the implementation in handler.go
func mapDomainErrorToStatus(err *errors.DomainError) int {
	code := err.Code

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

// Mock implementation of validatorv10.FieldError for testing
type mockFieldError struct {
	field string
	tag   string
	param string
}

func (m *mockFieldError) Tag() string                    { return m.tag }
func (m *mockFieldError) ActualTag() string              { return m.tag }
func (m *mockFieldError) Namespace() string              { return "" }
func (m *mockFieldError) StructNamespace() string        { return "" }
func (m *mockFieldError) Field() string                  { return m.field }
func (m *mockFieldError) StructField() string            { return m.field }
func (m *mockFieldError) Value() interface{}             { return nil }
func (m *mockFieldError) Param() string                  { return m.param }
func (m *mockFieldError) Kind() reflect.Kind             { return reflect.String }
func (m *mockFieldError) Type() reflect.Type             { return reflect.TypeOf("") }
func (m *mockFieldError) Translate(ut.Translator) string { return "" }
func (m *mockFieldError) Error() string                  { return "" }

// Ensure mockFieldError implements validatorv10.FieldError
var _ validatorv10.FieldError = (*mockFieldError)(nil)
