package gin

import (
	"net/http"
	"skeleton-golange-application/app/model"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

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
func (a *WebApp) Register(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Register")
	defer span.End()
	// prometheus
	a.metrics.RegisterAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	// Check if user already exists
	_, err := a.storage.Operations.FindUser(c.Request.Context(), data["email"], "email")
	if err == nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "user with this email exists"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcryptCost)
	if err != nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to create user"})
		return
	}

	user := model.User{
		ID:       uuid.New(),
		Name:     data["name"],
		Email:    data["email"],
		Password: passwordHash,
		Role:     data["role"],
	}

	err = a.storage.Operations.CreateUser(c.Request.Context(), user)
	if err != nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to create user"})
		return
	}

	a.metrics.RegisterSuccessCounter.Inc()
	c.JSON(http.StatusCreated, user)
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
func (a *WebApp) Login(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Login")
	defer span.End()
	a.metrics.LoginAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	user, err := a.storage.Operations.FindUser(c.Request.Context(), data["email"], "email")
	if err != nil {
		a.metrics.LoginErrorCounter.Inc()
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "user not found"})
		return
	}

	bcryptErr := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"]))

	if bcryptErr != nil {
		a.metrics.LoginErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "incorrect password"})
		return
	}

	dataSession := map[string]interface{}{
		"user_email": user.Email,
		"user_id":    user.ID.String(),
	}
	err = setSessionData(c, dataSession)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to save session data"})
		return
	}

	accessToken, refreshToken, err := a.generateTokensAndCookies(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to generate tokens and cookies"})
		return
	}

	a.logger.Debugf("jwt: %s", accessToken)
	a.logger.Debugf("refreshToken: %s", refreshToken)
	a.metrics.LoginSuccessCounter.Inc()

	loginResponse := model.OkLoginResponce{
		Email:        user.Email,
		Role:         user.Role,
		RefreshToken: refreshToken,
		OtpEnabled:   user.OtpEnabled,
	}

	c.JSON(http.StatusOK, loginResponse)
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
func (a *WebApp) DeleteUser(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "DeleteUser")
	defer span.End()
	a.metrics.DeleteUserAttemptCounter.Inc()

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

	// directive
	if email == "admin@admin.com" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "user cannot be deleted: admin@admin.com"})
		return
	}

	user, err := a.storage.Operations.FindUser(c.Request.Context(), email, "email")
	if err != nil {
		a.metrics.DeleteUserErrorCounter.Inc()
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "user not found"})
		return
	}

	err = a.storage.Operations.DeleteUser(c.Request.Context(), user.Email)
	if err != nil {
		a.metrics.DeleteUserErrorCounter.Inc()
		c.IndentedJSON(http.StatusNotFound, model.ErrorResponse{Message: "user not found"})
		return
	}
	a.metrics.DeleteUserSuccessCounter.Inc()
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
func (a *WebApp) Logout(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Logout")
	defer span.End()
	a.metrics.LogoutAttemptCounter.Inc()
	expires := time.Now().Add(-time.Hour)
	a.logger.Debugf("Expires: %s", expires)
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.SetCookie("refresh_token", "", -1, "", "", false, true)
	err := logoutSession(c)
	if err != nil {
		a.metrics.DeleteUserErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "session logout error"})
		return
	}
	a.metrics.LogoutSuccessCounter.Inc()
	c.JSON(http.StatusOK, model.OkResponse{Message: "success"})
}

// User godoc
// @Summary Get user information
// @Description Retrieves user information based on JWT in the request's cookies
// @Tags user-controller
// @Accept  */*
// @Produce json
// @Security    ApiKeyAuth
// @Success	200	{object} model.OkLoginResponce "Success"
// @Failure 401 {object} model.ErrorResponse "Unauthenticated"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Router /users/me [get]
func (a *WebApp) User(c *gin.Context) {
	// Start a new span for the GetAllTracks operation
	_, span := otel.Tracer("").Start(c.Request.Context(), "User")
	defer span.End()

	email, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, model.ErrorResponse{Message: "unauthenticated"})
		return
	}

	var user model.User
	user, err = a.storage.Operations.FindUser(c.Request.Context(), email, "email")
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, model.ErrorResponse{Message: "user not found"})
		return
	}

	loginResponse := model.OkLoginResponce{
		Email:        user.Email,
		Role:         user.Role,
		RefreshToken: user.RefreshToken,
		OtpEnabled:   user.OtpEnabled,
	}

	c.JSON(http.StatusOK, loginResponse)
}

// @Summary Refreshes the access token using a valid refresh token.
// @Description Validates the provided refresh token, generates a new access token, and returns it.
// @Tags user-controller
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param   refresh_token body model.ParamsRefreshTocken true "Refresh token"
// @Success 200 {object} model.ResponceRefreshTocken "Successfully refreshed access token"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid refresh token"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Invalid refresh token"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Router /users/refresh [post]
func (a *WebApp) refreshTokenHandler(c *gin.Context) {
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

	claims := jwt.MapClaims{}

	// Validate and parse the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(RefreshTokenSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "invalid refresh token"})
		return
	}

	userEmail, ok := claims["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "invalid user email"})
		return
	}

	// Check if the refresh token is stored and valid
	storedRefreshToken, err := a.storage.Operations.GetStoredRefreshToken(c.Request.Context(), userEmail)
	if err != nil || refreshToken != storedRefreshToken {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "invalid refresh token"})
		return
	}

	// Generate a new access token and refresh token
	user, err := a.storage.Operations.FindUser(c.Request.Context(), userEmail, "email")
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to get user"})
		return
	}

	accessToken, newRefreshToken, err := a.generateTokensAndCookies(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "failed to generate tokens and cookies"})
		return
	}

	// Respond with the new access token
	refreshResponse := model.ResponceRefreshTocken{
		RefreshToken: newRefreshToken,
		AccessToken:  accessToken,
	}

	c.JSON(http.StatusOK, refreshResponse)

	// Update the stored refresh token with the new one
	err = a.storage.Operations.SetStoredRefreshToken(c.Request.Context(), userEmail, newRefreshToken)
	if err != nil {
		a.logger.Errorf("Failed to update stored refresh token: %v", err)
		// Handle the error as needed
	}
}
