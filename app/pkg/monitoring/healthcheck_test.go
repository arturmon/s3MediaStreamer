package monitoring

import (
	"net/http"
	"net/http/httptest"
	"skeleton-golange-application/app/internal/config"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mocking the config.AppHealth variable to return true.
var mockAppHealth = true

func mockHealthGET(c *gin.Context) {
	config.AppHealth = mockAppHealth
	HealthGET(c)
}

func TestHealthGET(t *testing.T) {
	// Initialize a new Gin engine with the mockHandler.
	r := gin.New()
	r.GET("/health", mockHealthGET)

	// Create a request to the "/health" endpoint.
	req, _ := http.NewRequest("GET", "/health", http.NoBody) // Use http.NoBody instead of nil

	// Create a response recorder to record the response.
	w := httptest.NewRecorder()

	// Serve the request using the router.
	r.ServeHTTP(w, req)

	// Check the response status code to ensure the request was successful (HTTP 200 OK).
	assert.Equal(t, http.StatusOK, w.Code)

	// You can also check the response body or headers as needed.
	// For example, you can assert that the response contains {"status": "UP"} in the JSON body:.
	assert.JSONEq(t, `{"status": "UP"}`, w.Body.String())
}
