package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/errs"
	"github.com/labstack/echo/v4"
)

type Validatable interface {
	Validate() error
}

type CustomValidationError struct {
	Field   string
	Message string
}

type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return "Validation failed"
}

func BindAndValidate(c echo.Context, payload Validatable) error {
	if err := c.Bind(payload); err != nil {
		message := strings.Split(strings.Split(err.Error(), ",")[1], "message=")[1]
		return errs.NewBadRequestError(message, false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); fieldErrors != nil {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}

	return nil
}

func validateStruct(v Validatable) (string, []errs.FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []errs.FieldError) {
	var fieldErrors []errs.FieldError
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		customValidationErrors := err.(CustomValidationErrors)
		for _, err := range customValidationErrors {
			fieldErrors = append(fieldErrors, errs.FieldError{
				Field: err.Field,
				Error: err.Message,
			})
		}
	}

	for _, err := range validationErrors {
		fieldErrors = append(fieldErrors, errs.FieldError{
			Field: strings.ToLower(err.Field()),
			Error: validationMessage(err),
		})
	}

	return "Validation failed", fieldErrors
}

func validationMessage(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return "is required"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", err.Param())
		}
		return fmt.Sprintf("must be at least %s", err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("must not exceed %s characters", err.Param())
		}
		return fmt.Sprintf("must not exceed %s", err.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())
	case "email":
		return "must be a valid email address"
	case "e164":
		return "must be a valid phone number with country code"
	case "uuid":
		return "must be a valid UUID"
	case "uuidList":
		return "must be a comma-separated list of valid UUIDs"
	case "dive":
		return "some items are invalid"
	default:
		if err.Param() != "" {
			return fmt.Sprintf("%s: %s:%s", field, err.Tag(), err.Param())
		}
		return fmt.Sprintf("%s: %s", field, err.Tag())
	}
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}
