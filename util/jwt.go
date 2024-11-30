package util

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateClaims(username string, authorities []string, claims_type string) (jwt.MapClaims, error) {

	var tokenExpiredString = os.Getenv("AUTH_JWT_TOKEN_EXPIRED_MS")

	var claims jwt.MapClaims = jwt.MapClaims{}

	claims["authorities"] = authorities
	claims["isAccountNonExpired"] = true
	claims["isAccountNonLocked"] = true
	claims["isCredentialsNonExpired"] = true
	claims["isEnabled"] = true
	claims["sub"] = username
	now := time.Now()
	claims["iat"] = now.Unix()
	tokenExpired, err := strconv.ParseInt(tokenExpiredString, 10, 64)
	if err != nil {
		LogError("middleware", "generateClaims", "parse exp env error", err)
		tokenExpired = 86400000
	}

	claims["type"] = claims_type

	if claims_type == "main_token" {
		claims["exp"] = now.Add(time.Duration(tokenExpired) * time.Millisecond).Unix()
	} else if claims_type == "refresh_token" {
		claims["exp"] = now.Add(time.Duration(tokenExpired*3) * time.Millisecond).Unix()
	} else {
		return nil, errors.New("claims_token invalid")
	}

	return claims, nil
}

func JwtGenerateMainToken(username string, authorities []string) (string, error) {
	var secretKey = []byte(os.Getenv("AUTH_JWT_TOKEN_SECRET"))

	// claims (payload)
	claims, err := generateClaims(username, authorities, "main_token")
	if err != nil {
		LogError("util", "GenerateJWT", "generate claims failed", err)
		return "", err
	}

	// new jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		LogError("util", "GenerateJWT", "sign token failed", err)
		return "", err
	}

	return tokenString, nil
}

func JwtGenerateRefreshToken(username string, authorities []string) (string, error) {
	var secretKey = []byte(os.Getenv("AUTH_JWT_TOKEN_SECRET"))

	// claims (payload)
	claims, err := generateClaims(username, authorities, "refresh_token")
	if err != nil {
		LogError("util", "GenerateJWT", "generate claims failed", err)
		return "", err
	}

	// new jwt
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		LogError("util", "GenerateJWT", "sign token failed", err)
		return "", err
	}

	return tokenString, nil
}

func JwtExtractAllClaims(tokenJwt string, tokenType string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenJwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an error in parsing")
		}

		var secretKey = []byte(os.Getenv("AUTH_JWT_TOKEN_SECRET"))
		return secretKey, nil
	})

	if err != nil {
		LogError("util", "ExtractAllClaims", "parse token failed", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		LogError("util", "ExtractAllClaims", "parse claims failed", err)
		return nil, errors.New("could not parse claims")
	}

	if claims["type"] != tokenType {
		Log("ERROR", "util", "ExtractAllClaims", "type token invalid")
		return nil, errors.New("type token invalid")
	}

	Log("INFO", "util", "ExtractAllClaims", "claims: "+fmt.Sprintf("%v", claims))
	return claims, nil
}

func JwtIsTokenExpired(claims jwt.MapClaims) bool {
	expired := int64(claims["exp"].(float64))
	now := time.Now().Local().Unix()
	return expired < now
}

func JwtGetAuthorities(claims jwt.MapClaims) []string {
	authoritiesInterface := claims["authorities"].([]interface{})

	authorities := make([]string, len(authoritiesInterface))
	for i, v := range authoritiesInterface {
		authorities[i] = v.(string)
	}

	return authorities
}

func JwtGetUserName(claims jwt.MapClaims) string {
	subject, err := claims.GetSubject()
	if err != nil {
		LogError("middleware", "JwtGetUserName", "get subject error", err)
		return ""
	}

	return subject
}
