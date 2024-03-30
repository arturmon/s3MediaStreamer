package gin

import (
	"net/http"
	"os"
	"s3MediaStreamer/app/model"

	"github.com/gin-gonic/gin"
)

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func (a *WebApp) readUserIdAndRole(c *gin.Context) (string, string, error) {
	var err error
	userRole, ok := c.Get("userRole")
	if !ok {
		// Handle error: user role not found in context
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user role not found in context"})
		return "", "", err
	}

	userIDInterface, ok := c.Get("user_id")
	if !ok {
		// Handle error: user_id not found in context
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user_id not found in context"})
		return "", "", err
	}

	userIDString, ok := userIDInterface.(string)
	if !ok {
		// Handle error: user_id in context is not a string
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user_id in context is not a string"})
		return "", "", err
	}

	userRoleString, ok := userRole.(string)
	if !ok {
		// Handle error: user_id in context is not a string
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "role in context is not a string"})
		return "", "", err
	}
	return userRoleString, userIDString, nil
}
