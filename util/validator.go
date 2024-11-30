package util

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

func ValidateError(err error) (map[string]string, error) {
	Log("INFO", "util", "Validate", "error: "+err.Error())
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		Log("INFO", "util", "Validate", "success")
		out := make(map[string]string, len(ve))
		for _, fe := range ve {
			out[fe.Field()] = getErrorMessage(fe)
			Log("INFO", "util", "Validate", "field "+fe.Field())
		}

		return out, nil
	}
	return nil, nil
}

func getErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "this field is mandatory"
	case "min":
		return "minimum " + fieldError.Param() + " character(s)"
	case "max":
		return "maximum " + fieldError.Param() + " character(s)"
	case "alphanum":
		return "should be alphanumeric letter"
	case "uppercase":
		return "should be uppercase letter"
	case "lowercase":
		return "should be lowercase letter"
	case "contains":
		return "should be contains " + fieldError.Param() + " character(s)"
	case "endswith":
		return "should be ends with " + fieldError.Param() + " character(s)"
	case "startwith":
		return "should be start with " + fieldError.Param() + " character(s)"
	case "gte":
		return "should be greater than " + fieldError.Param()
	case "lte":
		return "should be less than " + fieldError.Param()
	case "gtfield":
		return "should be greater than field " + fieldError.Param()
	case "ltfield":
		return "should be less than field " + fieldError.Param()
	default:
		return fieldError.Tag()
	}
}
