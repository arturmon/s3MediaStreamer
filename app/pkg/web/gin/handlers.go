package gin

import (
	"encoding/json"
	"errors"
	"net/http"
	"skeleton-golange-application/app/internal/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
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
	UpdateAlbum(c *gin.Context)
}

// Ping godoc
// @Summary Application liveness check function
// @Description Check if the application server is running
// @Tags health-controller
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
// @Failure		401 {object} ErrorResponse "Unauthorized"
// @Failure		500 {object} ErrorResponse "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums [get]
func (a *WebApp) GetAllAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	a.metrics.GetAllAlbumsCounter.Inc()
	albums, err := a.storage.Operations.GetAllIssues()

	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	res, _ := json.Marshal(albums)
	c.IndentedJSON(http.StatusOK, albums)
	log.Debugf("Albums response: %s", res)
}

// PostAlbums	godoc
// @Summary		Adds an album from JSON.
// @Description adds an album from JSON received in the request body.
// @Tags		album-controller
// @Accept		json
// @Produce		json
// @Param		request body config.Album true "Album details"
// @Success     201 {object} config.Album  "Created"
// @Failure     400 {object} ErrorResponse  "Bad Request"
// @Failure     500 {object} ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/album [post]
func (a *WebApp) PostAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}
	// Increment the counter for each request handled by PostAlbums
	a.metrics.PostAlbumsCounter.Inc()
	var newAlbum config.Album

	newAlbum.ID = uuid.New()

	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()

	if bindErr := c.BindJSON(&newAlbum); bindErr != nil {
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

	err = a.storage.Operations.CreateIssue(&newAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error creating album"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// GetAlbumByID godoc
// @Summary		Album whose ID value matches the id.
// noinspection
// @Description locates the album whose ID value matches the id parameter sent by the client,
// @Description	then returns that album as a response.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code album"
// @Success     200 {object} config.Album  "OK"
// @Failure     401 {object} ErrorResponse  "Unauthorized"
// @Failure     404 {object} ErrorResponse  "Not Found"
// @Failure     500 {object} ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/{code} [get]
func (a *WebApp) GetAlbumByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with getting the album
	a.metrics.GetAlbumByIDCounter.Inc()

	id := c.Param("code")
	result, err := a.storage.Operations.GetIssuesByCode(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
// @Success     204 {object}  ErrorResponse   "No Content"
// @Failure     401 {object} ErrorResponse  "Unauthorized"
// @Failure     500 {object} ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/deleteAll [delete]
func (a *WebApp) GetDeleteAll(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// Increment the counter for each request handled by GetDeleteAll
	a.metrics.GetDeleteAllCounter.Inc()

	err = a.storage.Operations.DeleteAll()
	if err != nil {
		a.logger.Fatal(err)
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
// @Success     204 {object}  ErrorResponse   "No Content"
// @Failure     401 {object} ErrorResponse  "Unauthorized"
// @Failure     404 {object} ErrorResponse  "Not Found"
// @Failure     500 {object} ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/delete/{code} [delete]
func (a *WebApp) GetDeleteByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with deleting the album by ID
	a.metrics.GetDeleteByIDCounter.Inc()

	code := c.Param("code")

	_, err = a.storage.Operations.GetIssuesByCode(code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

// UpdateAlbum godoc
// @Summary                Updates an existing album with new data.
// @Description updates an existing album with new data based on the ID parameter sent by the client.
// @Tags                album-controller
// @Accept                json
// @Produce                json
// @Param                request body config.Album true "Updated album details"
// @Success     200 {object} config.Album  "OK"
// @Failure     400 {object} ErrorResponse  "Bad Request"
// @Failure     401 {object} ErrorResponse  "Unauthorized"
// @Failure     404 {object} ErrorResponse  "Not Found"
// @Failure     500 {object} ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router                /album/update [post]
func (a *WebApp) UpdateAlbum(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// Increment the counter for each request handled by UpdateAlbum
	a.metrics.UpdateAlbumCounter.Inc()

	var newAlbum config.Album

	newAlbum.UpdatedAt = time.Now()

	if bindErr := c.BindJSON(&newAlbum); bindErr != nil {
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

	existingAlbum, getErr := a.storage.Operations.GetIssuesByCode(newAlbum.Code)
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
		} else {
			a.logger.Error(getErr)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		}
		return
	}

	// Update only the fields that have new data
	if newAlbum.Title != "" {
		existingAlbum.Title = newAlbum.Title
	}
	if newAlbum.Artist != "" {
		existingAlbum.Artist = newAlbum.Artist
	}
	if newAlbum.Price != 0 {
		existingAlbum.Price = newAlbum.Price
	}
	if newAlbum.Description != "" {
		existingAlbum.Description = newAlbum.Description
	}
	if newAlbum.Completed != existingAlbum.Completed {
		existingAlbum.Completed = newAlbum.Completed
	}

	existingAlbum.UpdatedAt = time.Now()
	// Perform the update operation
	err = a.storage.Operations.UpdateIssue(&existingAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error updating album"})
		return
	}

	c.IndentedJSON(http.StatusOK, existingAlbum)
}
