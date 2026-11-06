package middleware

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoggerMiddleware_InfoMode(t *testing.T) {
	t.Setenv("LOG_MODE", "INFO")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	r := gin.New()
	r.Use(LoggerMiddleware)
	r.GET("/log", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := doRequest(r, http.MethodGet, "/log?x=1")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	out := buf.String()
	if !strings.Contains(out, "API") || !strings.Contains(out, "/log?x=1") || !strings.Contains(out, "status: 200") {
		t.Errorf("unexpected log output: %q", out)
	}
}
