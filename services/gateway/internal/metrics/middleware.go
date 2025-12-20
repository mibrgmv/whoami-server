package metrics

import (
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	uuidRegex = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	numRegex  = regexp.MustCompile(`^\d+$`)
)

func GinMiddleware(collector *Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		collector.HTTPInflightRequests.Inc()
		defer collector.HTTPInflightRequests.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := normalizePath(c.Request.URL.Path)

		collector.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		collector.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
	}
}

func normalizePath(path string) string {
	if idx := indexOf(path, '?'); idx != -1 {
		path = path[:idx]
	}

	path = uuidRegex.ReplaceAllString(path, ":uuid")

	result := ""
	segments := splitPath(path)

	for i, seg := range segments {
		if i > 0 {
			result += "/"
		}

		if i > 0 && numRegex.MatchString(seg) {
			result += ":id"
		} else {
			result += seg
		}
	}

	return result
}

func splitPath(path string) []string {
	if path == "" || path == "/" {
		return []string{"/"}
	}

	var segments []string
	start := 0

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				segments = append(segments, path[start:i])
			}
			start = i + 1
		}
	}

	if start < len(path) {
		segments = append(segments, path[start:])
	}

	return segments
}

func indexOf(s string, ch rune) int {
	for i, c := range s {
		if c == ch {
			return i
		}
	}
	return -1
}
