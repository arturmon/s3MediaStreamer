package gin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
// @Param       page query   int           true "Page number"
// @Param       page_size    query         int true "Number of items per page"
// @Param       sort_by      query         string false "Field to sort by (e.g., 'created_at')"
// @Param       sort_order   query         string false "Sort order ('asc' or 'desc')"
// @Param       filter       query         string false "Filter criteria"
// @Success		200 {array}  config.Album  "OK"
// @Failure		401 {object} errorResponse "Unauthorized"
// @Failure		500 {object} errorResponse "Internal Server Error"
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page or page_size parameters"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
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
// @Param		request body config.Album true "Album details"
// @Success     201 {object} config.Album  "Created"
// @Failure     400 {object} errorResponse  "Bad Request"
// @Failure     500 {object} errorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/add [post]
func (a *WebApp) PostAlbums(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by PostAlbums
	a.metrics.PostAlbumsCounter.Inc()
	var newAlbum model.Album

	newAlbum.ID = uuid.New()

	newAlbum.CreatedAt = time.Now()
	newAlbum.UpdatedAt = time.Now()
	newAlbum.Sender = "rest"

	// Чтение и вывод тела запроса
	requestBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		a.logger.Errorf("Error reading request body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}
	a.logger.Debugf("Received POST request body:\n%s", requestBody)

	// Восстановление состояния Body после чтения (иначе его нельзя будет прочитать в c.BindJSON)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	// Декодирование JSON и обработка запроса
	if bindErr := json.Unmarshal(requestBody, &newAlbum); bindErr != nil {
		a.logger.Debugf("Received POST request body:\n%s", bindErr)
		a.logger.Errorf("Invalid request payload: %v", bindErr)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}

	newAlbum.Title = strings.TrimSpace(newAlbum.Title)
	newAlbum.Artist = strings.TrimSpace(newAlbum.Artist)
	newAlbum.Code = strings.TrimSpace(newAlbum.Code)
	newAlbum.Description = strings.TrimSpace(newAlbum.Description)

	// Read the user_id from the session
	value, err := getSessionKey(c, "user_id")
	if err != nil {
		a.logger.Errorf("Error getting session value: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "could not get session value"})
		return
	}

	valueUUID, err := uuid.Parse(value.(string))
	if err != nil {
		a.logger.Errorf("Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "error converting value"})
		return
	}
	newAlbum.CreatorUser = valueUUID

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
// @Failure     401 {object} errorResponse  "Unauthorized"
// @Failure     404 {object} errorResponse  "Not Found"
// @Failure     500 {object} errorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/{code} [get]
func (a *WebApp) GetAlbumByID(c *gin.Context) {
	// Increment the session-based counter

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
// @Success     204 {object}  errorResponse   "No Content"
// @Failure     401 {object} errorResponse  "Unauthorized"
// @Failure     500 {object} errorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/deleteAll [delete]
func (a *WebApp) GetDeleteAll(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by GetDeleteAll
	a.metrics.GetDeleteAllCounter.Inc()

	err := a.storage.Operations.DeleteAll()
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
// @Success     204 {object}  errorResponse   "No Content"
// @Failure     401 {object} errorResponse  "Unauthorized"
// @Failure     404 {object} errorResponse  "Not Found"
// @Failure     500 {object} errorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/albums/delete/{code} [delete]
func (a *WebApp) GetDeleteByID(c *gin.Context) {
	// Increment the session-based counter

	// If user is authorized, proceed with deleting the album by ID
	a.metrics.GetDeleteByIDCounter.Inc()

	code := c.Param("code")

	_, err := a.storage.Operations.GetIssuesByCode(code)
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
// @Failure     400 {object} errorResponse  "Bad Request"
// @Failure     401 {object} errorResponse  "Unauthorized"
// @Failure     404 {object} errorResponse  "Not Found"
// @Failure     500 {object} errorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router                /albums/update [post]
func (a *WebApp) UpdateAlbum(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by UpdateAlbum
	a.metrics.UpdateAlbumCounter.Inc()

	var newAlbum model.Album

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
	if !newAlbum.Price.IsZero() {
		existingAlbum.Price = newAlbum.Price
	}

	if newAlbum.Description != "" {
		existingAlbum.Description = newAlbum.Description
	}
	existingAlbum.Sender = "rest"

	existingAlbum.UpdatedAt = time.Now()
	// Perform the update operation
	err := a.storage.Operations.UpdateIssue(&existingAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error updating album"})
		return
	}

	c.IndentedJSON(http.StatusOK, existingAlbum)
}
