package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"skeleton-golange-application/app/internal/config"
	"strings"
	"time"

	//"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"skeleton-golange-application/app/pkg/monitoring"
)

type Handler interface {
	Register(c *gin.Context)
}

// Ping			godoc
// @Summary     Application liveness check function
// @Description do ping
// @Tags        album-controller
// @Accept      */*
// @Produce     json
// @Success     200	{object}  web.ResponseRequest   "OK"
// @Failure		404 {string} string  "Not Found"
// @Router      /ping [get]
func Ping(c *gin.Context) {
	//prometheuse
	monitoring.PingCounter.Inc()

	//c.String(http.StatusOK, "pong")
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong"})

}

// GetAllAlbums	godoc
// @Summary		Show the list of all album.
// @Description responds with the list of all albums as JSON.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Success		200 {object} main.album	"ok"
// @Failure		404 {string} string  "Not Found"
// @Router		/albums [get]
func (a *App) GetAllAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with getting the albums
	monitoring.GetAlbumsCounter.Inc()
	issuesbyCode, err := a.storage.Operations.GetAllIssues()
	if err != nil {
		a.logger.Fatal(err)
	}
	res, _ := json.Marshal(issuesbyCode)
	c.IndentedJSON(http.StatusOK, issuesbyCode)
	fmt.Println(string(res))
}

// PostAlbums	godoc
// @Summary		Adds an album from JSON.
// @Description adds an album from JSON received in the request body.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code 	path string		true "Code"
// @Param		request body main.album true "query params"
// @Success     200 {object} main.album  "ok"
// @Failure     404 {string} string  "Not Found"
// @Router		/albums/:code [post]
func (a *App) PostAlbums(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with posting the albums
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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error checking if album code exists"})
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
// @Success     200 {object} main.album  "ok"
// @Failure     404 {string} string  "Not Found"
// @Router		/albums/:code [get]
func (a *App) GetAlbumByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with getting the album
	monitoring.GetAlbumsCounter.Inc()

	id := c.Param("code")
	result, err := a.storage.Operations.GetIssuesByCode(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
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
// @Success     200 {object}  web.ResponseRequest   "OK"
// @Failure     404 {string}  string  "Not Found"
// @Router		/albums/deleteAll [get]
func (a *App) GetDeleteAll(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with deleting all albums
	monitoring.GetAlbumsCounter.Inc()

	err = a.storage.Operations.DeleteAll()
	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error Delete all Album"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})
}

// GetDeleteByID godoc
// @Summary		Album whose ID value matches the code and delete.
// @Description locates the album whose ID value matches the id parameter delete.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      int     true  "Code album"
// @Success     200 {object}  web.ResponseRequest   "OK"
// @Failure     404 {string} string  "Not Found"
// @Router		/albums/delete/:code [get]
func (a *App) GetDeleteByID(c *gin.Context) {
	// Check if user is authorized
	_, err := a.checkAuthorization(c)
	if err != nil {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "unauthenticated"})
		return
	}

	// If user is authorized, proceed with deleting the album by ID
	monitoring.GetAlbumsCounter.Inc()

	id := c.Param("code")
	err = a.storage.Operations.DeleteOne(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})
}
