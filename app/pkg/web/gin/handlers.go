package gin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"math"
	"net/http"
	"skeleton-golange-application/app/model"
	"strconv"
)

type Handler interface {
	GetAllTracks(c *gin.Context)
	GetTrackByID(c *gin.Context)
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
