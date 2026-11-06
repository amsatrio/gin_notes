package request

import (
	"encoding/json"
	"testing"
)

func TestFilterMatchMode_String(t *testing.T) {
	tests := map[FilterMatchMode]string{
		CONTAINS:     "CONTAINS",
		BETWEEN:      "BETWEEN",
		EQUALS:       "EQUALS",
		NOT:          "NOT",
		LESS_THAN:    "LESS_THAN",
		GREATER_THAN: "GREATER_THAN",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFilterDataType_String(t *testing.T) {
	tests := map[FilterDataType]string{
		TEXT:    "TEXT",
		NUMBER:  "NUMBER",
		DATE:    "DATE",
		BOOLEAN: "BOOLEAN",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFilterMode_String(t *testing.T) {
	tests := map[FilterMode]string{
		AND: "AND",
		OR:  "OR",
	}
	for in, want := range tests {
		if got := in.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestFilter_JSON(t *testing.T) {
	f := Filter{Id: "fullName", Value: "Adi", MatchMode: CONTAINS, DataType: TEXT, Mode: AND}
	b, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"id":"fullName","value":"Adi","matchMode":"CONTAINS","dataType":"TEXT","mode":"AND"}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	var back Filter
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if back != f {
		t.Errorf("round trip mismatch: %+v vs %+v", back, f)
	}
}

func TestSort_JSON(t *testing.T) {
	b, err := json.Marshal(Sort{Id: "fullName", Desc: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"id":"fullName","desc":true}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestRequestAuth_JSON(t *testing.T) {
	b, err := json.Marshal(RequestAuth{Username: "a@b.com", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"username":"a@b.com","password":"secret"}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestRequestRefreshToken_JSON(t *testing.T) {
	b, err := json.Marshal(RequestRefreshToken{Token: "t", RefreshToken: "r"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"token":"t","refresh_token":"r"}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}
