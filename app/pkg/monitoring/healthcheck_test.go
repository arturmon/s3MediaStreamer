package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mocking the mockAppHealth variable to return true.
var mockAppHealth = true

func mockHealthGET(c *gin.Context, metrics *HealthMetrics) {
	// Set the mockAppHealth before the test.
	mockAppHealth = true // or false based on your test scenario.

	// Call HealthGET with both required arguments.
	HealthGET(c, metrics)
}

func TestHealthGET(t *testing.T) {
	// Create a new HealthMetrics instance.
	metrics := NewHealthMetrics()

	// Initialize a new Gin engine with the mockHandler.
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		mockHealthGET(c, metrics)
	})

	// Create a request to the "/health" endpoint.
	req, _ := http.NewRequest("GET", "/health", http.NoBody)

	// Create a response recorder to record the response.
	w := httptest.NewRecorder()

	// Serve the request using the router.
	r.ServeHTTP(w, req)

	// Check the response status code to ensure the request was successful (HTTP 200 OK).
	assert.Equal(t, http.StatusOK, w.Code)

	// You can also check the response body or headers as needed.
	// For example, you can assert that the response contains {"status": "UP"} in the JSON body:
	assert.JSONEq(t, `{"status": "UP"}`, w.Body.String())
}
