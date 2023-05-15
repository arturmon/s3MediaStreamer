package app

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

func (a *App) Register(c *gin.Context) {
	//prometheuse
	monitoring.RegisterCounter.Inc()
	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		return
	}
	// Check if user already exists
	_, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "user with this email exists"})
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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}
	c.IndentedJSON(http.StatusCreated, user)
}

func (a *App) Login(c *gin.Context) {
	//prometheuse
	monitoring.LoginCounter.Inc()
	var data map[string]string
	if err := c.BindJSON(&data); err != nil {
		return
	}
	var user config.User
	user, err := a.storage.Operations.FindUserToEmail(data["email"])
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "incorrect password"})
		//prometheuse
		monitoring.ErrPasswordCounter.Inc()
		return
	}
	var claims = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.Email,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), //1 day
	})
	token, err := claims.SignedString([]byte(SecretKey))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "could not login"})
		return
	}
	maxAge := 60 * 60 * 24
	c.SetCookie("jwt", token, maxAge, "/", "localhost", false, true)
	a.logger.Debug("jwt: %s", token)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
	return
}

func (a *App) User(c *gin.Context) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	claims := token.Claims.(*jwt.StandardClaims)
	var user config.User
	user, err = a.storage.Operations.FindUserToEmail(claims.Issuer)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
	}
	c.JSON(http.StatusOK, user)
	return
}

func (a *App) DeleteUser(c *gin.Context) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	claims := token.Claims.(*jwt.StandardClaims)
	err = a.storage.Operations.DeleteUser(claims.Issuer)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
	return
}

func (a *App) Logout(c *gin.Context) {
	Expires := time.Now().Add(-time.Hour)
	a.logger.Println("Expires: %s", Expires)
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
	return
}
