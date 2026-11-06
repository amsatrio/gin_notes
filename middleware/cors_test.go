package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware_Get(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := doRequest(r, http.MethodGet, "/ping")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("unexpected allow-origin: %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("unexpected allow-credentials: %q", got)
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected allow-headers to be set")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected allow-methods to be set")
	}
	if w.Body.String() != "pong" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	r := gin.New()
	called := false
	r.Use(CORSMiddleware())
	r.OPTIONS("/ping", func(c *gin.Context) { called = true })

	w := doRequest(r, http.MethodOptions, "/ping")
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if called {
		t.Error("handler must not be called for OPTIONS")
	}
}

func TestCORSGinMiddleware_AllowedOrigin(t *testing.T) {
	r := gin.New()
	r.Use(CORSGinMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("unexpected allow-origin: %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("unexpected allow-credentials: %q", got)
	}
}

func TestCORSGinMiddleware_AllowOriginFunc(t *testing.T) {
	r := gin.New()
	r.Use(CORSGinMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://github.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://github.com" {
		t.Errorf("unexpected allow-origin: %q", got)
	}
}
