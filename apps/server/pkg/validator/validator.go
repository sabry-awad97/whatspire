package validator

import (
	"regexp"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// e164Regex validates E.164 phone number format
var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// GetValidator returns the singleton validator instance
func GetValidator() *validator.Validate {
	once.Do(func() {
		validate = validator.New()
		// Register custom e164 validation
		_ = validate.RegisterValidation("e164", validateE164)
	})
	return validate
}

// validateE164 validates E.164 phone number format
func validateE164(fl validator.FieldLevel) bool {
	return e164Regex.MatchString(fl.Field().String())
}

// Validate validates a struct using the singleton validator
func Validate(s interface{}) error {
	return GetValidator().Struct(s)
}

// ValidationErrors extracts validation errors as a map
func ValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return map[string]string{"error": err.Error()}
	}

	errors := make(map[string]string)
	for _, e := range validationErrors {
		errors[e.Field()] = formatValidationError(e)
	}
	return errors
}

// formatValidationError formats a single validation error
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return "value is too short or too small"
	case "max":
		return "value is too long or too large"
	case "uuid":
		return "must be a valid UUID"
	case "e164":
		return "must be a valid E.164 phone number (e.g., +1234567890)"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return "must be one of the allowed values: " + e.Param()
	case "required_if":
		return "this field is required based on other field values"
	case "email":
		return "must be a valid email address"
	case "numeric":
		return "must be a numeric value"
	case "alpha":
		return "must contain only alphabetic characters"
	case "alphanum":
		return "must contain only alphanumeric characters"
	case "len":
		return "must be exactly " + e.Param() + " characters long"
	case "gt":
		return "must be greater than " + e.Param()
	case "gte":
		return "must be greater than or equal to " + e.Param()
	case "lt":
		return "must be less than " + e.Param()
	case "lte":
		return "must be less than or equal to " + e.Param()
	default:
		return "validation failed for " + e.Tag()
	}
}
