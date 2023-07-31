package gin

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/monitoring"
	"time"
)

const SecretKey = "secret"

// Register godoc
// @Summary		Registers a new user.
// @Description Register a new user with provided name, email, and password.
// @Tags		user-controller
// @Accept		*/*
// @Produce		json
// @Param		user body config.User true "Register User"
// @Success     201 {object} config.User  "Created"
// @Failure     400 {object} map[string]string "Bad Request - User with this email exists"
// @Failure     500 {object} map[string]string "Internal Server Error"
// @Router		/users/register [post]
func (a *WebApp) Register(c *gin.Context) {
	//prometheuse
	monitoring.RegisterAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}

	// Check if user already exists
	_, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "user with this email exists"})
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := config.User{
		Id:       uuid.New(),
		Name:     data["name"],
		Email:    data["email"],
		Password: password,
	}

	err = a.storage.Operations.CreateUser(user)
	if err != nil {
		monitoring.RegisterErrorCounter.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}
	monitoring.RegisterSuccessCounter.Inc()
	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary		Authenticates a user.
// @Description Authenticates a user with provided email and password.
// @Tags		user-controller
// @Accept		*/*
// @Produce		json
// @Param		login body config.User true "Login User"
// @Success     200 {object} map[string]string  "Success"
// @Failure     400 {object} map[string]string "Bad Request - Incorrect Password"
// @Failure     404 {object} map[string]string "Not Found - User not found"
// @Failure     500 {object} map[string]string "Internal Server Error"
// @Router		/users/login [post]
func (a *WebApp) Login(c *gin.Context) {
	//prometheuse
	monitoring.LoginAttemptCounter.Inc()

	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}

	var user config.User
	user, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		monitoring.LoginErrorCounter.Inc()
		c.JSON(http.StatusBadRequest, gin.H{"message": "incorrect password"})
		// Prometheus
		monitoring.ErrPasswordCounter.Inc()
		return
	}

	var claims = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 1 day
	})

	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not login"})
		return
	}

	maxAge := 60 * 60 * 24
	c.SetCookie("jwt", token, maxAge, "/", "localhost", false, true)
	a.logger.Debugf("jwt: %s", token)
	monitoring.LoginSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "success"})
	return
}

// DeleteUser godoc
// @Summary		Deletes a user.
// @Description Deletes the authenticated user.
// @Tags		user-controller
// @Accept		*/*
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} map[string]string "Success - User deleted"
// @Failure     401 {object} map[string]string "Unauthorized - User unauthenticated"
// @Failure     404 {object} map[string]string "Not Found - User not found"
// @Router		/users/delete [delete]
func (a *WebApp) DeleteUser(c *gin.Context) {
	monitoring.DeleteUserAttemptCounter.Inc()
	email, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	err = a.storage.Operations.DeleteUser(email)
	if err != nil {
		monitoring.DeleteUserErrorCounter.Inc()
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	monitoring.DeleteUserSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
	return
}

// Logout godoc
// @Summary		Logs out a user.
// @Description Clears the authentication cookie, logging out the user.
// @Tags		user-controller
// @Accept		*/*
// @Produce		json
// @Security	ApiKeyAuth
// @Success     200 {object} map[string]string  "Success"
// @Router		/users/logout [post]
func (a *WebApp) Logout(c *gin.Context) {
	monitoring.LogoutAttemptCounter.Inc()
	Expires := time.Now().Add(-time.Hour)
	a.logger.Printf("Expires: %s", Expires)
	c.SetCookie("jwt", "", -1, "", "", false, true)
	monitoring.LogoutSuccessCounter.Inc()
	c.JSON(http.StatusOK, gin.H{"message": "success"})
	return
}

// User godoc
// @Summary Get user information
// @Description Retrieves user information based on JWT in the request's cookies
// @Tags user-controller
// @Accept  */*
// @Produce json
// @Success 200 {object} config.User "Successfully retrieved user information"
// @Failure 401 {object} gin.H{"message": string} "Unauthenticated"
// @Failure 404 {object} gin.H{"message": string} "User not found"
// @Router /user [get]
func (a *WebApp) User(c *gin.Context) {
	email, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	var user config.User
	user, err = a.storage.Operations.FindUserToEmail(email)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
	return
}

func (a *WebApp) checkAuthorization(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		return "", err
	}
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return "", err
	}
	return claims.Issuer, nil
}
