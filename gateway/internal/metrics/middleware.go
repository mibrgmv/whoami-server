package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GinMiddleware(collector *Collector) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" ||
			c.Request.URL.Path == "/health" ||
			c.Request.URL.Path == "/swagger" ||
			strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		collector.IncHTTPInflight()
		defer collector.DecHTTPInflight()

		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		endpoint := normalizeEndpoint(c.FullPath())
		method := c.Request.Method

		responseSize := float64(c.Writer.Size())

		collector.RecordHTTPRequest(method, endpoint, status, duration, responseSize)

		api := extractAPI(c.Request.URL.Path)
		operation := extractOperation(method, c.Request.URL.Path)
		collector.RecordAPIRequest(api, operation)

		if status[0] == '4' || status[0] == '5' {
			errorType := "client"
			if status[0] == '5' {
				errorType = "server"
			}
			collector.RecordAPIError(api, errorType)
		}
	}
}

func normalizeEndpoint(path string) string {
	path = strings.Split(path, "?")[0]
	parts := strings.Split(path, "/")

	for i, part := range parts {
		if isUUID(part) || isNumericID(part) {
			parts[i] = ":id"
		}
	}

	return strings.Join(parts, "/")
}

func isUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	parts := strings.Split(s, "-")
	return len(parts) == 5 &&
		len(parts[0]) == 8 &&
		len(parts[1]) == 4 &&
		len(parts[2]) == 4 &&
		len(parts[3]) == 4 &&
		len(parts[4]) == 12
}

func isNumericID(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func extractAPI(path string) string {
	if strings.HasPrefix(path, "/api/v1/auth") {
		return "auth"
	} else if strings.HasPrefix(path, "/api/v1/quizzes") {
		return "quizzes"
	} else if strings.HasPrefix(path, "/api/v1/questions") {
		return "questions"
	} else if strings.HasPrefix(path, "/api/v1/users") {
		return "users"
	} else if strings.HasPrefix(path, "/api/v1/history") {
		return "history"
	}
	return "other"
}

func extractOperation(method, path string) string {
	method = strings.ToLower(method)

	switch method {
	case "get":
		if strings.Contains(path, "/api/v1/quizzes") && !strings.Contains(path, ":id") {
			return "list"
		}
		if strings.Contains(path, "/api/v1/questions") && !strings.Contains(path, ":id") {
			return "list"
		}
		return "get"
	case "post":
		return "create"
	case "put", "patch":
		return "update"
	case "delete":
		return "delete"
	default:
		return method
	}
}
