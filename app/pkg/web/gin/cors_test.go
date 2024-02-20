// gin_test.go
package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
)

func TestConfigCORS(t *testing.T) {
	corsConfig := ConfigCORS()

	assert.Equal(t, []string{"http://localhost:3000"}, corsConfig.AllowOrigins)
	assert.Equal(t, []string{"POST", "OPTIONS", "GET", "PATCH", "DELETE"}, corsConfig.AllowMethods)
	assert.Equal(t, []string{"Origin"}, corsConfig.ExposeHeaders)
	assert.Equal(t, []string{
		"Content-Type", "Content-Length", "Accept-Encoding",
		"X-CSRF-Token", "Authorization", "accept", "origin",
		"Cache-Control", "X-Requested-With", "Connection",
		"Transfer-Encoding",
	}, corsConfig.AllowHeaders)
	assert.True(t, corsConfig.AllowCredentials)
	assert.Equal(t, maxAgeDuration, corsConfig.MaxAge)
}

func TestHandleOptions(t *testing.T) {
	// Create a Gin router and add the handleOptions route
	router := SetupRouter()

	// Perform an OPTIONS request to /test with the correct origin
	req, err := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Serve the request to the router
	router.ServeHTTP(w, req)

	// Check if the response status code is 204
	assert.Equal(t, noContentStatus, w.Code)
}

func SetupRouter() *gin.Engine {
	router := gin.New()

	// Set up CORS middleware
	router.Use(cors.New(ConfigCORS()))

	// Define the handleOptions route
	router.GET("/test", handleOptions)

	return router
}
