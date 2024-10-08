package playlisthandler

import (
	"net/http"
	"s3MediaStreamer/app/handlers/REST/userhandler"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/playlist"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type PlaylistServiceInterface interface {
}

type Handler struct {
	playlistService playlist.Service
	userHandler     userhandler.Handler
}

func NewPlaylistHandler(playlistService playlist.Service, userHandler userhandler.Handler) *Handler {
	return &Handler{playlistService, userHandler}
}

// CreatePlaylist godoc
// @Summary Create a new playlist.
// @Description Creates a new playlist with the provided information.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param request body []model.PLayList true "PLayList details"
// @Success 201 {object} []model.PLayList "Playlist created successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid input"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/create [post]
func (h *Handler) CreatePlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "CreatePlaylist")
	defer span.End()

	// Parse the JSON request body
	if err := c.ShouldBindJSON(&model.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newPlaylist, err := h.playlistService.CreateNewPlaylist(c)
	if err != nil {
		c.JSON(err.Code, err.Err)
		return
	}
	// Return a success response
	c.IndentedJSON(http.StatusCreated, newPlaylist)
}

// DeletePlaylist godoc
// @Summary Delete a playlist by ID.
// @Description Delete a playlist based on its unique ID.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param id path string true "Playlist ID"
// @Success 204 "No Content"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 404 {object} model.ErrorResponse "Not Found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id} [delete]
func (h *Handler) DeletePlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "DeletePlaylist")
	defer span.End()
	// Get the playlist ID from the URL path
	playlistID := c.Param("playlist_id")
	errDeletePlaylist := h.playlistService.DeletePlaylistForUser(c, userContext.UserRole, userContext.UserID, playlistID)
	if errDeletePlaylist != nil {
		c.JSON(errDeletePlaylist.Code, errDeletePlaylist.Err)
		return
	}
	// Return a success response
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// AddToPlaylist godoc
// @Summary Add a track to a playlist
// @Description Add a track to an existing playlist.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param playlist_id path string true "Playlist ID"
// @Param track_id path string true "Track ID"
// @Success 201 {string} string "Track added to the playlist successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 404 {object} model.ErrorResponse "Playlist or track not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id}/{track_id} [post]
func (h *Handler) AddToPlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "AddToPlaylist")
	defer span.End()
	// Extract playlist ID and track ID from path parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")
	parentID := c.DefaultQuery("parent_id", "")

	errAddToPlaylist := h.playlistService.AddTrackToPlaylist(
		c,
		userContext.UserRole,
		userContext.UserID,
		playlistID,
		trackID,
		parentID,
	)
	if errAddToPlaylist != nil {
		c.JSON(errAddToPlaylist.Code, errAddToPlaylist.Err)
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{"message": "Track or playlist added to the playlist successfully"})
}

// ListTracksFromPlaylist godoc
// @Summary Get tracks from a playlist.
// @Description Get tracks from a playlist by providing the playlist ID.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param playlist_id path string true "Playlist ID"
// @Success 200 {object} model.PlaylistTracksResponse "Tracks retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id} [get]
func (h *Handler) ListTracksFromPlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ListTracksFromPlaylist")
	defer span.End()

	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	response, errListTracksFromPlaylist := h.playlistService.GetTracksInPlaylist(
		c,
		userContext.UserRole,
		userContext.UserID,
		playlistID,
	)
	if errListTracksFromPlaylist != nil {
		c.JSON(errListTracksFromPlaylist.Code, errListTracksFromPlaylist.Err)
		return
	}

	// Return the response in JSON format
	c.JSON(http.StatusOK, response)
}

// ListPlaylists godoc
// @Summary Get all playlists
// @Description Retrieves all playlists available in the storage.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Success 200 {object} model.PlaylistsResponse "Playlists retrieved successfully"
// @Failure 404 {object} model.ErrorResponse "Playlists not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/get [get]
func (h *Handler) ListPlaylists(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ListAllPlaylist")
	defer span.End()

	response, errListTracksFromPlaylist := h.playlistService.GetUserPlaylists(
		c,
		userContext.UserRole,
		userContext.UserID,
	)
	if errListTracksFromPlaylist != nil {
		c.JSON(errListTracksFromPlaylist.Code, errListTracksFromPlaylist.Err)
		return
	}

	// Return the response in JSON format
	c.JSON(http.StatusOK, response)
}

// RemoveFromPlaylist godoc
// @Summary Remove a track from the playlist.
// @Description Remove a track from the specified playlist.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param playlist_id path string true "Playlist ID"
// @Param track_id path string true "Track ID"
// @Success 200 {string} string "Track removed from playlist successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 404 {object} model.ErrorResponse "Playlist or track not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id}/{track_id} [delete]
func (h *Handler) RemoveFromPlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "RemoveFromPlaylist")
	defer span.End()
	// Get the playlist ID and track ID from the request parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")

	errRemoveFromPlaylist := h.playlistService.RemoveTrackFromPlaylist(
		c,
		userContext.UserRole,
		userContext.UserID,
		playlistID,
		trackID,
	)
	if errRemoveFromPlaylist != nil {
		c.JSON(errRemoveFromPlaylist.Code, errRemoveFromPlaylist.Err)
		return
	}
	// Return a success response
	c.JSON(http.StatusOK, "Track removed from playlist successfully")
}

// ClearPlaylist godoc
// @Summary Clear a playlist by removing all tracks from it.
// @Description Removes all tracks from a playlist, effectively clearing it.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param playlist_id path string true "Playlist ID"
// @Success 200 {string} string "Playlist cleared successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 404 {object} model.ErrorResponse "Playlist not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id}/clear [delete]
func (h *Handler) ClearPlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ClearPlaylist")
	defer span.End()
	// Get the playlist ID from the URL parameters
	playlistID := c.Param("playlist_id")

	errRemoveFromPlaylist := h.playlistService.ClearAllTracksInPlaylist(
		c,
		userContext.UserRole,
		userContext.UserID,
		playlistID,
	)
	if errRemoveFromPlaylist != nil {
		c.JSON(errRemoveFromPlaylist.Code, errRemoveFromPlaylist.Err)
		return
	}

	// Return a success response
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// AddTracksToPlaylist godoc
// @Summary Set tracks in a playlist.
// @Description Set tracks in a playlist by providing a list of track IDs.
// @Tags playlist-controller
// @Accept json
// @Produce json
// @Param playlist_id path string true "Playlist ID"
// @Param track_ids body []string true "List of track IDs to set in the playlist"
// @Success 200 {string} string "Tracks set in the playlist successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /playlist/{playlist_id}/tracks [post]
func (h *Handler) AddTracksToPlaylist(c *gin.Context, userContext *model.UserContext) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "AddTracksToPlaylist")
	defer span.End()

	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	// Parse the JSON request body into the request struct
	var request model.SetPlaylistTrackOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service function
	if err := h.playlistService.AddTracksToPlaylist(
		c.Request.Context(),
		userContext.UserRole,
		userContext.UserID,
		playlistID,
		&request,
		true,
	); err != nil {
		c.JSON(err.Code, model.ErrorResponse{Message: err.Err})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Tracks set in the playlist successfully"})
}
