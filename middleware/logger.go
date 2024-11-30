package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/util"
)

func LoggerMiddleware(c *gin.Context) {
	clientIP := c.ClientIP()
	targetAPI := c.Request.RequestURI

	startTime := time.Now()

	// Process the request
	c.Next()

	elapsedTime := time.Since(startTime)

	statusCode := c.Writer.Status()

	util.LogAPI(clientIP, targetAPI, statusCode, elapsedTime.String())
}
