package middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"

	"github.com/amsatrio/gin_notes/util"
)

func CompressMiddleware() gin.HandlerFunc {
	util.Log("INFO", "middleware", "CompressMiddleware", "init gzip compression")
	return gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".pdf", ".mp4"}), gzip.WithExcludedPaths([]string{"/test/"}))
}
