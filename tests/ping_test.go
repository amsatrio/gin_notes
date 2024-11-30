package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestPing(t *testing.T) {
	var router = SetUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v2/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGetMBiodata(t *testing.T) {
	var router = SetUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/m_biodata/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
