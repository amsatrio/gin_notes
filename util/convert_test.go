package util

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/amsatrio/gin_notes/model/response"
)

func TestCamelCaseToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"no capitals", "name", "name"},
		{"no capitals with digits", "field123", "field123"},
		{"single capital", "fullName", "full_name"},
		{"already lowercase snake", "full_name", "full_name"},
		{"only capitals is lowered", "ABC", "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CamelCaseToSnakeCase(tt.input); got != tt.want {
				t.Errorf("CamelCaseToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCountCapitalLetters(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"abc", 0},
		{"aBc", 1},
		{"ABC", 3},
		{"ÀÉ", 2},
	}
	for _, tt := range tests {
		if got := countCapitalLetters(tt.input); got != tt.want {
			t.Errorf("countCapitalLetters(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestResponseToByte(t *testing.T) {
	ts := time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)
	res := response.Response{
		Path:      "/v1/test",
		Timestamp: response.JSONTime{Time: ts},
		Status:    200,
		Message:   "success",
		Data:      map[string]string{"k": "v"},
	}

	b, err := ResponseToByte(res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		t.Fatalf("expected json encoder output ending with newline, got %q", string(b))
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded["path"] != "/v1/test" {
		t.Errorf("path = %v", decoded["path"])
	}
	if decoded["status"] != float64(200) {
		t.Errorf("status = %v", decoded["status"])
	}
	if decoded["message"] != "success" {
		t.Errorf("message = %v", decoded["message"])
	}
	if decoded["timestamp"] != "2024-02-16 10:33:10" {
		t.Errorf("timestamp = %v", decoded["timestamp"])
	}
}

func TestResponseToByte_EncodeError(t *testing.T) {
	res := response.Response{Data: make(chan int)}
	b, err := ResponseToByte(res)
	if err == nil {
		t.Fatal("expected error for unencodable data")
	}
	if b != nil {
		t.Errorf("expected nil bytes, got %q", string(b))
	}
}
