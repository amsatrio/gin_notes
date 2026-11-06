package util

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

type validatorSample struct {
	Required  string `validate:"required"`
	Min       string `validate:"min=3"`
	Max       string `validate:"max=2"`
	Alphanum  string `validate:"alphanum"`
	Upper     string `validate:"uppercase"`
	Lower     string `validate:"lowercase"`
	Contains  string `validate:"contains=xyz"`
	EndsWith  string `validate:"endswith=end"`
	StartWith string `validate:"startswith=beg"`
	Gte       int    `validate:"gte=10"`
	Lte       int    `validate:"lte=5"`
	GtField   int    `validate:"gtfield=Gte"`
	LtField   int    `validate:"ltfield=Lte"`
	Email     string `validate:"email"`
}

func TestValidateError_AllTags(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")

	s := validatorSample{
		Required:  "",
		Min:       "a",
		Max:       "abcdef",
		Alphanum:  "!!",
		Upper:     "abc",
		Lower:     "ABC",
		Contains:  "abc",
		EndsWith:  "abc",
		StartWith: "abc",
		Gte:       1,
		Lte:       100,
		GtField:   0,
		LtField:   1000,
		Email:     "not-an-email",
	}

	err := validator.New().Struct(s)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	got, err := ValidateError(err)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[string]string{
		"Required":  "this field is mandatory",
		"Min":       "minimum 3 character(s)",
		"Max":       "maximum 2 character(s)",
		"Alphanum":  "should be alphanumeric letter",
		"Upper":     "should be uppercase letter",
		"Lower":     "should be lowercase letter",
		"Contains":  "should be contains xyz character(s)",
		"EndsWith":  "should be ends with end character(s)",
		"StartWith": "startswith", // handler only knows "startwith"; falls to default
		"Gte":       "should be greater than 10",
		"Lte":       "should be less than 5",
		"GtField":   "should be greater than field Gte",
		"LtField":   "should be less than field Lte",
		"Email":     "email", // default branch returns tag
	}

	if len(got) != len(want) {
		t.Errorf("got %d errors, want %d: %v", len(got), len(want), got)
	}
	for field, msg := range want {
		if got[field] != msg {
			t.Errorf("field %s: got %q, want %q", field, got[field], msg)
		}
	}
}

func TestValidateError_NotValidationErrors(t *testing.T) {
	got, err := ValidateError(errors.New("plain error"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil map, got %v", got)
	}
}

// fakeFieldError implements validator.FieldError for the "startwith" tag,
// which cannot be produced by the validator library itself.
type fakeFieldError struct {
	validator.FieldError
	tag   string
	param string
}

func (f fakeFieldError) Tag() string   { return f.tag }
func (f fakeFieldError) Param() string { return f.param }

func TestGetErrorMessage_StartWith(t *testing.T) {
	got := getErrorMessage(fakeFieldError{tag: "startwith", param: "abc"})
	want := "should be start with abc character(s)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGetErrorMessage_Default(t *testing.T) {
	got := getErrorMessage(fakeFieldError{tag: "custom"})
	if got != "custom" {
		t.Errorf("got %q, want %q", got, "custom")
	}
}
