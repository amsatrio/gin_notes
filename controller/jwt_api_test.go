package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amsatrio/gin_notes/util"
)

func loginBody(user, pass string) string {
	return fmt.Sprintf(`{"username":%q,"password":%q}`, user, pass)
}

type authData struct {
	Data struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		ExpiredIn    string `json:"expired_in"`
	} `json:"data"`
}

func doLogin(t *testing.T, r http.Handler) (authData, *httptest.ResponseRecorder) {
	t.Helper()
	w := doRequest(r, http.MethodPost, "/login", loginBody(seedEmail, seedPassword))
	expectStatus(t, w, http.StatusOK)
	var out authData
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out, w
}

func TestJwtLogin_Success(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	out, w := doLogin(t, r)
	if out.Data.Token == "" {
		t.Fatalf("empty token: %s", w.Body.String())
	}
	if _, err := util.JwtExtractAllClaims(out.Data.Token, "main_token"); err != nil {
		t.Fatalf("token should be a valid main token: %v", err)
	}
	if w.Header().Get("Set-Cookie") == "" {
		t.Fatalf("expected session cookie to be set")
	}
}

func TestJwtLogin_WrongCredentials(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	w := doRequest(r, http.MethodPost, "/login", loginBody(seedEmail, "wrong-password"))
	expectStatus(t, w, http.StatusUnauthorized)

	w = doRequest(r, http.MethodPost, "/login", loginBody("nobody@example.com", seedPassword))
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestJwtLogin_BindFail(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	// validation error (missing fields / invalid email)
	expectStatus(t, doRequest(r, http.MethodPost, "/login", `{}`), http.StatusBadRequest)
	expectStatus(t, doRequest(r, http.MethodPost, "/login", loginBody("not-an-email", "pw")), http.StatusBadRequest)
	// malformed json (non validation error)
	expectStatus(t, doRequest(r, http.MethodPost, "/login", `{bad`), http.StatusBadRequest)
}

func TestJwtLogin_DBError(t *testing.T) {
	db := setupDB(t)
	r := setupRouter()
	closeDB(t, db)

	w := doRequest(r, http.MethodPost, "/login", loginBody(seedEmail, seedPassword))
	expectStatus(t, w, http.StatusUnauthorized)
}

func TestJwtLogout(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	w := doRequest(r, http.MethodPost, "/logout", "")
	expectStatus(t, w, http.StatusOK)
	if got := decodeBody(t, w)["message"]; got != "success" {
		t.Fatalf("message: %v", got)
	}
}

func TestJwtRefreshToken_Success(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	out, _ := doLogin(t, r)

	// A refresh token fails main_token extraction (type mismatch) and then
	// succeeds as refresh_token. Token field is required by binding only.
	body := fmt.Sprintf(`{"token":%q,"refresh_token":%q}`, out.Data.Token, out.Data.RefreshToken)
	w := doRequest(r, http.MethodPost, "/refresh", body)
	expectStatus(t, w, http.StatusOK)

	var res authData
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if res.Data.Token == "" {
		t.Fatalf("expected new token: %s", w.Body.String())
	}
	if res.Data.RefreshToken != out.Data.RefreshToken {
		t.Fatalf("refresh token should be echoed back")
	}
	if _, err := util.JwtExtractAllClaims(res.Data.Token, "main_token"); err != nil {
		t.Fatalf("new token should be a main token: %v", err)
	}
}

func TestJwtRefreshToken_MainTokenAsRefreshIsRejected(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	out, _ := doLogin(t, r)

	// main token passes main_token extraction -> "token is valid" error
	body := fmt.Sprintf(`{"token":%q,"refresh_token":%q}`, out.Data.Token, out.Data.Token)
	w := doRequest(r, http.MethodPost, "/refresh", body)
	expectStatus(t, w, http.StatusBadRequest)
	if got := decodeBody(t, w)["data"]; got != "token is valid" {
		t.Fatalf("unexpected data: %v", got)
	}
}

func TestJwtRefreshToken_InvalidRefreshToken(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	body := `{"token":"abc","refresh_token":"not-a-jwt"}`
	expectStatus(t, doRequest(r, http.MethodPost, "/refresh", body), http.StatusBadRequest)
}

func TestJwtRefreshToken_BindFail(t *testing.T) {
	setupDB(t)
	r := setupRouter()

	expectStatus(t, doRequest(r, http.MethodPost, "/refresh", `{}`), http.StatusBadRequest)
	expectStatus(t, doRequest(r, http.MethodPost, "/refresh", `{bad`), http.StatusBadRequest)
}
