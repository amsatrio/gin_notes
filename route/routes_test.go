package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	AppRoutes(r)
	return r
}

func TestAppRoutesRegistered(t *testing.T) {
	t.Setenv("REDIS_ENABLE", "false")
	t.Setenv("AUTH_JWT_ENABLE", "false")

	r := newRouter()
	registered := map[string]bool{}
	for _, ri := range r.Routes() {
		registered[ri.Method+" "+ri.Path] = true
	}

	expected := []string{
		"POST /v1/auth/login",
		"GET /v1/auth/logout",
		"POST /v1/auth/refresh_token",
		"GET /doc/swagger-ui/*any",
	}
	for _, name := range []string{"m_biodata", "m_role", "m_user", "m_notes", "t_reset_password", "t_token"} {
		base := "/v1/" + name
		expected = append(expected,
			"POST "+base,
			"GET "+base,
			"PUT "+base+"/:id",
			"GET "+base+"/:id",
			"PUT "+base+"/delete/:id",
			"DELETE "+base+"/:id",
			"GET "+base+"/header",
		)
	}
	for _, e := range expected {
		if !registered[e] {
			t.Errorf("route not registered: %s", e)
		}
	}
}

func TestNoRoute(t *testing.T) {
	t.Setenv("REDIS_ENABLE", "false")
	t.Setenv("AUTH_JWT_ENABLE", "false")

	r := newRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/does/not/exist", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v (%s)", err, w.Body.String())
	}
	if body["data"] != "path not found" {
		t.Errorf("data = %v, want 'path not found'", body["data"])
	}
	if body["message"] != "failed" {
		t.Errorf("message = %v, want 'failed'", body["message"])
	}
	if s, ok := body["status"].(float64); !ok || int(s) != http.StatusNotFound {
		t.Errorf("status field = %v, want 404", body["status"])
	}
}

func TestSwaggerRoute(t *testing.T) {
	r := newRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/doc/swagger-ui/index.html", nil))
	if w.Code != http.StatusOK {
		t.Errorf("swagger index status = %d, want 200", w.Code)
	}
}
