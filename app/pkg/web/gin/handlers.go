package gin

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"skeleton-golange-application/app/internal/config"
	"strings"
	"time"

	//"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"skeleton-golange-application/app/pkg/monitoring"
)

type Handler interface {
	GetAllAlbums(c *gin.Context)
	PostAlbums(c *gin.Context)
	GetAlbumByID(c *gin.Context)
	GetDeleteAll(c *gin.Context)
	GetDeleteByID(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
	DeleteUser(c *gin.Context)
	Logout(c *gin.Context)
	User(c *gin.Context)
	checkAuthorization(c *gin.Context) (string, error)
}

// Ping godoc
// @Summary Application liveness check function
// @Description Check if the application server is running
// @Tags health-check
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong"})
}

// GetAllAlbums	godoc
// @Summary		Show the list of all albums.
// @Description responds with the list of all albums as JSON.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Success		200 {array} config.Album	"OK"
// @Failure		401 {object} map[string]string "Unauthorized"
// @Failure		500 {object} map[string]string "Internal Server Error"
// @Router		/albums [get]
func (a *WebApp) GetAllAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	monitoring.GetAllAlbumsCounter.Inc()
	albums, err := a.storage.Operations.GetAllIssues()

	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	res, _ := json.Marshal(albums)
	c.IndentedJSON(http.StatusOK, albums)
	fmt.Println(string(res))
}

// PostAlbums	godoc
// @Summary		Adds an album from JSON.
// @Description adds an album from JSON received in the request body.
// @Tags		album-controller
// @Accept		json
// @Produce		json
// @Param		code 	path string		true "Code"
// @Param		request body config.Album true "Album details"
// @Success     201 {object} config.Album  "Created"
// @Failure     400 {object} map[string]string  "Bad Request"
// @Failure     500 {object} map[string]string  "Internal Server Error"
// @Router		/albums/:code [post]
func (a *WebApp) PostAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	// Increment the counter for each request handled by PostAlbums
	monitoring.PostAlbumsCounter.Inc()
	var newAlbum config.Album

	newAlbum.ID = uuid.New()
	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()

	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}
	newAlbum.Title = strings.TrimSpace(newAlbum.Title)
	newAlbum.Artist = strings.TrimSpace(newAlbum.Artist)
	newAlbum.Code = strings.TrimSpace(newAlbum.Code)
	newAlbum.Description = strings.TrimSpace(newAlbum.Description)

	if newAlbum.Code == "" || newAlbum.Artist == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "empty required fields `Code` or `Artist`"})
		return
	}

	_, err = a.storage.Operations.GetIssuesByCode(newAlbum.Code)
	if err == nil {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": "album code already exists"})
		return
	}

	err = a.storage.Operations.CreateIssue(newAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error creating album"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newAlbum)
	return
}

// GetAlbumByID godoc
// @Summary		Album whose ID value matches the id.
// @Description locates the album whose ID value matches the id parameter sent by the client, then returns that album as a response.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code album"
// @Success     200 {object} config.Album  "OK"
// @Failure     401 {object} map[string]string  "Unauthorized"
// @Failure     404 {object} map[string]string  "Not Found"
// @Failure     500 {object} map[string]string  "Internal Server Error"
// @Router		/albums/:code [get]
func (a *WebApp) GetAlbumByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with getting the album
	monitoring.GetAlbumByIDCounter.Inc()

	id := c.Param("code")
	result, err := a.storage.Operations.GetIssuesByCode(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

// GetDeleteAll godoc
// @Summary		Complete removal of all albums.
// @Description Delete ALL.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Success     204 {object}  map[string]string   "No Content"
// @Failure     401 {object} map[string]string  "Unauthorized"
// @Failure     500 {object} map[string]string  "Internal Server Error"
// @Router		/albums [delete]
func (a *WebApp) GetDeleteAll(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// Increment the counter for each request handled by GetDeleteAll
	monitoring.GetDeleteAllCounter.Inc()

	err = a.storage.Operations.DeleteAll()
	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error Delete all Album"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, gin.H{"message": "OK"})
}

// GetDeleteByID godoc
// @Summary		Deletes album whose ID value matches the code.
// @Description locates the album whose ID value matches the id parameter and deletes it.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code album"
// @Success     204 {object}  map[string]string   "No Content"
// @Failure     401 {object} map[string]string  "Unauthorized"
// @Failure     404 {object} map[string]string  "Not Found"
// @Failure     500 {object} map[string]string  "Internal Server Error"
// @Router		/albums/:code [delete]
func (a *WebApp) GetDeleteByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with deleting the album by ID
	monitoring.GetDeleteByIDCounter.Inc()

	code := c.Param("code")

	_, err = a.storage.Operations.GetIssuesByCode(code)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		}
		return
	}

	err = a.storage.Operations.DeleteOne(code)
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error deleting album"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, gin.H{"message": "OK"})
}
