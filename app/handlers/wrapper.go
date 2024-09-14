package handlers

import (
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/user"

	"github.com/gin-gonic/gin"
)

type WrapperServiceInterface interface{}

type WrapperHandler struct {
	userService user.Service
	logger      *logs.Logger
}

func NewTrackHandler(
	userService user.Service,
	logger *logs.Logger,
) *WrapperHandler {
	return &WrapperHandler{
		userService,
		logger,
	}
}

// Wrapper with user role and ID check, including logging
func (h *WrapperHandler) WrapWithUserCheck(handler func(c *gin.Context, userContext *model.UserContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.logger.Println("Checking user role and ID")

		// Выполняем проверку userRole и userID
		userRole, userID, err := h.ReadUserIDAndRole(c)
		if err != nil {
			h.logger.Printf("Error reading user role and ID: %v", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to read user id and role"})
			return
		}

		// Создаём экземпляр UserContext
		userContext := &model.UserContext{
			UserRole: userRole,
			UserID:   userID,
		}

		// Логируем успешную проверку
		h.logger.Printf("User role: %s, User ID: %s", userRole, userID)

		// Вызываем исходный хендлер с userContext
		handler(c, userContext)
	}
}

func (h *WrapperHandler) ReadUserIDAndRole(c *gin.Context) (string, string, error) {
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
