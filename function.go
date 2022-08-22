package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	pingCounter.Inc()

	//c.String(http.StatusOK, "pong")
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong"})

}

// getAlbums	godoc
// @Summary		Show the list of all album.
// @Description responds with the list of all albums as JSON.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Success		200 {object} main.album	"ok"
// @Failure		404 {string} string  "Not Found"
// @Router		/albums [get]
func GetAllAlbums(c *gin.Context) {
	//prometheuse
	getAlbumsCounter.Inc()

	issuesbyCode, _ := GetAllIssues()
	res, _ := json.Marshal(issuesbyCode)
	c.IndentedJSON(http.StatusOK, issuesbyCode)
	fmt.Println(string(res))
}

// postAlbums	godoc
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
func PostAlbums(c *gin.Context) {
	//prometheuse
	postAlbumsCounter.Inc()

	var newAlbum album

	newAlbum.ID = primitive.NewObjectID()
	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()

	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	CreateIssue(newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID godoc
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
func GetAlbumByID(c *gin.Context) {
	//prometheuse
	getAlbumsCounter.Inc()

	id := c.Param("code")
	result, err := GetIssuesByCode(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	}
	c.IndentedJSON(http.StatusOK, result)
}

// getDeleteAll godoc
// @Summary		Complete removal of all albums.
// @Description Delete ALL.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Success     200 {object}  web.ResponseRequest   "OK"
// @Failure     404 {string}  string  "Not Found"
// @Router		/albums/deleteAll [get]
func GetDeleteAll(c *gin.Context) {
	//prometheuse
	getAlbumsCounter.Inc()

	err := DeleteAll()
	if err != nil {
		log.Fatal(err)
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Error Delete all Album"})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})
}

// getDeleteByID godoc
// @Summary		Album whose ID value matches the code and delete.
// @Description locates the album whose ID value matches the id parameter delete.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      int     true  "Code album"
// @Success     200 {object}  web.ResponseRequest   "OK"
// @Failure     404 {string} string  "Not Found"
// @Router		/albums/delete/:code [get]
func GetDeleteByID(c *gin.Context) {
	//prometheuse
	getAlbumsCounter.Inc()

	id := c.Param("code")
	err := DeleteOne(id)

	log.Info(id)
	log.Info(err)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Delete code not found"})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})

}
