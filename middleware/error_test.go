package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/model/response"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newErrorEngine(handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(CustomErrorApiMiddleware())
	r.GET("/err", handler)
	return r
}

func doRequest(r http.Handler, method, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestCustomErrorApiMiddleware_NoError(t *testing.T) {
	r := newErrorEngine(func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := doRequest(r, http.MethodGet, "/err")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestCustomErrorApiMiddleware_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		message    interface{}
		setMessage bool
		wantStatus int
		wantData   bool
	}{
		{"too many request", constant.ErrorTooManyRequest, "slow down", true, http.StatusTooManyRequests, true},
		{"auth header empty", constant.ErrorAuthorizationHeaderIsEmpty, "msg", true, http.StatusUnauthorized, true},
		{"auth empty", constant.ErrorAuthorizationIsEmpty, nil, false, http.StatusUnauthorized, false},
		{"auth header invalid", constant.ErrorAuthorizationHeaderIsInvalid, "msg", true, http.StatusUnauthorized, true},
		{"auth token expired", constant.ErrorAuthorizationTokenExpired, "msg", true, http.StatusUnauthorized, true},
		{"authentication failed", constant.ErrorAuthenticationFailed, "msg", true, http.StatusUnauthorized, true},
		{"permission denied", constant.ErrorPermissionDenied, "msg", true, http.StatusForbidden, false},
		{"request invalid", constant.ErrorRequestInvalid, "bad", true, http.StatusBadRequest, true},
		{"redis delete failed", constant.ErrorRedisDeleteFailed, "redis", true, http.StatusInternalServerError, true},
		{"default case", constant.ErrorDataNotFound, "not found", true, http.StatusBadRequest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newErrorEngine(func(c *gin.Context) {
				c.Set(constant.ERROR_KEY, tt.err)
				if tt.setMessage {
					c.Set(constant.ERROR_MESSAGE, tt.message)
				}
			})

			w := doRequest(r, http.MethodGet, "/err")
			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d (%s)", tt.wantStatus, w.Code, w.Body.String())
			}

			var res response.Response
			if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if res.Status != tt.wantStatus {
				t.Errorf("expected body status %d, got %d", tt.wantStatus, res.Status)
			}
			if res.Message != tt.err.Error() {
				t.Errorf("expected message %q, got %q", tt.err.Error(), res.Message)
			}
			if res.Path != "/err" {
				t.Errorf("expected path /err, got %q", res.Path)
			}
			if tt.wantData && tt.setMessage && res.Data != tt.message {
				t.Errorf("expected data %v, got %v", tt.message, res.Data)
			}
			if !tt.wantData && res.Data != nil {
				t.Errorf("expected nil data, got %v", res.Data)
			}
		})
	}
}

func TestCustomErrorApiMiddleware_PanicRecovery(t *testing.T) {
	r := newErrorEngine(func(c *gin.Context) {
		panic("boom")
	})

	w := doRequest(r, http.MethodGet, "/err")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var res response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if res.Message != "boom" {
		t.Errorf("expected message boom, got %q", res.Message)
	}
	if res.Status != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", res.Status)
	}
}

func TestHttpErrorException(t *testing.T) {
	r := gin.New()
	r.GET("/fail", func(c *gin.Context) {
		HttpErrorException(c, http.StatusTeapot, errors.New("custom failure"))
	})

	w := doRequest(r, http.MethodGet, "/fail")
	if w.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", w.Code)
	}

	var res response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if res.Data != "custom failure" {
		t.Errorf("unexpected data: %v", res.Data)
	}
	if res.Message != "failed" {
		t.Errorf("unexpected message: %q", res.Message)
	}
	if res.Path != "/fail" {
		t.Errorf("unexpected path: %q", res.Path)
	}
}

// CustomErrorMiddleware calls recover() outside of a deferred function, so it
// can never actually recover a panic (recover only works inside defer). Only
// the happy path (Next without panic followed by Abort) is reachable.
func TestCustomErrorMiddleware_NoPanic(t *testing.T) {
	r := gin.New()
	handlerCalled := false
	r.Use(CustomErrorMiddleware)
	r.GET("/ok", func(c *gin.Context) {
		handlerCalled = true
		c.String(http.StatusOK, "fine")
	})

	w := doRequest(r, http.MethodGet, "/ok")
	if !handlerCalled {
		t.Fatal("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "fine" {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
