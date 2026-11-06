package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/util"
)

const testSessionHeader = "X-Test-Session"

func setJwtEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AUTH_JWT_ENABLE", "true")
	t.Setenv("AUTH_JWT_TOKEN_SECRET", "test-secret-key")
	t.Setenv("AUTH_JWT_TOKEN_EXPIRED_MS", "3600000")
}

// newJwtEngine builds an engine with a cookie session store. A helper
// middleware copies the X-Test-Session header into the session auth_token.
func newJwtEngine(paths ...string) *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("test_session", cookie.NewStore([]byte("session-secret"))))
	r.Use(func(c *gin.Context) {
		if v := c.GetHeader(testSessionHeader); v != "" {
			s := sessions.Default(c)
			s.Set("auth_token", v)
		}
	})
	r.Use(CustomErrorApiMiddleware())
	r.Use(JwtMiddleware())
	for _, p := range paths {
		r.GET(p, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"username":    c.GetString("username"),
				"authorities": c.MustGet("authorities"),
			})
		})
	}
	return r
}

func jwtRequest(r http.Handler, path, bearer, session string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	if session != "" {
		req.Header.Set(testSessionHeader, session)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestJwtMiddleware_Disabled(t *testing.T) {
	t.Setenv("AUTH_JWT_ENABLE", "false")
	r := gin.New()
	r.Use(JwtMiddleware())
	r.GET("/secure", func(c *gin.Context) { c.String(http.StatusOK, "open") })

	w := doRequest(r, http.MethodGet, "/secure")
	if w.Code != http.StatusOK || w.Body.String() != "open" {
		t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
	}
}

func TestJwtMiddleware_Whitelist(t *testing.T) {
	setJwtEnv(t)
	r := gin.New()
	r.Use(CustomErrorApiMiddleware())
	r.Use(JwtMiddleware())
	for _, p := range []string{"/doc/swagger-ui/index.html", "/v1/auth/login", "/v1/auth/refresh_token", "/v1/health/public", "/v1/health/status"} {
		r.GET(p, func(c *gin.Context) { c.String(http.StatusOK, "public") })
	}

	for _, p := range []string{"/doc/swagger-ui/index.html", "/v1/auth/login", "/v1/auth/refresh_token", "/v1/health/public", "/v1/health/status"} {
		w := doRequest(r, http.MethodGet, p)
		if w.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", p, w.Code)
		}
	}
}

func TestJwtMiddleware_ProtectedWithoutToken(t *testing.T) {
	setJwtEnv(t)
	r := newJwtEngine("/v1/notes")

	w := jwtRequest(r, "/v1/notes", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJwtAuthentication_TokenCases(t *testing.T) {
	setJwtEnv(t)
	r := newJwtEngine("/v1/notes")

	valid, err := util.JwtGenerateMainToken("alice", []string{"ROLE_ADMIN"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := util.JwtGenerateMainToken("bob", []string{"ROLE_DOKTER"})
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := util.JwtGenerateRefreshToken("alice", []string{"ROLE_ADMIN"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		bearer  string
		session string
		want    int
	}{
		{"empty token: no session, no header", "", "", http.StatusUnauthorized},
		{"session but empty header", "", valid, http.StatusUnauthorized},
		{"session but header without Bearer prefix", valid, valid, http.StatusUnauthorized},
		{"session and header mismatch", "Bearer " + other, valid, http.StatusUnauthorized},
		{"invalid token", "Bearer invalid.token.value", "invalid.token.value", http.StatusUnauthorized},
		{"wrong token type", "Bearer " + refresh, refresh, http.StatusUnauthorized},
		{"valid token", "Bearer " + valid, valid, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := jwtRequest(r, "/v1/notes", tt.bearer, tt.session)
			if w.Code != tt.want {
				t.Fatalf("expected %d, got %d (%s)", tt.want, w.Code, w.Body.String())
			}
		})
	}
}

func TestJwtAuthentication_ValidSetsContext(t *testing.T) {
	setJwtEnv(t)
	r := newJwtEngine("/v1/notes")

	token, err := util.JwtGenerateMainToken("alice", []string{"ROLE_ADMIN", "ROLE_DOKTER"})
	if err != nil {
		t.Fatal(err)
	}

	w := jwtRequest(r, "/v1/notes", "Bearer "+token, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"username":"alice"`) || !strings.Contains(body, "ROLE_ADMIN") || !strings.Contains(body, "ROLE_DOKTER") {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestJwtAuthorization_Roles(t *testing.T) {
	roles := []struct {
		name string
		role string
		mw   gin.HandlerFunc
	}{
		{"admin", "ROLE_ADMIN", JwtAuthorizationAdmin},
		{"pasien", "ROLE_PASIEN", JwtAuthorizationPasien},
		{"faskes", "ROLE_FASKES", JwtAuthorizationFaskes},
		{"dokter", "ROLE_DOKTER", JwtAuthorizationDokter},
	}

	build := func(authorities []string, mw gin.HandlerFunc) *gin.Engine {
		r := gin.New()
		r.Use(CustomErrorApiMiddleware())
		r.Use(func(c *gin.Context) {
			c.Set("authorities", authorities)
		})
		r.GET("/role", mw, func(c *gin.Context) { c.String(http.StatusOK, "granted") })
		return r
	}

	for _, tt := range roles {
		t.Run(tt.name+" allowed", func(t *testing.T) {
			r := build([]string{"ROLE_OTHER", tt.role}, tt.mw)
			w := doRequest(r, http.MethodGet, "/role")
			if w.Code != http.StatusOK || w.Body.String() != "granted" {
				t.Fatalf("expected granted, got %d %s", w.Code, w.Body.String())
			}
		})
		t.Run(tt.name+" denied", func(t *testing.T) {
			r := build([]string{"ROLE_OTHER"}, tt.mw)
			w := doRequest(r, http.MethodGet, "/role")
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
		})
	}
}
