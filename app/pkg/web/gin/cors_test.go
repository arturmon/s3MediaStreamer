package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware(t *testing.T) {
	// Create a test router with the CORSMiddleware attached
	router := gin.Default()
	router.Use(CORSMiddleware())

	// Define a test endpoint for testing
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test endpoint")
	})

	// Create a mock request to the "/test" endpoint
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")

	// Create a response recorder to record the response
	w := httptest.NewRecorder()

	// Serve the request using the router
	router.ServeHTTP(w, req)

	// Check the response headers and status code to ensure the middleware is working correctly
	if w.Code != 200 {
		t.Errorf("Expected status code 200, but got %d", w.Code)
	}
	// noinspection
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "http://localhost:3000",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		"Access-Control-Allow-Methods":     "POST, OPTIONS, GET, PUT, DELETE",
	}

	for key, value := range expectedHeaders {
		if w.Header().Get(key) != value {
			t.Errorf("Expected header '%s' to be '%s', but got '%s'", key, value, w.Header().Get(key))
		}
	}
}
