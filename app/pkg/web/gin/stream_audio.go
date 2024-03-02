package gin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"skeleton-golange-application/app/model"

	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
)

// Audio godoc
// @Summary Stream audio files.
// @Description Streams audio files in the specified directory as MP3 or FLAC.
// @Tags track-controller
// @Accept */*
// @Produce application/x-mpegURL
// @Param playlist_id path string false "Playlist ID"
// @Param control path string false "Control operation playlist play"
// @Success 200 {array} model.Track "OK"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /audio/{playlist_id} [get]
func (a *WebApp) Audio(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Audio")
	defer span.End()
	// Assuming you have a function that retrieves or generates the M3U8 playlist
	playlistID := c.Param("playlist_id")
	tracks, err := a.playPlaylist(c.Request.Context(), playlistID)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	if len(*tracks) == 0 {
		c.JSON(http.StatusOK, "No tracks to play")
		return
	}

	playlist := a.generateM3U8Playlist(tracks)
	a.PlayM3UPlaylist(playlist, c)
}

// StreamM3U godoc
// @Summary Stream audio files.
// @Description Streams audio files in the specified directory as MP3 or FLAC.
// @Tags track-controller
// @Accept */*
// @Produce audio/mpeg
// @Produce audio/flac
// @Produce application/octet-stream
// @Param playlist_id path string false "Playlist ID"
// @Param control path string false "Control operation playlist play"
// @Success 200 {array} model.Track "OK"
// @Failure 404 {object} model.ErrorResponse "Segment not found"
// @Failure 406 {object} model.ErrorResponse "Segment not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /audio/stream/{segment} [get]
func (a *WebApp) StreamM3U(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "StreamM3U")
	defer span.End()
	segmentPath := c.Param("segment")
	var track *model.Track

	track, err := a.storage.Operations.GetTracksByColumns(c.Request.Context(), segmentPath, "_id")
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	findObject, err := a.S3.FindObjectFromVersion(context.Background(), track.S3Version)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
		return
	}

	fileName, err := a.S3.DownloadFilesS3(context.Background(), findObject.Key)
	if err != nil {
		c.IndentedJSON(http.StatusNotAcceptable, gin.H{"error": "Error downloading file"})
		return
	}

	c.Header("Content-Type", findObject.Metadata.Get("Content-Type"))
	c.Header("Content-Disposition", "inline; filename="+findObject.Key)
	c.Header("Content-Length", fmt.Sprintf("%d", findObject.Size))
	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Duration", fmt.Sprintf("%d", track.Duration)) // second

	// Open the file
	f, err := a.S3.OpenTemplateFile(fileName)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error reading file data"})
		return
	}
	defer f.Close()

	// Use a channel to signal completion
	done := make(chan struct{})
	defer close(done)

	// Use a goroutine to stream the file and check for client disconnection
	go func() {
		defer func() {
			// Clean up after streaming is done
			err = a.S3.CleanTemplateFile(fileName)
			if err != nil {
				a.logger.Errorf("Error cleaning up file: %v", err)
			}
		}()

		_, err = io.Copy(c.Writer, f)
		if err != nil {
			// Log the error, but don't treat it as a critical error
			a.logger.Errorf("Error streaming audio: %v", err)
		}
	}()

	// Wait for client disconnect notification
	<-c.Writer.CloseNotify()
	// Client disconnected, clean up and return
	a.logger.Info("Client disconnected, stopping streaming.")
}

func (a *WebApp) generateM3U8Playlist(filePaths *[]model.Track) []*model.PlaylistM3U {
	var playlist []*model.PlaylistM3U

	var prefixURI = "stream/"
	for _, track := range *filePaths {
		segment := &model.PlaylistM3U{
			URI:      prefixURI + track.ID.String(),
			Title:    filepath.Base(track.Artist) + " - " + filepath.Base(track.Title),
			Duration: track.Duration.Seconds(),
		}
		playlist = append(playlist, segment)
	}
	return playlist
}

func (a *WebApp) PlayM3UPlaylist(playlist []*model.PlaylistM3U, c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "PlayM3UPlaylist")
	defer span.End()
	c.Header("Content-Type", "application/x-mpegURL")
	// c.Header("Content-Type", "application/json")

	_, err := fmt.Fprintf(c.Writer, "#EXTM3U\n")
	if err != nil {
		a.logger.Errorf("Error writing ENDLIST information: %v", err)
		return
	}

	// Write each segment information
	for _, segment := range playlist {
		_, err = fmt.Fprintf(c.Writer, "#EXTINF:%d,%s\n%s\n", int(segment.Duration), segment.Title, segment.URI)
		if err != nil {
			a.logger.Errorf("Error writing segment information: %v", err)
			return
		}
	}
}

func (a *WebApp) playPlaylist(ctx context.Context, playlistID string) (*[]model.Track, error) {
	// Get the playlist and its tracks

	playlist, _, err := a.storage.Operations.GetPlayListByID(context.Background(), playlistID)
	if err != nil {
		return nil, err
	}

	sortTracks, err := a.storage.Operations.GetAllTracksByPositions(ctx, playlist.ID.String())
	if err != nil {
		return nil, err
	}

	// If there is no previous track to play, return an error or handle it as needed
	return &sortTracks, nil
}
