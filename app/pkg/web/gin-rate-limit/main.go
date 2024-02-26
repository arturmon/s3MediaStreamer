package gin_rate_limit

import (
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

// SetupRateLimiter initializes and applies rate limiting to specific routes.
func SetupRateLimiter(router *gin.Engine) {
	// Initialize limiter
	store := ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
		Limit: 100,         // Maximum number of requests
		Rate:  time.Second, // Time window in minutes
	})

	mw := ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})

	// Apply limiter to selected routes
	router.Use(mw)
}
