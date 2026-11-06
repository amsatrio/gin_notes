package response

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONTime_MarshalJSON(t *testing.T) {
	jt := JSONTime{Time: time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)}
	b, err := jt.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := `"2024-02-16 10:33:10"`; string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	// via json.Marshal
	b, err = json.Marshal(jt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != `"2024-02-16 10:33:10"` {
		t.Errorf("unexpected output %s", b)
	}
}

func TestJSONTime_UnmarshalJSON(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var jt JSONTime
		if err := jt.UnmarshalJSON([]byte(`"2024-02-16 10:33:10"`)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)
		if !jt.Time.Equal(want) {
			t.Errorf("got %v, want %v", jt.Time, want)
		}
	})

	t.Run("error", func(t *testing.T) {
		var jt JSONTime
		if err := jt.UnmarshalJSON([]byte(`"not-a-time"`)); err == nil {
			t.Fatal("expected error, got nil")
		}
		if !jt.IsZero() {
			t.Errorf("expected zero time on error, got %v", jt.Time)
		}
	})
}

func TestJSONTime_Scan(t *testing.T) {
	now := time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)

	tests := []struct {
		name    string
		input   interface{}
		want    time.Time
		wantErr bool
	}{
		{"nil", nil, time.Time{}, false},
		{"time.Time", now, now, false},
		{"bytes ok", []byte(`"2024-02-16 10:33:10"`), now, false},
		{"bytes invalid", []byte(`"bad"`), time.Time{}, true},
		{"string ok", `"2024-02-16 10:33:10"`, now, false},
		{"string invalid", `"bad"`, time.Time{}, true},
		{"unsupported", 12345, time.Time{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var jt JSONTime
			err := jt.Scan(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !jt.Time.Equal(tc.want) {
				t.Errorf("got %v, want %v", jt.Time, tc.want)
			}
		})
	}
}

func TestJSONTime_Value(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		v, err := JSONTime{}.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Errorf("got %v, want nil", v)
		}
	})

	t.Run("non-zero", func(t *testing.T) {
		now := time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)
		v, err := JSONTime{Time: now}.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, ok := v.(time.Time)
		if !ok {
			t.Fatalf("expected time.Time, got %T", v)
		}
		if !got.Equal(now) {
			t.Errorf("got %v, want %v", got, now)
		}
	})
}
