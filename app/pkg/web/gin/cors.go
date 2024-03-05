package gin

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/cors"
)

const (
	MaxAgeDuration  = 12 * time.Hour
	NoContentStatus = 204 // Define a constant for HTTP status code 204.
)

func ConfigCORS() cors.Config {
	return cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		AllowMethods:  []string{"POST", "OPTIONS", "GET", "PATCH", "DELETE"},
		ExposeHeaders: []string{"Origin"},
		AllowHeaders: []string{
			"Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept", "origin",
			"Cache-Control", "X-Requested-With", "Connection",
			"Transfer-Encoding",
		},
		AllowCredentials: true,
		MaxAge:           MaxAgeDuration,
	}
}

func HandleOptions(c *gin.Context) {
	c.AbortWithStatus(NoContentStatus)
}
