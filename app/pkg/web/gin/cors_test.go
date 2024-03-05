// gin_test.go
package gin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	gin_ext "github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"

	"s3MediaStreamer/app/pkg/web/gin"
)

func TestConfigCORS(t *testing.T) {
	corsConfig := gin.ConfigCORS()

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
	assert.Equal(t, gin.MaxAgeDuration, corsConfig.MaxAge)
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
	assert.Equal(t, gin.NoContentStatus, w.Code)
}

func SetupRouter() *gin_ext.Engine {
	router := gin_ext.New()

	// Set up CORS middleware
	router.Use(cors.New(gin.ConfigCORS()))

	// Define the handleOptions route
	router.GET("/test", gin.HandleOptions)

	return router
}
