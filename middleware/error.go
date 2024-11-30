package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/model/response"
	"github.com/amsatrio/gin_notes/util"
)

func CustomErrorMiddleware(c *gin.Context) {
	c.Next()
	err := recover()
	if err != nil {
		// Pass the error and stack trace to the error page template
		stackTrace := debug.Stack()
		statusCode := http.StatusInternalServerError
		c.HTML(statusCode, "error.html", gin.H{
			"errorMessage": err,
			"stackTrace":   string(stackTrace),
			"title":        "Error " + http.StatusText(statusCode) + "!",
		})
		fmt.Printf("Recovered from panic: %v\n", err)
		fmt.Println(string(stackTrace))
		return
	}

	c.Abort()
}

func CustomErrorApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// recovery after panic
		defer func() {
			if err := recover(); err != nil {
				util.Log("ERROR", "middleware", "CustomErrorApiMiddleware", fmt.Sprintf("recovery system, error: %v", err))
				er := response.Response{
					Data:      nil,
					Status:    http.StatusInternalServerError,
					Message:   fmt.Sprintf("%v", err),
					Timestamp: response.JSONTime{Time: time.Now()},
					Path:      c.FullPath(),
				}
				c.AbortWithStatusJSON(er.Status, er)
				return
			}
		}()

		c.Next()

		errorValu, ok := c.Get(constant.ERROR_KEY)
		if !ok {
			return
		}
		errorMessage, ok := c.Get(constant.ERROR_MESSAGE)
		if !ok {
			// errorMessage = fmt.Sprintf("%v", errorMessage)
			errorMessage = nil
		}
		util.Log("ERROR", "middleware", "CustomErrorApiMiddleware", fmt.Sprintf("error key: %v, error message: %v", errorValu, errorMessage))

		switch errorValu {
		// status 500
		case constant.ErrorRedisDeleteFailed:
			er := response.Response{
				Data:      errorMessage,
				Status:    http.StatusInternalServerError,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)
		// status 429
		case constant.ErrorTooManyRequest:
			er := response.Response{
				Data:      errorMessage,
				Status:    http.StatusTooManyRequests,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)
		// status 401
		case constant.ErrorAuthorizationHeaderIsEmpty,
			constant.ErrorAuthorizationIsEmpty,
			constant.ErrorAuthorizationHeaderIsInvalid,
			constant.ErrorAuthorizationTokenExpired,
			constant.ErrorAuthenticationFailed:
			er := response.Response{
				Data:      errorMessage,
				Status:    http.StatusUnauthorized,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)

		// status 403
		case constant.ErrorPermissionDenied:
			er := response.Response{
				Data:      nil,
				Status:    http.StatusForbidden,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)

		case constant.ErrorRequestInvalid:
			er := response.Response{
				Data:      errorMessage,
				Status:    http.StatusBadRequest,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)
		// status 400
		default:
			er := response.Response{
				Data:      errorMessage,
				Status:    http.StatusBadRequest,
				Message:   fmt.Sprintf("%v", errorValu),
				Timestamp: response.JSONTime{Time: time.Now()},
				Path:      c.FullPath(),
			}
			c.AbortWithStatusJSON(er.Status, er)
		}
	}
}

func HttpErrorException(c *gin.Context, status int, err error) {
	er := response.Response{
		Data:      err.Error(),
		Status:    status,
		Message:   "failed",
		Timestamp: response.JSONTime{Time: time.Now()},
		Path:      c.FullPath(),
	}
	c.AbortWithStatusJSON(er.Status, er)
}
