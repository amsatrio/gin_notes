package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/constant"
	"github.com/amsatrio/gin_notes/initializer"
	"github.com/amsatrio/gin_notes/model/response"
	"github.com/amsatrio/gin_notes/util"
)

// Custom responseCaptureWriter to capture response data
type responseCaptureWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RedisMiddleware(c *gin.Context) {
	if os.Getenv("REDIS_ENABLE") == "false" {
		c.Next()
		return
	}

	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		c.Writer.Header().Set("Content-Encoding", "gzip")
	}

	// Generate a cache key based on the request URL
	cacheKey := c.Request.RequestURI
	// queryParams := c.Request.URL.Query()
	// queryString := queryParams.Encode()
	// if queryString != "" {
	// 	cacheKey += "?" + queryString
	// }
	patternKey := strings.Split(cacheKey, "?")[0] + "*"

	// skip cache if path contains swagger
	if strings.Contains(cacheKey, "swagger") {
		c.Next()
		return
	}
	if strings.Contains(cacheKey, "file") {
		c.Next()
		return
	}

	// check http method
	method := c.Request.Method
	if method == http.MethodDelete || method == http.MethodPut || method == http.MethodPost {
		if patternKey == "*" {
			c.Next()
			return
		}
		_, err := deleteRedisDataContainsKey(patternKey)
		if err != nil {
			c.Set(constant.ERROR_KEY, constant.ErrorRedisDeleteFailed)
			c.Abort()
			return
		}
		c.Next()
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err == nil {
		id, _ := extractIDFromBody(body)
		if id != 0 {
			cacheKey += ":id=" + strconv.FormatFloat(id, 'f', -1, 64)
		}
	}

	// Check if the data is cached in Redis
	cachedData, err := initializer.RDB.Get(initializer.RCTX, cacheKey).Result()
	if err == nil {
		util.Log("INFO", "middleware", "RedisMiddleware", "data found on cache "+cacheKey)

		cachedData = strings.ReplaceAll(cachedData, "\\", "")
		cachedData = strings.TrimPrefix(cachedData, "\"")
		cachedData = strings.TrimSuffix(cachedData, "\"")

		res := &response.Response{}

		err := json.Unmarshal([]byte(cachedData), &res)
		if err != nil {
			util.LogError("middleware", "RedisMiddleware", "Error: Unmarshal failed ", err)
			return
		}
		res.Timestamp = response.JSONTime{Time: time.Now()}
		resBytes := new(bytes.Buffer)
		err = json.NewEncoder(resBytes).Encode(res)
		if err != nil {
			util.LogError("middleware", "RedisMiddleware", "Error: convert Encode Response failed ", err)
			return
		}

		c.Data(http.StatusOK, "application/json", resBytes.Bytes())
		c.Abort()
		return
	}
	util.Log("INFO", "middleware", "RedisMiddleware", "data not found on cache "+cacheKey)

	// Data not found in Redis, capture the response before it's written
	responseWriter := &responseCaptureWriter{c.Writer, bytes.NewBuffer(nil)}
	c.Writer = responseWriter
	c.Next()

	// Cache the response data in Redis
	if c.Writer.Status() == http.StatusOK {

		// Cache the compressed data in Redis
		err = initializer.RDB.Set(initializer.RCTX, cacheKey, responseWriter.Body.Bytes(), 0).Err()
		if err != nil {
			util.Log("INFO", "middleware", "RedisMiddleware", "Failed to cache data in Redis "+cacheKey)
		}
	}
}

func extractIDFromBody(body []byte) (float64, error) {
	// Parse the 'id' value from JSON
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	id, ok := data["id"].(float64)
	if !ok {
		return 0, errors.New("invalid 'id' format")
	}

	return id, nil
}

func deleteRedisDataContainsKey(kePattern string) (count int, err error) {
	var foundedRecordCount int = 0
	iter := initializer.RDB.Scan(initializer.RCTX, 0, kePattern, 0).Iterator()
	for iter.Next(initializer.RCTX) {
		initializer.RDB.Del(initializer.RCTX, iter.Val())
		foundedRecordCount++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return foundedRecordCount, nil
}
