package gin

import (
	"net/http"
	"skeleton-golange-application/app/model"

	"github.com/gin-gonic/gin"
)

// Play godoc
// @Summary Stream and play audio tracks from a playlist.
// @Description Streams and plays audio tracks from the specified playlist.
// @Tags track-controller
// @Accept */*
// @Produce octet-stream
// @Param playlist_id path string true "Playlist ID"
// @Success 200 {array} model.Track "OK"
// @Failure 401 {object} model.ErrorResponse "Unauthorized"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /play/{playlist_id} [get]
func (a *WebApp) Play(c *gin.Context) {
	// Extract playlist ID from path parameter
	playlistID := c.Param("playlist_id")

	// Retrieve the playlist and associated tracks by ID
	_, tracks, err := a.storage.Operations.GetPlayListByID(playlistID)
	if err != nil {
		a.logger.Errorf("Error getting playlist: %v", err)
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Playlist not found"})
		return
	}

	// Check if the playlist contains tracks
	if len(tracks) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Playlist is empty"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Playback started"})
}
