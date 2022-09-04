package app

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"skeleton-golange-application/app/internal/config"
	"time"

	//"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"skeleton-golange-application/app/pkg/client/mongodb"
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
	//issuesbyCode, err := mongodb.GetAllIssues(a.cfg)
	issuesbyCode, err := mongodb.GetAllIssues(a.mongoConn)
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
	//prometheuse
	monitoring.PostAlbumsCounter.Inc()

	var newAlbum config.Album

	newAlbum.ID = primitive.NewObjectID()
	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()

	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	_, err := mongodb.GetIssuesByCode(a.cfg, newAlbum.Code)
	if err == mongo.ErrNoDocuments {
		mongodb.CreateIssue(a.cfg, newAlbum)
		c.IndentedJSON(http.StatusCreated, newAlbum)
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "document with this code exists"})

}

// GetAlbumByID godoc
// @Summary		Album whose ID value matches the id.
// @Description locates the album whose ID value matches the id parameter sent by the client, then returns that album as a response.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code album"
// @Success     200 {object} main.album  "ok"
// @Failure		400 {object} web.getAllAlbums_other "We need Code!!"
// @Failure     404 {string} string  "Not Found"
// @Router		/albums/:code [get]
func (a *App) GetAlbumByID(c *gin.Context) {
	//prometheuse
	monitoring.GetAlbumsCounter.Inc()

	id := c.Param("code")
	result, err := mongodb.GetIssuesByCode(a.cfg, id)
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

	err := mongodb.DeleteAll(a.cfg)
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
	err := mongodb.DeleteOne(a.cfg, id)

	log.Info(id)
	log.Info(err)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Delete code not found"})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})

}
