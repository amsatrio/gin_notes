package util

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testJwtSecret = "unit-test-secret"

func setupJwtEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AUTH_JWT_TOKEN_SECRET", testJwtSecret)
	t.Setenv("AUTH_JWT_TOKEN_EXPIRED_MS", "60000")
	t.Setenv("LOG_MODE", "ERROR")
}

func TestGenerateClaims_MainToken(t *testing.T) {
	setupJwtEnv(t)

	claims, err := generateClaims("alice", []string{"ROLE_USER"}, "main_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims["sub"] != "alice" || claims["type"] != "main_token" {
		t.Errorf("unexpected claims: %v", claims)
	}
	for _, k := range []string{"isAccountNonExpired", "isAccountNonLocked", "isCredentialsNonExpired", "isEnabled"} {
		if claims[k] != true {
			t.Errorf("claim %s = %v, want true", k, claims[k])
		}
	}
	iat := claims["iat"].(int64)
	exp := claims["exp"].(int64)
	if exp-iat != 60 {
		t.Errorf("exp-iat = %d, want 60", exp-iat)
	}
}

func TestGenerateClaims_RefreshTokenHasTripleExpiry(t *testing.T) {
	setupJwtEnv(t)

	claims, err := generateClaims("alice", nil, "refresh_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	iat := claims["iat"].(int64)
	exp := claims["exp"].(int64)
	if exp-iat != 180 {
		t.Errorf("exp-iat = %d, want 180", exp-iat)
	}
}

func TestGenerateClaims_InvalidEnvFallsBackToDefault(t *testing.T) {
	setupJwtEnv(t)
	t.Setenv("AUTH_JWT_TOKEN_EXPIRED_MS", "not-a-number")

	claims, err := generateClaims("alice", nil, "main_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	diff := claims["exp"].(int64) - claims["iat"].(int64)
	if diff != 86400 {
		t.Errorf("exp-iat = %d, want 86400", diff)
	}
}

func TestGenerateClaims_InvalidType(t *testing.T) {
	setupJwtEnv(t)

	claims, err := generateClaims("alice", nil, "bogus")
	if err == nil {
		t.Fatal("expected error for invalid claims type")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %v", claims)
	}
	if err.Error() != "claims_token invalid" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJwtGenerateMainToken_RoundTrip(t *testing.T) {
	setupJwtEnv(t)

	token, err := JwtGenerateMainToken("bob", []string{"A", "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	claims, err := JwtExtractAllClaims(token, "main_token")
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if got := JwtGetUserName(claims); got != "bob" {
		t.Errorf("username = %q", got)
	}
	auth := JwtGetAuthorities(claims)
	if len(auth) != 2 || auth[0] != "A" || auth[1] != "B" {
		t.Errorf("authorities = %v", auth)
	}
	if JwtIsTokenExpired(claims) {
		t.Error("token should not be expired")
	}
}

func TestJwtGenerateRefreshToken_RoundTrip(t *testing.T) {
	setupJwtEnv(t)

	token, err := JwtGenerateRefreshToken("carol", []string{"X"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := JwtExtractAllClaims(token, "refresh_token")
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if got := JwtGetUserName(claims); got != "carol" {
		t.Errorf("username = %q", got)
	}

	// a refresh token must not be accepted as a main token
	if _, err := JwtExtractAllClaims(token, "main_token"); err == nil {
		t.Error("expected type mismatch error")
	}
}

func TestJwtExtractAllClaims_TypeMismatch(t *testing.T) {
	setupJwtEnv(t)

	token, err := JwtGenerateMainToken("bob", nil)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := JwtExtractAllClaims(token, "refresh_token")
	if err == nil || claims != nil {
		t.Fatalf("expected type mismatch, got claims=%v err=%v", claims, err)
	}
	if err.Error() != "type token invalid" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJwtExtractAllClaims_InvalidToken(t *testing.T) {
	setupJwtEnv(t)

	if _, err := JwtExtractAllClaims("not.a.jwt", "main_token"); err == nil {
		t.Error("expected error for malformed token")
	}
	if _, err := JwtExtractAllClaims("", "main_token"); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestJwtExtractAllClaims_WrongSecret(t *testing.T) {
	setupJwtEnv(t)

	token, err := JwtGenerateMainToken("bob", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("AUTH_JWT_TOKEN_SECRET", "another-secret")

	if _, err := JwtExtractAllClaims(token, "main_token"); err == nil {
		t.Error("expected signature error")
	}
}

func TestJwtExtractAllClaims_UnexpectedSigningMethod(t *testing.T) {
	setupJwtEnv(t)

	tok := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"type": "main_token"})
	s, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := JwtExtractAllClaims(s, "main_token"); err == nil {
		t.Error("expected error for unsupported signing method")
	}
}

func TestJwtExtractAllClaims_ExpiredToken(t *testing.T) {
	setupJwtEnv(t)

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"type": "main_token",
		"exp":  time.Now().Add(-time.Hour).Unix(),
	})
	s, err := tok.SignedString([]byte(testJwtSecret))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := JwtExtractAllClaims(s, "main_token"); err == nil {
		t.Error("expected expiry error")
	}
}

func TestJwtIsTokenExpired(t *testing.T) {
	past := jwt.MapClaims{"exp": float64(time.Now().Add(-time.Hour).Unix())}
	if !JwtIsTokenExpired(past) {
		t.Error("expected expired")
	}
	future := jwt.MapClaims{"exp": float64(time.Now().Add(time.Hour).Unix())}
	if JwtIsTokenExpired(future) {
		t.Error("expected not expired")
	}
}

func TestJwtGetAuthorities(t *testing.T) {
	claims := jwt.MapClaims{"authorities": []interface{}{"ROLE_A", "ROLE_B"}}
	got := JwtGetAuthorities(claims)
	if len(got) != 2 || got[0] != "ROLE_A" || got[1] != "ROLE_B" {
		t.Errorf("got %v", got)
	}

	empty := JwtGetAuthorities(jwt.MapClaims{"authorities": []interface{}{}})
	if len(empty) != 0 {
		t.Errorf("expected empty, got %v", empty)
	}
}

func TestJwtGetUserName(t *testing.T) {
	t.Setenv("LOG_MODE", "ERROR")

	if got := JwtGetUserName(jwt.MapClaims{"sub": "dave"}); got != "dave" {
		t.Errorf("got %q", got)
	}
	// missing subject yields empty string without error
	if got := JwtGetUserName(jwt.MapClaims{}); got != "" {
		t.Errorf("got %q", got)
	}
	// non-string subject yields an error and empty string
	if got := JwtGetUserName(jwt.MapClaims{"sub": 123}); got != "" {
		t.Errorf("got %q", got)
	}
}
