package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCompressMiddleware_ReturnsHandler(t *testing.T) {
	if h := CompressMiddleware(); h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCompressMiddleware_Gzip(t *testing.T) {
	r := gin.New()
	r.Use(CompressMiddleware())
	payload := strings.Repeat("hello gzip ", 200)
	r.GET("/data", func(c *gin.Context) { c.String(http.StatusOK, payload) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", got)
	}

	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("invalid gzip body: %v", err)
	}
	defer zr.Close()
	data, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("read gzip: %v", err)
	}
	if string(data) != payload {
		t.Error("decompressed payload mismatch")
	}
}

func TestCompressMiddleware_NoGzip(t *testing.T) {
	r := gin.New()
	r.Use(CompressMiddleware())
	r.GET("/data", func(c *gin.Context) { c.String(http.StatusOK, "plain") })

	w := doRequest(r, http.MethodGet, "/data")
	if got := w.Header().Get("Content-Encoding"); got != "" {
		t.Errorf("expected no encoding, got %q", got)
	}
	if w.Body.String() != "plain" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
