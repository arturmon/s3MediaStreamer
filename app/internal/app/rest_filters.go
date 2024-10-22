package app

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var excludedPaths = map[string]bool{
	"/health/liveness":  true,
	"/health/readiness": true,
	"/metrics":          true,
}

func excludePathsFromTracing(excluded map[string]bool) otelgin.GinFilter {
	return func(c *gin.Context) bool {
		// Skip tracing for this endpoint
		if _, ok := excluded[c.Request.URL.Path]; ok {
			return false
		}
		return true
	}
}
