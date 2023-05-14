package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	//prometheuse
	monitoring.GetAlbumsCounter.Inc()
	//issuesbyCode, err := mongodb.GetAllIssues(a.cfg, a.mongoClient)
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
	// prometheuse
	monitoring.PostAlbumsCounter.Inc()

	var newAlbum config.Album

	newAlbum.ID = primitive.NewObjectID()
	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()

	if err := c.BindJSON(&newAlbum); err != nil {
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

	//_, err := mongodb.GetIssuesByCode(a.cfg, a.mongoClient, newAlbum.Code)
	_, err := a.storage.Operations.GetIssuesByCode(newAlbum.Code)
	if err != mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "document with this code exists"})
		return
	}
	//mongodb.CreateIssue(a.cfg, a.mongoClient, newAlbum)
	a.storage.Operations.CreateIssue(newAlbum)
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
	// prometheuse
	monitoring.GetAlbumsCounter.Inc()

	id := c.Param("code")
	//result, err := mongodb.GetIssuesByCode(a.cfg, a.mongoClient, id)
	result, err := a.storage.Operations.GetIssuesByCode(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
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
	//prometheuse
	monitoring.GetAlbumsCounter.Inc()

	//err := mongodb.DeleteAll(a.cfg, a.mongoClient)
	err := a.storage.Operations.DeleteAll()
	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error Delete all Album"})
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
	//prometheuse
	monitoring.GetAlbumsCounter.Inc()

	id := c.Param("code")
	//err := mongodb.DeleteOne(a.cfg, a.mongoClient, id)
	err := a.storage.Operations.DeleteOne(id)
	log.Info(id)
	log.Info(err)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Delete code not found"})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})

}
