package response

import (
	"encoding/json"
	"testing"
	"time"
)

func TestResponse_Marshal(t *testing.T) {
	r := Response{
		Path:      "/v1/m_biodata",
		Timestamp: JSONTime{Time: time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)},
		Status:    200,
		Message:   "success",
		Data:      map[string]string{"k": "v"},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"path":"/v1/m_biodata","timestamp":"2024-02-16 10:33:10","status":200,"message":"success","data":{"k":"v"}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}

	var back Response
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !back.Timestamp.Equal(r.Timestamp.Time) || back.Status != 200 {
		t.Errorf("round trip mismatch: %+v", back)
	}
}

func TestResponseTimestamp_Marshal(t *testing.T) {
	ts := time.Date(2024, 2, 16, 10, 33, 10, 0, time.UTC)
	r := ResponseTimestamp{Path: "/p", Timestamp: ts, Status: 201, Message: "created", Data: nil}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var back ResponseTimestamp
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !back.Timestamp.Equal(ts) || back.Path != "/p" || back.Status != 201 {
		t.Errorf("round trip mismatch: %+v", back)
	}
}

func TestPage_Marshal(t *testing.T) {
	p := Page{
		Content:          []string{"a", "b"},
		Pageable:         Pageable{Offset: 0, PageNumber: 0, PageSize: 5, Paged: true, Sort: Sort{Sorted: true}},
		Sort:             Sort{Empty: true},
		TotalPages:       1,
		TotalElements:    2,
		Size:             5,
		Number:           0,
		NumberOfElements: 2,
		Last:             true,
		First:            true,
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	for _, k := range []string{"content", "pageable", "sort", "totalPages", "totalElements", "size", "number", "numberOfElements", "last", "first", "empty"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing key %q in %s", k, b)
		}
	}
}

func TestPageable_Marshal(t *testing.T) {
	b, err := json.Marshal(Pageable{PageSize: 5, Paged: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"offset":0,"pageNumber":0,"pageSize":5,"paged":true,"unPaged":false,"sort":{"empty":false,"unsorted":false,"sorted":false}}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestSort_Marshal(t *testing.T) {
	b, err := json.Marshal(Sort{Empty: true, Unsorted: false, Sorted: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"empty":true,"unsorted":false,"sorted":true}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestResponseAuth_Marshal(t *testing.T) {
	b, err := json.Marshal(ResponseAuth{Token: "t", RefreshToken: "r", ExpiredIn: "200"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"token":"t","refresh_token":"r","expired_in":"200"}`
	if string(b) != want {
		t.Errorf("got %s, want %s", b, want)
	}
}
