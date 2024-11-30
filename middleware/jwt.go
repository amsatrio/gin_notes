package middleware

import (
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/util"
)

func JwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		util.Log("INFO", "middleware", "JwtMiddleware", "start")

		if os.Getenv("AUTH_JWT_ENABLE") != "true" {
			c.Next()
			return
		}

		whiteListPath := []string{
			"/doc/swagger-ui",
			"/v1/auth/login",
			"/v1/auth/refresh_token",
			"/v1/health/public",
			"/v1/health/status",
		}

		for _, value := range whiteListPath {
			if strings.HasPrefix(c.FullPath(), value) {
				c.Next()
				return
			}
		}

		JwtAuthentication(c)

	}
}

func getJwtToken(c *gin.Context) string {
	// get token from session
	tokenSession := ""
	session := sessions.Default(c)
	token := session.Get("auth_token")
	if token == nil {
		util.Log("INFO", "middleware", "getJwtToken", "session token is empty")
		return ""
	}
	tokenSession = token.(string)

	// get token from header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		util.Log("INFO", "middleware", "getJwtToken", "header token is empty")
		return ""
	}
	if !strings.HasPrefix(tokenString, "Bearer ") {
		util.Log("INFO", "middleware", "getJwtToken", "header token is invalid")
		return ""
	}
	tokenHeader := strings.Replace(tokenString, "Bearer ", "", 1)

	if tokenSession != tokenHeader {
		util.Log("INFO", "middleware", "getJwtToken", "session token and header token is not equals")
		return ""
	}

	return tokenHeader
}

func JwtAuthentication(c *gin.Context) {
	util.Log("INFO", "middleware", "JwtAuthentication", "start")
	tokenJwt := getJwtToken(c)
	if tokenJwt == "" {
		util.Log("INFO", "middleware", "JwtAuthentication", "token is empty")
		c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationIsEmpty)
		c.Abort()
		return
	}

	// validate token
	jwt_claim, err := util.JwtExtractAllClaims(tokenJwt, "main_token")
	if err != nil {
		util.Log("INFO", "middleware", "JwtAuthentication", "token is invalid")
		c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationHeaderIsInvalid)
		c.Abort()
		return
	}

	// check expired
	if util.JwtIsTokenExpired(jwt_claim) {
		util.Log("INFO", "middleware", "JwtAuthentication", "token is expired")
		c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationTokenExpired)
		c.Abort()
		return
	}

	// get authorities
	authorities := util.JwtGetAuthorities(jwt_claim)
	c.Set("authorities", authorities)
	//util.Log("INFO", "middleware", "JwtAuthentication", "authorities: "+fmt.Sprintf("%v", authorities))

	// get username
	username := util.JwtGetUserName(jwt_claim)
	c.Set("username", username)

	// util.Log("INFO", "middleware", "JwtAuthentication", "username is "+username)

	c.Next()

}
func JwtAuthorizationAdmin(c *gin.Context) {
	authorities := c.MustGet("authorities").([]string)

	for _, v := range authorities {
		if v == "ROLE_ADMIN" {
			c.Next()
			return
		}
	}

	util.Log("INFO", "middleware", "JwtAuthorizationAdmin", "authorities invalid")
	c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationHeaderIsInvalid)
	c.Abort()

	_ = authorities
}
func JwtAuthorizationPasien(c *gin.Context) {
	authorities := c.MustGet("authorities").([]string)

	for _, v := range authorities {
		if v == "ROLE_PASIEN" {
			c.Next()
			return
		}
	}

	util.Log("INFO", "middleware", "JwtAuthorizationPasien", "authorities invalid")
	c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationHeaderIsInvalid)
	c.Abort()

	_ = authorities
}
func JwtAuthorizationFaskes(c *gin.Context) {
	authorities := c.MustGet("authorities").([]string)

	for _, v := range authorities {
		if v == "ROLE_FASKES" {
			c.Next()
			return
		}
	}

	util.Log("INFO", "middleware", "JwtAuthorizationFaskes", "authorities invalid")
	c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationHeaderIsInvalid)
	c.Abort()

	_ = authorities
}
func JwtAuthorizationDokter(c *gin.Context) {
	authorities := c.MustGet("authorities").([]string)

	for _, v := range authorities {
		if v == "ROLE_DOKTER" {
			c.Next()
			return
		}
	}

	util.Log("INFO", "middleware", "JwtAuthorizationDokter", "authorities invalid")
	c.Set(constant.ERROR_KEY, constant.ErrorAuthorizationHeaderIsInvalid)
	c.Abort()

	_ = authorities
}
