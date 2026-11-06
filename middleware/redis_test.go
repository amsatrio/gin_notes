package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/model/response"
)

// setupRedis starts a miniredis instance and wires it into the initializer.
func setupRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr := miniredis.RunT(t)

	oldClient := initializer.RDB
	initializer.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	initializer.RCTX = t.Context()
	t.Cleanup(func() {
		_ = initializer.RDB.Close()
		initializer.RDB = oldClient
	})

	t.Setenv("REDIS_ENABLE", "true")
	return mr
}

func newRedisEngine(hits *int) *gin.Engine {
	r := gin.New()
	r.Use(CustomErrorApiMiddleware())
	r.Use(RedisMiddleware)

	handler := func(c *gin.Context) {
		if hits != nil {
			*hits++
		}
		c.JSON(http.StatusOK, response.Response{
			Path:      c.Request.URL.Path,
			Timestamp: response.JSONTime{Time: time.Now()},
			Status:    http.StatusOK,
			Message:   "success",
			Data:      "payload",
		})
	}
	for _, p := range []string{"/v1/notes", "/swagger/index.html", "/v1/file/download"} {
		r.GET(p, handler)
	}
	r.POST("/v1/notes", handler)
	r.PUT("/v1/notes", handler)
	r.DELETE("/v1/notes", handler)
	return r
}

func TestRedisMiddleware_Disabled(t *testing.T) {
	t.Setenv("REDIS_ENABLE", "false")
	hits := 0
	r := newRedisEngine(&hits)

	w := doRequest(r, http.MethodGet, "/v1/notes")
	if w.Code != http.StatusOK || hits != 1 {
		t.Fatalf("expected passthrough, got code=%d hits=%d", w.Code, hits)
	}
}

func TestRedisMiddleware_SwaggerAndFileSkip(t *testing.T) {
	mr := setupRedis(t)
	hits := 0
	r := newRedisEngine(&hits)

	for _, p := range []string{"/swagger/index.html", "/v1/file/download"} {
		w := doRequest(r, http.MethodGet, p)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", p, w.Code)
		}
	}
	if hits != 2 {
		t.Errorf("expected handler called twice, got %d", hits)
	}
	if len(mr.Keys()) != 0 {
		t.Errorf("expected nothing cached, got %v", mr.Keys())
	}
}

func TestRedisMiddleware_GzipHeader(t *testing.T) {
	setupRedis(t)
	r := newRedisEngine(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %q", got)
	}
}

func TestRedisMiddleware_CacheMissThenHit(t *testing.T) {
	mr := setupRedis(t)
	hits := 0
	r := newRedisEngine(&hits)

	// miss: handler runs and response gets cached
	w := doRequest(r, http.MethodGet, "/v1/notes?page=1")
	if w.Code != http.StatusOK || hits != 1 {
		t.Fatalf("miss: code=%d hits=%d", w.Code, hits)
	}
	if !mr.Exists("/v1/notes?page=1") {
		t.Fatalf("expected key cached, keys=%v", mr.Keys())
	}

	// hit: served from cache, handler not invoked
	w = doRequest(r, http.MethodGet, "/v1/notes?page=1")
	if w.Code != http.StatusOK {
		t.Fatalf("hit: expected 200, got %d", w.Code)
	}
	if hits != 1 {
		t.Errorf("expected handler to not run on cache hit, hits=%d", hits)
	}
	var res response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v (%s)", err, w.Body.String())
	}
	if res.Message != "success" || res.Data != "payload" {
		t.Errorf("unexpected cached response: %+v", res)
	}
}

func TestRedisMiddleware_BodyIDCacheKey(t *testing.T) {
	mr := setupRedis(t)
	hits := 0
	r := newRedisEngine(&hits)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/notes", strings.NewReader(`{"id": 42}`))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !mr.Exists("/v1/notes:id=42") {
		t.Fatalf("expected id-based cache key, keys=%v", mr.Keys())
	}

	// body without a valid id => plain key
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/notes", strings.NewReader(`{"id": "x"}`))
	r.ServeHTTP(w, req)
	if !mr.Exists("/v1/notes") {
		t.Fatalf("expected plain cache key, keys=%v", mr.Keys())
	}
}

