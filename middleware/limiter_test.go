package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func newLimiterEngine(h gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(CustomErrorApiMiddleware())
	r.Use(h)
	r.GET("/limited", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestRateLimiter_Allow(t *testing.T) {
	t.Setenv("LIMITER_MAX_REQUEST_PER_SECOND", "10000")
	r := newLimiterEngine(RateLimiter())

	for i := 0; i < 5; i++ {
		w := doRequest(r, http.MethodGet, "/limited")
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimiter_InvalidEnvUsesDefault(t *testing.T) {
	t.Setenv("LIMITER_MAX_REQUEST_PER_SECOND", "not-a-number")
	r := newLimiterEngine(RateLimiter())

	w := doRequest(r, http.MethodGet, "/limited")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimitterPerClient_Allow(t *testing.T) {
	t.Setenv("LIMITER_MAX_REQUEST_PER_SECOND", "10000")
	r := newLimiterEngine(RateLimitterPerClient())

	for i := 0; i < 5; i++ {
		w := doRequest(r, http.MethodGet, "/limited")
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimitterPerClient_InvalidEnvUsesDefault(t *testing.T) {
	t.Setenv("LIMITER_MAX_REQUEST_PER_SECOND", "")
	r := newLimiterEngine(RateLimitterPerClient())

	w := doRequest(r, http.MethodGet, "/limited")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimitterPerClient_TooManyRequests(t *testing.T) {
	// rate of 1/s with burst of 100: the 101st+ immediate request is rejected
	t.Setenv("LIMITER_MAX_REQUEST_PER_SECOND", "1")
	r := newLimiterEngine(RateLimitterPerClient())

	got429 := false
	for i := 0; i < 300; i++ {
		w := doRequest(r, http.MethodGet, "/limited")
		if w.Code == http.StatusTooManyRequests {
			got429 = true
			break
		}
	}
	if !got429 {
		t.Fatal("expected to hit the 429 path")
	}
}
