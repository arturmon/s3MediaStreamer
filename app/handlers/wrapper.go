package handlers

import (
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/session"
	"s3MediaStreamer/app/services/user"

	"github.com/gin-gonic/gin"
)

type WrapperServiceInterface interface{}

type WrapperHandler struct {
	userService user.Service
	session     *session.Service
	logger      *logs.Logger
}

func NewTrackHandler(
	userService user.Service,
	session *session.Service,
	logger *logs.Logger,
) *WrapperHandler {
	return &WrapperHandler{
		userService,
		session,
		logger,
	}
}

// WrapWithUserCheck with user role and ID check, including logging.
func (h *WrapperHandler) WrapWithUserCheck(handler func(c *gin.Context, userContext *model.UserContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.logger.Info("Checking user role and ID in session data")

		// Perform user role and user ID check
		userRole, userID, err := h.ReadUserIDAndRole(c)
		if err != nil {
			h.logger.Warnf("Error reading user role and ID: %v", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to read user id and role"})
			return
		}
		// read user email in session
		userEmail, err := h.session.GetSessionKey(c, "user_email")
		if err != nil {
			h.logger.Warnf("Error reading user email: %v", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to read user email"})
			return
		}
		// Create an instance of UserContext
		userContext := &model.UserContext{
			UserRole:  userRole,
			UserEmail: userEmail.(string),
			UserID:    userID,
		}

		// Log the successfully extracted user role and ID
		h.logger.Slog().Info("Success extracted in session data", "sessionData", h.logger.ToLogFields(userContext).MaskFields())

		// Call the original handler with userContext
		handler(c, userContext)
	}
}

func (h *WrapperHandler) ReadUserIDAndRole(c *gin.Context) (string, string, error) {
	var err error
	userRole, ok := c.Get("userRole")
	if !ok {
		// Handle error: user role not found in context
		h.logger.Warnf("User role not found in context")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user role not found in context"})
		return "", "", err
	}

	userIDInterface, ok := c.Get("user_id")
	if !ok {
		// Handle error: user_id not found in context
		h.logger.Warnf("User ID not found in context")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user_id not found in context"})
		return "", "", err
	}

	userIDString, ok := userIDInterface.(string)
	if !ok {
		// Handle error: user_id in context is not a string
		h.logger.Warnf("User ID in context is not a string")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "user_id in context is not a string"})
		return "", "", err
	}

	userRoleString, ok := userRole.(string)
	if !ok {
		// Handle error: user_id in context is not a string
		h.logger.Warnf("User role in context is not a string")
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "role in context is not a string"})
		return "", "", err
	}
	return userRoleString, userIDString, nil
}
