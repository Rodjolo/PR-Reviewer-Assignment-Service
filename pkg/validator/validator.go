// Package validator provides request validation functionality using go-playground/validator.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct based on validation tags.
func Validate(data interface{}) error {
	return validate.Struct(data)
}

// FormatValidationErrors formats validation errors into a human-readable string.
func FormatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range validationErrors {
			message := formatFieldError(fieldError)
			messages = append(messages, message)
		}
		return strings.Join(messages, "; ")
	}
	return err.Error()
}

func formatFieldError(fieldError validator.FieldError) string {
	field := strings.ToLower(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fieldError.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fieldError.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fieldError.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fieldError.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fieldError.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
