package gin

import (
	"fmt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const SecretKey = "secret"
const bcryptCost = 14
const jwtExpirationHours = 24
const secondsInOneMinute = 60
const minutesInOneHour = 60
const hoursInOneDay = 24

// Register godoc
// @Summary		Registers a new user.
// @Description Register a new user with provided name, email, and password.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Param		user body config.User true "Register User"
// @Success     201 {object} config.User  "Created"
// @Failure     400 {object} ErrorResponse "Bad Request - User with this email exists"
// @Failure     500 {object} ErrorResponse "Internal Server Error"
// @Router		/users/register [post]
func (a *WebApp) Register(c *gin.Context) {
	// prometheus
	a.metrics.RegisterAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}

	// Check if user already exists
	_, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err == nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, gin.H{"message": "user with this email exists"})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcryptCost)
	if err != nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}

	user := config.User{
		ID:       uuid.New(),
		Name:     data["name"],
		Email:    data["email"],
		Password: passwordHash,
	}

	err = a.storage.Operations.CreateUser(user)
	if err != nil {
		a.metrics.RegisterErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
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
// @Param		login body config.User true "Login User"
// @Success     200 {object} ErrorResponse  "Success"
// @Failure     400 {object} ErrorResponse "Bad Request - Incorrect Password"
// @Failure     404 {object} ErrorResponse "Not Found - User not found"
// @Failure     500 {object} ErrorResponse "Internal Server Error"
// @Router		/users/login [post]
func (a *WebApp) Login(c *gin.Context) {
	a.metrics.LoginAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}

	user, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err != nil {
		a.metrics.LoginErrorCounter.Inc()
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	bcryptErr := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"]))
	if bcryptErr != nil {
		a.metrics.LoginErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, gin.H{"message": "incorrect password"})
		a.metrics.ErrPasswordCounter.Inc() // Prometheus
		return
	}

	var key = []byte(SecretKey)
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": user.Email,
		"exp": time.Now().Add(time.Hour * jwtExpirationHours).Unix(), // 1 day
	})

	token, err := claims.SignedString(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not login"})
		return
	}

	maxAge := secondsInOneMinute * minutesInOneHour * hoursInOneDay
	c.SetCookie("jwt", token, maxAge, "/", "localhost", false, true)
	a.logger.Debugf("jwt: %s", token)
	a.metrics.LoginSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// DeleteUser godoc
// @Summary		Deletes a user.
// @Description Deletes the authenticated user.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} string "Success - User deleted"
// @Failure     401 {object} ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     404 {object} ErrorResponse "Not Found - User not found"
// @Router		/users/delete [delete]
func (a *WebApp) DeleteUser(c *gin.Context) {
	a.metrics.DeleteUserAttemptCounter.Inc()
	email, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	err = a.storage.Operations.DeleteUser(email)
	if err != nil {
		a.metrics.DeleteUserErrorCounter.Inc()
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	a.metrics.DeleteUserSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// Logout godoc
// @Summary		Logs out a user.
// @Description Clears the authentication cookie, logging out the user.
// @Tags		user-controller
// @Accept		json
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} ErrorResponse  "Success"
// @Router		/users/logout [post]
func (a *WebApp) Logout(c *gin.Context) {
	a.metrics.LogoutAttemptCounter.Inc()
	expires := time.Now().Add(-time.Hour)
	a.logger.Debugf("Expires: %s", expires)
	c.SetCookie("jwt", "", -1, "", "", false, true)
	a.metrics.LogoutSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// User godoc
// @Summary Get user information
// @Description Retrieves user information based on JWT in the request's cookies
// @Tags user-controller
// @Accept  */*
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} UserResponse "Successfully retrieved user information"
// @Failure 401 {object} ErrorResponse "Unauthenticated"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /user [get]
func (a *WebApp) User(c *gin.Context) {
	email, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, ErrorResponse{Message: "unauthenticated"})
		return
	}

	var user config.User
	user, err = a.storage.Operations.FindUserToEmail(email)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, ErrorResponse{Message: "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UserResponse represents the response object for the user information endpoint.
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	// Add other fields from the config.User struct that you want to expose in the response.
}

// ErrorResponse represents the response object for error responses.
type ErrorResponse struct {
	Message string `json:"message"`
}

func (a *WebApp) checkAuthorization(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		return "", err
	}

	key := []byte(SecretKey)
	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid JWT token")
	}

	return claims["iss"].(string), nil
}
