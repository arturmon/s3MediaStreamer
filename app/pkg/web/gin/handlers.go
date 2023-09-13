package gin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"skeleton-golange-application/model"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
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
// @Success 200 {object} model.UserResponse "OK"
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, model.OkResponse{Message: "pong"})
}

// GetAllAlbums	godoc
// @Summary		Show the list of all albums.
// @Description responds with the list of all albums as JSON.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param       page query   int           false "Page number"
// @Param       page_size    query         int false "Number of items per page"
// @Param       sort_by      query         string false "Field to sort by (e.g., 'created_at')"
// @Param       sort_order   query         string false "Sort order ('asc' or 'desc')"
// @Param       filter       query         string false "Filter criteria ('I0001' or '=I0001')"
// @Success		200 {array}  model.Album  "OK"
// @Failure		401 {object} model.ErrorResponse "Unauthorized"
// @Failure		500 {object} model.ErrorResponse "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums [get]
func (a *WebApp) GetAllAlbums(c *gin.Context) {
	a.metrics.GetAllAlbumsCounter.Inc()
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")

	// Retrieve sorting and filtering parameters from the query
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	filter := c.DefaultQuery("filter", "")

	// Convert page, pageSize, and totalPages to integers
	pageInt, errPage := strconv.Atoi(page)
	pageSizeInt, errPageSize := strconv.Atoi(pageSize)
	if errPage != nil || errPageSize != nil {
		a.logger.Error("Invalid page or page_size parameters")
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid page or page_size parameters"})
		return
	}
	var validSortOrders = map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
	}

	// Check if provided sort_order parameter is valid
	if _, validSortOrderExists := validSortOrders[sortOrder]; !validSortOrderExists {
		sortOrder = "desc" // Default to descending order
	}

	// Calculate the offset based on the pagination parameters
	offset := (pageInt - 1) * pageSizeInt

	// Retrieve paginated albums from the storage
	albums, countTotal, err := a.storage.Operations.GetAlbums(offset, pageSizeInt, sortBy, sortOrder, filter)
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
		return
	}

	// Calculate total pages based on total count and page size
	totalPages := int(math.Ceil(float64(countTotal) / float64(pageSizeInt)))

	res, _ := json.Marshal(albums)

	baseURL := "http" // По умолчанию HTTP
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		baseURL = proto
	}

	baseURL = fmt.Sprintf("%s://%s", baseURL, c.Request.Host)
	c.Header("X-Total-Count", strconv.Itoa(countTotal))
	c.Header("X-Total-Pages", strconv.Itoa(totalPages))
	c.Header("Link", generatePaginationLinks(baseURL, c.FullPath(), pageInt, totalPages, pageSize))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count,X-Total-Pages,Link")
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.IndentedJSON(http.StatusOK, albums)
	log.Debugf("Albums response: %s", res)
}

