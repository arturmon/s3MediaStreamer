package userhandler

import (
	"net/http"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/acl"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/monitoring"
	"s3MediaStreamer/app/services/user"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type UserServiceInterface interface {
	ReadUserIdAndRole(c *gin.Context) (string, string, error)
}

type Handler struct {
	acl         acl.Service
	userService user.Service
	authService auth.Service
	metrics     *monitoring.Metrics
}

func NewUserHandler(acl acl.Service, userService user.Service, authService auth.Service, metrics *monitoring.Metrics) *Handler {
	return &Handler{acl, userService, authService, metrics}
}

// Register godoc
// @Summary		Registers a new user.
// @Description Register a new user with provided name, email, and password.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Security    ApiKeyAuth
// @Param		user body model.User true "Register User"
// @Success     201 {object} model.UserResponse  "Created"
// @Failure     400 {object} model.ErrorResponse "Bad Request - User with this email exists"
// @Failure     401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     500 {object} model.ErrorResponse "Internal Server Error"
// @Router		/users/register [post]
func (h *Handler) Register(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Register")
	defer span.End()

	h.metrics.RegisterAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	register, err := h.userService.Register(c.Request.Context(), data)
	if err != nil {
		h.metrics.RegisterErrorCounter.Inc()
		c.JSON(err.Code, err.Err)
		return
	}

	h.metrics.RegisterSuccessCounter.Inc()

	c.JSON(http.StatusCreated, register)
}

// Login godoc
// @Summary		Authenticates a user.
// @Description Authenticates a user with provided email and password.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Param		login body model.LoginInput true "Login User"
// @Success     200 {object} model.OkLoginResponce  "Success"
// @Failure     400 {object} model.ErrorResponse "Bad Request - Incorrect Password"
// @Failure     404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure     500 {object} model.ErrorResponse "Internal Server Error"
// @Router		/users/login [post]
func (h *Handler) Login(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Login")
	defer span.End()

	h.metrics.LoginAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		h.metrics.LoginErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	login, err := h.userService.Login(c, data)
	if err != nil {
		if err.Code == http.StatusUnauthorized {
			h.metrics.ErrPasswordCounter.Inc() // Increment the incorrect password counter
		} else {
			h.metrics.LoginErrorCounter.Inc() // Increment the login error counter
		}
		c.JSON(err.Code, err.Err)
		return
	}

	h.metrics.LoginSuccessCounter.Inc()

	c.JSON(http.StatusOK, login)
}

// DeleteUser godoc
// @Summary		Deletes a user.
// @Description Deletes the authenticated user.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} string "Success - User deleted"
// @Failure     401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     404 {object} model.ErrorResponse "Not Found - User not found"
// @Router		/users/delete [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "DeleteUser")
	defer span.End()

	h.metrics.DeleteUserAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}
	email, ok := data["email"]
	if !ok {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "email not provided"})
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), email)
	if err != nil {
		h.metrics.DeleteUserErrorCounter.Inc()
		c.JSON(err.Code, err.Err)
	}

	h.metrics.DeleteUserSuccessCounter.Inc()

	c.JSON(http.StatusOK, model.OkResponse{Message: "user deleted"})
}

// Logout godoc
// @Summary		Logs out a user.
// @Description Clears the authentication cookie, logging out the user.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} model.OkResponse  "Success"
// @Failure     401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     500 {object} model.ErrorResponse "Internal Server Error"
// @Router		/users/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Logout")
	defer span.End()

	h.metrics.LogoutAttemptCounter.Inc()

	err := h.userService.Logout(c)
	if err != nil {
		c.JSON(err.Code, err.Err)
		return
	}
	h.metrics.LogoutSuccessCounter.Inc()

	c.JSON(http.StatusOK, model.OkResponse{Message: "success"})
}

// User godoc
// @Summary Get user information
// @Description Retrieves user_handler information based on JWT in the request's cookies
// @Tags user-controller
// @Accept  */*
// @Produce json
// @Security    ApiKeyAuth
// @Success	200	{object} model.OkLoginResponce "Success"
// @Failure 401 {object} model.ErrorResponse "Unauthenticated"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /users/me [get]
func (h *Handler) User(c *gin.Context) {
	// Start h new span for the GetAllTracks operation
	_, span := otel.Tracer("").Start(c.Request.Context(), "User")
	defer span.End()

	email, err := h.acl.CheckAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, model.ErrorResponse{Message: "unauthenticated"})
		return
	}

	userInfo, errService := h.userService.User(c, email)

	if errService != nil {
		c.JSON(errService.Code, errService.Err)
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// RefreshTokenHandler godoc
// @Summary Refreshes the access token using a valid refresh token.
// @Description Validates the provided refresh token, generates a new access token, and returns it.
// @Tags user-controller
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param   refresh_token body model.ParamsRefreshTocken true "Refresh token"
// @Success 200 {object} model.ResponceRefreshTocken "Successfully refreshed access token"
// @Failure 400 {object} model.ErrorResponse "Bad Request - invalid request payload"
// @Failure 400 {object} model.ErrorResponse "Bad Request - invalid refresh token"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - invalid user email"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - invalid refresh token"
// @Failure 500 {object} model.ErrorResponse "failed to get user"
// @Failure 500 {object} model.ErrorResponse "failed to generate tokens and cookies"
// @Router /users/refresh [post]
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "refreshTokenHandler")
	defer span.End()

	var data map[string]string

	// Parse the JSON request body into the data map
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	refreshToken, exists := data["refresh_token"]
	if !exists {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "invalid refresh token"})
		return
	}

	responce, err := h.userService.RefreshTocken(c, refreshToken)
	if err != nil {
		c.JSON(err.Code, err.Err)
		return
	}

	c.JSON(http.StatusOK, responce)
}
