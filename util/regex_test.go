package util

import "testing"

func TestValidateLetterNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"abc", true},
		{"ABC123", true},
		{"hello world 42", true},
		{"", false},
		{"abc!", false},
		{"abc_def", false},
		{"a-b", false},
		{"select * from", false},
	}
	for _, tt := range tests {
		if got := ValidateLetterNumber(tt.input); got != tt.want {
			t.Errorf("ValidateLetterNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"user@example.com", true},
		{"first.last@sub.example.co.id", true},
		{"a-b_c@ex-ample.org", true},
		{"", false},
		{"plainaddress", false},
		{"@example.com", false},
		{"user@", false},
		{"user@example", false},
		{"user@example.c", false},
		{"user @example.com", false},
	}
	for _, tt := range tests {
		if got := ValidateEmail(tt.input); got != tt.want {
			t.Errorf("ValidateEmail(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