func TestRedisMiddleware_CachedInvalidData(t *testing.T) {
	mr := setupRedis(t)
	r := newRedisEngine(nil)

	if err := mr.Set("/v1/notes", "not-json"); err != nil {
		t.Fatal(err)
	}
	hits := 0
	r = newRedisEngine(&hits)
	w := doRequest(r, http.MethodGet, "/v1/notes")
	// middleware logs the unmarshal error and returns, so the chain continues to the handler
	if w.Code != http.StatusOK || hits != 1 {
		t.Errorf("expected handler to run, code=%d hits=%d", w.Code, hits)
	}
}

func TestRedisMiddleware_CacheSetFailureIsTolerated(t *testing.T) {
	mr := setupRedis(t)
	r := gin.New()
	r.Use(RedisMiddleware)
	r.GET("/v1/notes", func(c *gin.Context) {
		// redis goes away after the cache lookup, so Set fails
		mr.Close()
		c.String(http.StatusOK, "ok")
	})

	w := doRequest(r, http.MethodGet, "/v1/notes")
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestRedisMiddleware_WriteMethodsInvalidateCache(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			mr := setupRedis(t)
			hits := 0
			r := newRedisEngine(&hits)

			_ = mr.Set("/v1/notes", "a")
			_ = mr.Set("/v1/notes?page=1", "b")
			_ = mr.Set("/v1/notes:id=1", "c")
			_ = mr.Set("/other", "keep")

			w := doRequest(r, method, "/v1/notes")
			if w.Code != http.StatusOK || hits != 1 {
				t.Fatalf("expected handler to run, code=%d hits=%d", w.Code, hits)
			}
			if mr.Exists("/v1/notes") || mr.Exists("/v1/notes?page=1") || mr.Exists("/v1/notes:id=1") {
				t.Errorf("expected notes keys deleted, keys=%v", mr.Keys())
			}
			if !mr.Exists("/other") {
				t.Error("unrelated key must remain")
			}
		})
	}
}

func TestRedisMiddleware_InvalidateFailure(t *testing.T) {
	mr := setupRedis(t)
	hits := 0
	r := newRedisEngine(&hits)

	mr.Close()

	w := doRequest(r, http.MethodPost, "/v1/notes")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if hits != 0 {
		t.Errorf("handler must not run, hits=%d", hits)
	}
}

func TestExtractIDFromBody(t *testing.T) {
	id, err := extractIDFromBody([]byte(`{"id": 7}`))
	if err != nil || id != 7 {
		t.Errorf("expected 7, got %v (%v)", id, err)
	}
	if _, err := extractIDFromBody([]byte(`{"id": "7"}`)); err == nil {
		t.Error("expected error for non numeric id")
	}
	if _, err := extractIDFromBody([]byte(`{}`)); err == nil {
		t.Error("expected error for missing id")
	}
	if _, err := extractIDFromBody([]byte(`not json`)); err == nil {
		t.Error("expected error for invalid json")
	}
}

func TestDeleteRedisDataContainsKey(t *testing.T) {
	mr := setupRedis(t)
	_ = mr.Set("/v1/a:1", "x")
	_ = mr.Set("/v1/a:2", "y")
	_ = mr.Set("/v1/b:1", "z")

	count, err := deleteRedisDataContainsKey("/v1/a*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}
	if !mr.Exists("/v1/b:1") {
		t.Error("expected /v1/b:1 to remain")
	}

	mr.Close()
	if _, err := deleteRedisDataContainsKey("/v1/*"); err == nil {
		t.Error("expected error when redis is down")
	}
}

func TestResponseCaptureWriter(t *testing.T) {
	setupRedis(t)
	r := gin.New()
	r.Use(RedisMiddleware)
	r.GET("/v1/notes", func(c *gin.Context) {
		c.String(http.StatusOK, "captured")
	})
	w := doRequest(r, http.MethodGet, "/v1/notes")
	if w.Body.String() != "captured" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
	got, err := initializer.RDB.Get(initializer.RCTX, "/v1/notes").Result()
	if err != nil || got != "captured" {
		t.Errorf("expected cached body, got %q (%v)", got, err)
	}
}
