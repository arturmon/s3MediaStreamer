package app

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/mongodb"
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

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := config.User{
		Id:       primitive.NewObjectID(),
		Name:     data["name"],
		Email:    data["email"],
		Password: password,
	}

	//database.DB.Create(&user)
	err := mongodb.CreateUser(a.cfg, a.mongoClient, user)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "user with this email exists"})
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

	//database.DB.Where("email = ?", data["email"]).First(&user)
	user, err := mongodb.FindUserToEmail(a.cfg, a.mongoClient, data["email"])
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
	}
	//c.IndentedJSON(http.StatusOK, user)
	// TODO Чтение из БД
	/*
		if user.Id == 0 {
			// TODO Status -> IndentedJSON
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
	*/

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "incorrect password"})
		//prometheuse
		monitoring.ErrPasswordCounter.Inc()
		return
	}

	var claims = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		//Issuer:    user.Id.Hex(),                         // user.Email,                            // strconv.Itoa(int(user.Id)),,,,
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

	user, err = mongodb.FindUserToEmail(a.cfg, a.mongoClient, claims.Issuer)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "user not found"})
	}

	c.JSON(http.StatusOK, user)
	return
}

func (a *App) Logout(c *gin.Context) {
	Expires := time.Now().Add(-time.Hour)

	a.logger.Println("Expires: %s", Expires)

	//c.SetCookie("jwt", "", -1, "/", "localhsot", false, true)
	c.SetCookie("jwt", "", -1, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "success"})
	/*
		c.SetCookie("semaphore", "", -1, "/", "", false, true)
		c.AbortWithStatus(204)
	*/
	return
}
