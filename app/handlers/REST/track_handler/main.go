package track_handler

import (
	"fmt"
	"net/http"
	"s3MediaStreamer/app/services/track"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type TrackServiceInterface interface {
}
type TrackHandler struct {
	trackService track.TrackService
}

func NewTrackHandler(trackService track.TrackService) *TrackHandler {
	return &TrackHandler{trackService}
}

// GetAllTracks	godoc
// @Summary		Show the list of all tracks.
// @Description responds with the list of all tracks as JSON.
// @Tags		track_handler-controller
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
func (h *TrackHandler) GetAllTracks(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "GetAllTracks")
	defer span.End()

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")

	// Retrieve sorting and filtering parameters from the query
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	filter := c.DefaultQuery("filter", "")

	baseURL := "http" // По умолчанию HTTP
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		baseURL = proto
	}
	tracks, pageInt, countTotal, totalPages, err := h.trackService.GetTracksService(c, page, pageSize, filter, sortBy, sortOrder)
	if err != nil {
		c.JSON(err.Code, err.Err)
	}

	baseURL = fmt.Sprintf("%s://%s", baseURL, c.Request.Host)
	c.Header("X-Total-Count", strconv.Itoa(countTotal))
	c.Header("X-Total-Pages", strconv.Itoa(totalPages))
	c.Header("Link", generatePaginationLinks(baseURL, c.FullPath(), pageInt, totalPages, pageSize))
	c.Header("Access-Control-Expose-Headers", "X-Total-Count,X-Total-Pages,Link")
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.IndentedJSON(http.StatusOK, tracks)

}

// GetTrackByID godoc
// @Summary		Track whose ID value matches the id.
// noinspection
// @Description locates the track_handler whose ID value matches the id parameter sent by the client,
// @Description	then returns that track_handler as a response.
// @Tags		track_handler-controller
// @Accept		*/*
// @Produce		json
// @Param		code    path      string     true  "Code track_handler"
// @Success     200 {object} model.Track  "OK"
// @Failure     401 {object} model.ErrorResponse  "Unauthorized"
// @Failure     404 {object} model.ErrorResponse  "Not Found"
// @Failure     500 {object} model.ErrorResponse  "Internal Server Error"
// @Security    ApiKeyAuth
// @Router		/tracks/{code} [get]
func (h *TrackHandler) GetTrackByID(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "GetTrackByID")
	defer span.End()
	id := c.Param("code")
	// Increment the session-based counter
	result, err := h.trackService.GetTrackByID(c, id)
	if err != nil {
		c.JSON(err.Code, err.Err)
	}
	c.IndentedJSON(http.StatusOK, result)
}
