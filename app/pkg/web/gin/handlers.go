package gin

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"skeleton-golange-application/app/model"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

type Handler interface {
	GetAllTracks(c *gin.Context)
	PostTracks(c *gin.Context)
	GetTrackByID(c *gin.Context)
	GetDeleteAll(c *gin.Context)
	GetDeleteByID(c *gin.Context)
	Register(c *gin.Context)
	Login(c *gin.Context)
	DeleteUser(c *gin.Context)
	Logout(c *gin.Context)
	User(c *gin.Context)
	checkAuthorization(c *gin.Context) (string, error)
	UpdateTrack(c *gin.Context)
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

// GetAllTracks	godoc
// @Summary		Show the list of all tracks.
// @Description responds with the list of all tracks as JSON.
// @Tags		track-controller
// @Accept		*/*
// @Produce		json
// @Param       page query   int           false "Page number"
// @Param       page_size    query         int false "Number of items per page"
// @Param       sort_by      query         string false "Field to sort by (e.g., 'created_at')"
// @Param       sort_order   query         string false "Sort order ('asc' or 'desc')"
// @Param       filter       query         string false "Filter criteria ('I0001' or '=I0001')"
// @Success		200 {array}  model.Track  "OK"
// @Failure		400 {object} model.ErrorResponse "Invalid page or page_size parameters"
// @Failure		401 {object} model.ErrorResponse "Unauthorized"
// @Failure		500 {object} model.ErrorResponse "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks [get]
func (a *WebApp) GetAllTracks(c *gin.Context) {
	a.metrics.GetAllTracksCounter.Inc()
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

	// Retrieve paginated tracks from the storage
	tracks, countTotal, err := a.storage.Operations.GetTracks(offset, pageSizeInt, sortBy, sortOrder, filter)
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
		return
	}

	// Calculate total pages based on total count and page size
	totalPages := int(math.Ceil(float64(countTotal) / float64(pageSizeInt)))

	res, _ := json.Marshal(tracks)

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
	c.IndentedJSON(http.StatusOK, tracks)
	log.Debugf("Tracks response: %s", res)
}

// PostTracks	godoc
// @Summary		Adds an track from JSON.
// @Description adds an track from JSON received in the request body.
// @Tags		track-controller
// @Accept		json
// @Produce		json
// @Param		request body []model.Track true "Track details"
// @Success     201 {object} []model.Track  "Created"
// @Failure     400 {object} model.ErrorResponse  "Bad Request"
// @Failure     401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure     409 {object} model.ErrorResponse  "track code already exists"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks/add [post]
func (a *WebApp) PostTracks(c *gin.Context) {
	a.metrics.PostTracksCounter.Inc()

	var tracks []model.Track

	// Decode the request body as an array of model.Track
	if err := c.BindJSON(&tracks); err != nil {
		a.logger.Errorf("Invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}

	// Create an array to store insertion errors, if any
	var insertionErrors []error

	systemUser, errSystemUser := a.storage.Operations.FindUser(a.cfg.RESTSystemUser, "email")
	if errSystemUser != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Error find system user"})
		return
	}

	// Loop through the tracks and prepare them
	for i := range tracks {
		track := &tracks[i]

		track.ID = uuid.New()
		track.CreatedAt = time.Now()
		track.UpdatedAt = time.Now()
		track.Sender = systemUser.Name

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
		track.CreatorUser = valueUUID

		track.Title = strings.TrimSpace(track.Title)
		track.Artist = strings.TrimSpace(track.Artist)
		track.Code = strings.TrimSpace(track.Code)
		track.Description = strings.TrimSpace(track.Description)
		track.S3Version = strings.TrimSpace(track.S3Version)

		if track.Code == "" || track.Artist == "" || track.S3Version == "" {
			c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "empty required fields `Code` or `Artist` or `Path`"})
			return
		}

		// Check if the track code already exists
		_, err = a.storage.Operations.GetTracksByColumns(track.Code, "code")
		if err == nil {
			insertionErrors = append(insertionErrors, fmt.Errorf("track code already exists for track %s", track.Title))
			continue
		}
	}

	// Check if there were insertion errors
	if len(insertionErrors) > 0 {
		errorMessages := make([]string, len(insertionErrors))
		for i, err := range insertionErrors {
			errorMessages[i] = err.Error()
		}
		c.IndentedJSON(http.StatusConflict, model.ErrorResponse{Message: "Some tracks could not be inserted"})
		return
	}

	// Insert all tracks into the database
	if err := a.storage.Operations.CreateTracks(tracks); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, tracks)
}

