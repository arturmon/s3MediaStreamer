package gin

import (
	"time"

	"github.com/gin-contrib/cors"
)

const maxAgeDuration = 12 * time.Hour

func ConfigCORS() cors.Config {
	return cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"POST, OPTIONS, GET, PUT, DELETE"},
		AllowHeaders: []string{"Origin"},
		ExposeHeaders: []string{"Content-Type, Content-Length, Accept-Encoding, " +
			"X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           maxAgeDuration,
	}
}
