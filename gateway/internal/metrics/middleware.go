package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GinMiddleware(collector *Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/metrics" || path == "/health" || strings.HasPrefix(path, "/swagger") {
			c.Next()
			return
		}

		collector.IncInflight()
		defer collector.DecInflight()

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		normalizedPath := normalizePath(c.FullPath())
		if normalizedPath == "" {
			normalizedPath = "not_found"
		}

		collector.RecordRequest(c.Request.Method, normalizedPath, status, duration)
	}
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == ":id" || part == ":path" || strings.HasPrefix(part, ":") {
			parts[i] = ":id"
		}
	}

	return strings.Join(parts, "/")
}
