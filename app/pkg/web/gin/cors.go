package gin

import (
	"github.com/gin-gonic/gin"
	"time"

	"github.com/gin-contrib/cors"
)

const maxAgeDuration = 12 * time.Hour

func ConfigCORS() cors.Config {
	return cors.Config{
		AllowOrigins:  []string{"http://localhost:3000"},
		AllowMethods:  []string{"POST", "OPTIONS", "GET", "PUT", "DELETE"},
		ExposeHeaders: []string{"Origin"},
		AllowHeaders: []string{
			"Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept", "origin",
			"Cache-Control", "X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           maxAgeDuration,
	}
}

func handleOptions(c *gin.Context) {
	c.AbortWithStatus(204)
	return
}
