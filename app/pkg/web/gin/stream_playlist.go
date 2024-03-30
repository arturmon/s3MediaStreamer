package gin

import (
	"net/http"
	"s3MediaStreamer/app/model"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"

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
	_, span := otel.Tracer("").Start(c.Request.Context(), "CreatePlaylist")
	defer span.End()
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
	if err = a.storage.Operations.CreatePlayListName(c.Request.Context(), newPlaylist); err != nil {
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
// @Router /playlist/{playlist_id} [delete]
func (a *WebApp) DeletePlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "DeletePlaylist")
	defer span.End()
	// Get the playlist ID from the URL path
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists in the database
	if !a.storage.Operations.PlaylistExists(c.Request.Context(), playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
		return
	}

	// Check if the playlist is not empty
	if err := a.storage.Operations.ClearPlayList(c.Request.Context(), playlistID); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to clear playlist"})
		return
	}

	// Delete the playlist from the database
	if err := a.storage.Operations.DeletePlaylist(c.Request.Context(), playlistID); err != nil {
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
// @Router /playlist/{playlist_id}/{track_id} [post]
func (a *WebApp) AddToPlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "AddToPlaylist")
	defer span.End()
	// Extract playlist ID and track ID from path parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")

	// Check if the playlist and track exist (you should implement this)
	if !a.storage.Operations.PlaylistExists(c.Request.Context(), playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}
	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
		return
	}

	_, err = a.storage.Operations.GetTracksByColumns(c.Request.Context(), trackID, "_id")

	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
		return
	}

	// Add the track to the playlist (you should implement this)
	if err = a.storage.Operations.AddTrackToPlaylist(c.Request.Context(), playlistID, trackID); err != nil {
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
// @Router /playlist/{playlist_id}/{track_id} [delete]
func (a *WebApp) RemoveFromPlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "RemoveFromPlaylist")
	defer span.End()
	// Get the playlist ID and track ID from the request parameters
	playlistID := c.Param("playlist_id")
	trackID := c.Param("track_id")

	// Check if the playlist exists
	_, _, err := a.storage.Operations.GetPlayListByID(c.Request.Context(), playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
		return
	}

	// Check if the track exists
	_, err = a.storage.Operations.GetTracksByColumns(c.Request.Context(), trackID, "_id")
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
		return
	}

	// Remove the track from the playlist
	if err = a.storage.Operations.RemoveTrackFromPlaylist(c.Request.Context(), playlistID, trackID); err != nil {
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
// @Router /playlist/{playlist_id}/clear [delete]
func (a *WebApp) ClearPlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ClearPlaylist")
	defer span.End()
	// Get the playlist ID from the URL parameters
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists
	if !a.storage.Operations.PlaylistExists(c.Request.Context(), playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
		return
	}

	// Clear the playlist by removing all tracks (you should implement this logic)
	if err := a.storage.Operations.ClearPlayList(c.Request.Context(), playlistID); err != nil {
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
// @Router /playlist/{playlist_id} [post]
func (a *WebApp) SetFromPlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "SetFromPlaylist")
	defer span.End()
	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists (you should implement this logic)
	if !a.storage.Operations.PlaylistExists(c.Request.Context(), playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
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
		_, err := a.storage.Operations.GetTracksByColumns(c.Request.Context(), trackID, "_id")
		if err != nil {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Track not found"})
			return
		}
	}

	// Update the track order in the playlist (you should implement this logic)
	if err := a.storage.Operations.UpdatePlaylistTrackOrder(c.Request.Context(), playlistID, request.TrackOrder); err != nil {
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
func (a *WebApp) ListTracksFromPlaylist(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ListTracksFromPlaylist")
	defer span.End()
	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	// Check if the playlist exists (you should implement this logic)
	if !a.storage.Operations.PlaylistExists(c.Request.Context(), playlistID) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	userRole, userID, err := a.readUserIdAndRole(c)
	playlistCreateUser, err := a.storage.Operations.GetUserAtPlayList(c, playlistID)
	if err != nil {
		return
	}

	if userRole != "admin" || userID != playlistCreateUser {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "you are not an administrator or this is not your playlist"})
		return
	}

	// Get playlist and tracks
	playlist, tracks, err := a.storage.Operations.GetPlayListByID(c.Request.Context(), playlistID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	response := model.PlaylistTracksResponse{
		Playlist: playlist,
		Tracks:   make([]model.Track, len(tracks)),
	}

	// Iterate over tracks and create response objects
	for i, track := range tracks {
		response.Tracks[i] = model.Track{
			ID:          track.ID,
			CreatedAt:   track.CreatedAt,
			UpdatedAt:   track.UpdatedAt,
			Album:       track.Album,
			AlbumArtist: track.AlbumArtist,
			Composer:    track.Composer,
			Genre:       track.Genre,
			Lyrics:      track.Lyrics,
			Title:       track.Title,
			Artist:      track.Artist,
			Year:        track.Year,
			Comment:     track.Comment,
			Disc:        track.Disc,
			DiscTotal:   track.DiscTotal,
			Duration:    track.Duration,
			SampleRate:  track.SampleRate,
			Bitrate:     track.Bitrate,
			Sender:      track.Sender,
			CreatorUser: track.CreatorUser,
			Likes:       track.Likes,
			S3Version:   track.S3Version,
		}
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
func (a *WebApp) ListPlaylists(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ListAllPlaylist")
	defer span.End()

	userRole, userID, err := a.readUserIdAndRole(c)

	var playlists []model.PLayList

	if userRole == "admin" {
		playlists, err = a.storage.Operations.GetAllPlayList(c.Request.Context(), "admin")
	} else {
		playlists, err = a.storage.Operations.GetAllPlayList(c.Request.Context(), userID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlists not found"})
		return
	}

	response := model.PlaylistsResponse{
		PLayLists: make([]model.PLayList, len(playlists)),
	}

	for i, playlist := range playlists {
		response.PLayLists[i] = model.PLayList{
			ID:          playlist.ID,
			CreatedAt:   playlist.CreatedAt,
			Level:       playlist.Level,
			Title:       playlist.Title,
			Description: playlist.Description,
			CreatorUser: playlist.CreatorUser,
		}
	}
	// Return the response in JSON format
	c.JSON(http.StatusOK, response)
}
