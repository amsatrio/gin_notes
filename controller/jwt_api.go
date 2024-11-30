package controller

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/model/request"
	"github.com/amsatrio/gin_notes/model/response"
	"github.com/amsatrio/gin_notes/service"
	"github.com/amsatrio/gin_notes/util"
)

func JwtLogin(c *gin.Context) {
	body := request.RequestAuth{}
	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		out, _ := util.ValidateError(err)
		if out != nil {
			c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
			c.Set(constant.ERROR_MESSAGE, out)
			c.Abort()
			return
		}
		util.Log("ERROR", "controllers", "JwtLogin", "bind error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// login logic
	jwtService := service.NewJwtServiceImpl(initializer.DB)
	authorities, err := jwtService.JwtAuthenticate(c, body)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorAuthenticationFailed)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// generate token
	token, err := util.JwtGenerateMainToken(body.Username, authorities)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}
	// generate refresh token
	refreshToken, err := util.JwtGenerateRefreshToken(body.Username, authorities)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// save token to session
	session := sessions.Default(c)
	session.Set("auth_token", token)
	err = session.Save()
	if err != nil {
		util.Log("ERROR", "controllers", "JwtLogin", "save token to session error: "+err.Error())
	}

	responseAuth := &response.ResponseAuth{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiredIn:    os.Getenv("AUTH_JWT_TOKEN_EXPIRED_MS"),
	}

	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = responseAuth
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

func JwtLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("auth_token")
	session.Save()
	res := &response.Response{}
	res.Timestamp = response.JSONTime{Time: time.Now()}
	res.Data = nil
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}

func JwtRefreshToken(c *gin.Context) {
	body := request.RequestRefreshToken{}
	err := c.ShouldBindBodyWithJSON(&body)
	if err != nil {
		out, _ := util.ValidateError(err)
		if out != nil {
			c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
			c.Set(constant.ERROR_MESSAGE, out)
			c.Abort()
			return
		}
		util.Log("ERROR", "controllers", "JwtRefreshToken", "bind error: "+err.Error())
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// verify token
	claims, err := util.JwtExtractAllClaims(body.RefreshToken, "main_token")
	if err == nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, "token is valid")
		c.Abort()
		return
	}
	_ = claims

	// verify refresh token
	refreshClaims, err := util.JwtExtractAllClaims(body.RefreshToken, "refresh_token")
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// re-generate token
	authorities := util.JwtGetAuthorities(refreshClaims)

	subject, err := refreshClaims.GetSubject()
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}
	token, err := util.JwtGenerateMainToken(subject, authorities)
	if err != nil {
		c.Set(constant.ERROR_KEY, constant.ErrorRequestInvalid)
		c.Set(constant.ERROR_MESSAGE, err.Error())
		c.Abort()
		return
	}

	// save token to session
	session := sessions.Default(c)
	session.Set("auth_token", token)
	if err := session.Save(); err != nil {
		util.Log("ERROR", "controllers", "JwtRefreshToken", "save token to session error: "+err.Error())
	}

	responseAuth := &response.ResponseAuth{
		Token:        token,
		RefreshToken: body.RefreshToken,
		ExpiredIn:    os.Getenv("AUTH_JWT_TOKEN_EXPIRED_MS"),
	}

	res := &response.ResponseTimestamp{}
	res.Timestamp = time.Now()
	res.Data = responseAuth
	res.Status = http.StatusOK
	res.Message = "success"
	res.Path = c.FullPath()

	c.JSON(res.Status, res)
}