// PostAlbums	godoc
// @Summary		Adds an album from JSON.
// @Description adds an album from JSON received in the request body.
// @Tags		album-controller
// @Accept		json
// @Produce		json
// @Param		request body []model.Album true "Album details"
// @Success     201 {object} []model.Album  "Created"
// @Failure     400 {object} model.ErrorResponse  "Bad Request"
// @Failure     401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     409 {object} model.ErrorResponse  "album code already exists"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/add [post]
func (a *WebApp) PostAlbums(c *gin.Context) {
	a.metrics.PostAlbumsCounter.Inc()

	var albums []model.Album

	// Decode the request body as an array of model.Album
	if err := c.BindJSON(&albums); err != nil {
		a.logger.Errorf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	// Create an array to store insertion errors, if any
	var insertionErrors []error

	// Loop through the albums and prepare them
	for i := range albums {
		album := &albums[i]

		album.ID = uuid.New()
		album.CreatedAt = time.Now()
		album.UpdatedAt = time.Now()
		album.Sender = "rest"

		// Read user_id from the session
		value, err := getSessionKey(c, "user_id")
		if err != nil {
			a.logger.Errorf("Error getting session value: %v", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "could not get session value"})
			return
		}

		valueUUID, err := uuid.Parse(value.(string))
		if err != nil {
			a.logger.Errorf("Error: %v", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "error converting value"})
			return
		}
		album.CreatorUser = valueUUID

		album.Title = strings.TrimSpace(album.Title)
		album.Artist = strings.TrimSpace(album.Artist)
		album.Code = strings.TrimSpace(album.Code)
		album.Description = strings.TrimSpace(album.Description)

		if album.Code == "" || album.Artist == "" {
			c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "empty required fields `Code` or `Artist`"})
			return
		}

		// Check if the album code already exists
		_, err = a.storage.Operations.GetAlbumsByCode(album.Code)
		if err == nil {
			insertionErrors = append(insertionErrors, fmt.Errorf("album code already exists for album %s", album.Title))
			continue
		}
	}

	// Check if there were insertion errors
	if len(insertionErrors) > 0 {
		errorMessages := make([]string, len(insertionErrors))
		for i, err := range insertionErrors {
			errorMessages[i] = err.Error()
		}
		c.IndentedJSON(http.StatusConflict, model.ErrorResponse{Message: "Some albums could not be inserted"})
		return
	}

	// Insert all albums into the database
	if err := a.storage.Operations.CreateAlbums(albums); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, albums)
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
// @Success     200 {object} model.Album  "OK"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/{code} [get]
func (a *WebApp) GetAlbumByID(c *gin.Context) {
	// Increment the session-based counter

	// If user is authorized, proceed with getting the album
	a.metrics.GetAlbumByIDCounter.Inc()

	id := c.Param("code")
	result, err := a.storage.Operations.GetAlbumsByCode(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "album not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
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
// @Success     204 {object} model.OkResponse   "No Content"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/deleteAll [delete]
func (a *WebApp) GetDeleteAll(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by GetDeleteAll
	a.metrics.GetDeleteAllCounter.Inc()

	err := a.storage.Operations.DeleteAlbumsAll()
	if err != nil {
		a.logger.Fatal(err)
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Error Delete all Album"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// GetDeleteByID godoc
// @Summary		Deletes album whose ID value matches the code.
// @Description locates the album whose ID value matches the id parameter and deletes it.
// @Tags		album-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code album"
// @Success     204 {object} model.OkResponse   "No Content"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/delete/{code} [delete]
func (a *WebApp) GetDeleteByID(c *gin.Context) {
	// Increment the session-based counter

	// If user is authorized, proceed with deleting the album by ID
	a.metrics.GetDeleteByIDCounter.Inc()

	code := c.Param("code")

	_, err := a.storage.Operations.GetAlbumsByCode(code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "album not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
		}
		return
	}

	err = a.storage.Operations.DeleteAlbums(code)
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "error deleting album"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// UpdateAlbum godoc
// @Summary                Updates an existing album with new data.
// @Description updates an existing album with new data based on the ID parameter sent by the client.
// @Tags                album-controller
// @Accept              json
// @Produce             json
// @Param               request body model.Album true "Updated album details"
// @Success     200 {object} model.Album  "OK"
// @Failure     400 {object} model.ErrorResponse  "Bad Request"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router                /albums/update [patch]
func (a *WebApp) UpdateAlbum(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by UpdateAlbum
	a.metrics.UpdateAlbumCounter.Inc()

	var newAlbum model.Album

	newAlbum.UpdatedAt = time.Now()

	if bindErr := c.BindJSON(&newAlbum); bindErr != nil {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}
	newAlbum.Title = strings.TrimSpace(newAlbum.Title)
	newAlbum.Artist = strings.TrimSpace(newAlbum.Artist)
	newAlbum.Code = strings.TrimSpace(newAlbum.Code)
	newAlbum.Description = strings.TrimSpace(newAlbum.Description)

	if newAlbum.Code == "" || newAlbum.Artist == "" {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "empty required fields `Code` or `Artist`"})
		return
	}

	existingAlbum, getErr := a.storage.Operations.GetAlbumsByCode(newAlbum.Code)
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "album not found"})
		} else {
			a.logger.Error(getErr)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: getErr.Error()})
		}
		return
	}

	if newAlbum.Title != "" {
		existingAlbum.Title = newAlbum.Title
	}
	if newAlbum.Artist != "" {
		existingAlbum.Artist = newAlbum.Artist
	}
	if !newAlbum.Price.IsZero() {
		existingAlbum.Price = newAlbum.Price
	}

	if newAlbum.Description != "" {
		existingAlbum.Description = newAlbum.Description
	}
	existingAlbum.Likes = newAlbum.Likes

	existingAlbum.Sender = "rest"

	existingAlbum.UpdatedAt = time.Now()
	// Perform the update operation
	err := a.storage.Operations.UpdateAlbums(&existingAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, existingAlbum)
}