// GetTrackByID godoc
// @Summary		Track whose ID value matches the id.
// noinspection
// @Description locates the track whose ID value matches the id parameter sent by the client,
// @Description	then returns that track as a response.
// @Tags		track-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code track"
// @Success     200 {object} model.Track  "OK"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks/{code} [get]
func (a *WebApp) GetTrackByID(c *gin.Context) {
	// Increment the session-based counter

	// If user is authorized, proceed with getting the track
	a.metrics.GetTrackByIDCounter.Inc()

	id := c.Param("code")
	result, err := a.storage.Operations.GetTracksByColumns(id, "code")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "track not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
		}
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

// GetDeleteAll godoc
// @Summary		Complete removal of all tracks.
// @Description Delete ALL.
// @Tags		track-controller
// @Accept		*/*
// @Produce		json
// @Success     204 {object} model.OkResponse   "No Content"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks/deleteAll [delete]
func (a *WebApp) GetDeleteAll(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by GetDeleteAll
	a.metrics.GetDeleteAllCounter.Inc()

	err := a.storage.Operations.DeleteTracksAll()
	if err != nil {
		a.logger.Fatal(err)
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Error Delete all Track"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// GetDeleteByID godoc
// @Summary		Deletes track whose ID value matches the code.
// @Description locates the track whose ID value matches the id parameter and deletes it.
// @Tags		track-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code track"
// @Success     204 {object} model.OkResponse   "No Content"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks/delete/{code} [delete]
func (a *WebApp) GetDeleteByID(c *gin.Context) {
	// Increment the session-based counter

	// If user is authorized, proceed with deleting the track by ID
	a.metrics.GetDeleteByIDCounter.Inc()

	code := c.Param("code")

	_, err := a.storage.Operations.GetTracksByColumns(code, "code")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "track not found"})
		} else {
			a.logger.Error(err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Internal Server Error"})
		}
		return
	}

	err = a.storage.Operations.DeleteTracks(code, "code")
	if err != nil {
		a.logger.Error(err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "error deleting track"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// UpdateTrack godoc
// @Summary                Updates an existing track with new data.
// @Description updates an existing track with new data based on the ID parameter sent by the client.
// @Tags                track-controller
// @Accept              json
// @Produce             json
// @Param               request body model.Track true "Updated track details"
// @Success     200 {object} model.Track  "OK"
// @Failure     400 {object} model.ErrorResponse  "Bad Request"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router                /tracks/update [patch]
func (a *WebApp) UpdateTrack(c *gin.Context) {
	// Increment the session-based counter

	// Increment the counter for each request handled by UpdateTrack
	a.metrics.UpdateTrackCounter.Inc()

	var newTrack model.Track

	newTrack.UpdatedAt = time.Now()

	if bindErr := c.BindJSON(&newTrack); bindErr != nil {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "invalid request payload"})
		return
	}
	newTrack.Title = strings.TrimSpace(newTrack.Title)
	newTrack.Artist = strings.TrimSpace(newTrack.Artist)
	newTrack.Code = strings.TrimSpace(newTrack.Code)
	newTrack.Description = strings.TrimSpace(newTrack.Description)

	if newTrack.Code == "" || newTrack.Artist == "" {
		c.IndentedJSON(http.StatusBadRequest, model.ErrorResponse{Message: "empty required fields `Code` or `Artist`"})
		return
	}

	existingTrack, getErr := a.storage.Operations.GetTracksByColumns(newTrack.Code, "code")
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "track not found"})
		} else {
			a.logger.Error(getErr)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: getErr.Error()})
		}
		return
	}

	if newTrack.Title != "" {
		existingTrack.Title = newTrack.Title
	}
	if newTrack.Artist != "" {
		existingTrack.Artist = newTrack.Artist
	}
	if !newTrack.Price.IsZero() {
		existingTrack.Price = newTrack.Price
	}

	if newTrack.Description != "" {
		existingTrack.Description = newTrack.Description
	}
	existingTrack.Likes = newTrack.Likes

	existingTrack.Sender = "rest"

	existingTrack.UpdatedAt = time.Now()
	// Perform the update operation
	err := a.storage.Operations.UpdateTracks(existingTrack)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, existingTrack)
}
