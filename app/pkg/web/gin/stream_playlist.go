package gin

import (
	"net/http"
	"skeleton-golange-application/app/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
func (a *WebApp) CreatePlaylist(c *gin.Context) {
	// Define a struct to parse the request body
	var playlistRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Level       string `json:"level"`
	}

	// Parse the JSON request body
	if err := c.ShouldBindJSON(&playlistRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Convert the 'Level' field from string to int64
	level, err := strconv.ParseInt(playlistRequest.Level, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'level' format"})
		return
	}

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

	// Generate a unique ID for the new playlist (you can use your own method)
	playlistID := uuid.New()

	// Create a new playlist in the database
	newPlaylist := model.PLayList{
		ID:          playlistID,
		CreatedAt:   time.Now(),
		Level:       level,
		Title:       playlistRequest.Title,
		Description: playlistRequest.Description,
		CreatorUser: valueUUID,
	}

	// Save the new playlist in the database
	if err = a.storage.Operations.CreatePlayListName(newPlaylist); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create playlist"})
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
// @Router /playlist/delete/{id} [delete]
func (a *WebApp) DeletePlaylist(c *gin.Context) {
	// Get the playlist ID from the URL path
	playlistID := c.Param("id")

	// Check if the playlist exists in the database
	if !a.storage.Operations.PlaylistExists(playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	// Check if the playlist is not empty
	if err := a.storage.Operations.ClearPlayList(playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to clear playlist"})
		return
	}

	// Delete the playlist from the database
	if err := a.storage.Operations.DeletePlaylist(playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to delete playlist"})
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
// @Router /playlist/{playlist_id}/add/track/{track_id} [post]
func (a *WebApp) AddToPlaylist(c *gin.Context) {
	// Extract playlist ID and track ID from path parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")

	// Check if the playlist and track exist (you should implement this)
	if !a.storage.Operations.PlaylistExists(playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	_, err := a.storage.Operations.GetTracksByColumns(trackID, "_id")

	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
		return
	}

	// Add the track to the playlist (you should implement this)
	if err = a.storage.Operations.AddTrackToPlaylist(playlistID, trackID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to add track to playlist"})
		return
	}

	// Return a success response
	c.JSON(http.StatusCreated, gin.H{"message": "Track added to the playlist successfully"})
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
// @Router /playlist/{playlist_id}/track/{track_id} [delete]
func (a *WebApp) RemoveFromPlaylist(c *gin.Context) {
	// Get the playlist ID and track ID from the request parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")

	// Check if the playlist exists
	_, _, err := a.storage.Operations.GetPlayListByID(playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	// Check if the track exists
	_, err = a.storage.Operations.GetTracksByColumns(trackID, "_id")
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
		return
	}

	// Remove the track from the playlist
	if err = a.storage.Operations.RemoveTrackFromPlaylist(playlistID, trackID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to remove track from playlist"})
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
// @Router /playlists/{playlist_id}/clear [delete]
func (a *WebApp) ClearPlaylist(c *gin.Context) {
	// Get the playlist ID from the URL parameters
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists
	if !a.storage.Operations.PlaylistExists(playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	// Clear the playlist by removing all tracks (you should implement this logic)
	if err := a.storage.Operations.ClearPlayList(playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to clear playlist"})
		return
	}

	// Return a success response
	c.IndentedJSON(http.StatusNoContent, model.OkResponse{Message: "OK"})
}

// SetFromPlaylist godoc
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
// @Router /playlist/{playlist_id}/set [post]
func (a *WebApp) SetFromPlaylist(c *gin.Context) {
	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists (you should implement this logic)
	if !a.storage.Operations.PlaylistExists(playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	// Define a variable to hold the request data
	var request SetPlaylistTrackOrderRequest

	// Parse the JSON request body into the request struct
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the provided track order contains valid track IDs (you should implement this logic)
	for _, trackID := range request.TrackOrder {
		_, err := a.storage.Operations.GetTracksByColumns(trackID, "_id")
		if err != nil {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
			return
		}
	}

	// Update the track order in the playlist (you should implement this logic)
	if err := a.storage.Operations.UpdatePlaylistTrackOrder(playlistID, request.TrackOrder); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to update track order"})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "Track order updated successfully"})
}

// SetPlaylistTrackOrderRequest Define a struct to match the expected JSON structure.
type SetPlaylistTrackOrderRequest struct {
	TrackOrder []string `json:"track_order"`
}
