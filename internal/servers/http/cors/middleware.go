package cors

import (
	"fmt"
	"github.com/gin-gonic/gin"
	httpcfg "github.com/mibrgmv/whoami-server-shared/config/api/http"
	"net/http"
	"strings"
)

func Middleware(config httpcfg.CORS) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(config.AllowedOrigins) > 0 {
			if contains(config.AllowedOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				origin := c.GetHeader("Origin")
				if origin != "" && contains(config.AllowedOrigins, origin) {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
		}

		if len(config.AllowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		if len(config.AllowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if config.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
