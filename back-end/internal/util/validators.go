package util

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var validate = validator.New()

func init() {
	// Use json tag names in error messages instead of Go struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if value, ok := field.Interface().(decimal.Decimal); ok {
			f, _ := value.Float64()
			return f
		}
		return nil
	}, decimal.Decimal{})
}

func ValidateStruct[T any](s *T) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return map[string]string{"error": err.Error()}
	}

	fields := make(map[string]string)
	for _, e := range ve {
		fields[e.Field()] = fieldError(e)
	}
	return fields
}

func fieldError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must me at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "eqfield":
		return fmt.Sprintf("must match %s", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "uuid4":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed validation on '%s'", e.Tag())
	}

}
