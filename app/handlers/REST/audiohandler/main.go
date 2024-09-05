package audiohandler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
)

type AudioServiceInterface interface {
	GenerateM3U8Playlist(filePaths *[]model.Track) []*model.PlaylistM3U
	PlayM3UPlaylist(playlist []*model.PlaylistM3U, c *gin.Context)
	PlayPlaylist(ctx context.Context, playlistID string) (*[]model.Track, error)
	StreamM3UReadFileService(ctx context.Context, segmentPath string) (*minio.ObjectInfo, string, *os.File, *model.Track, *model.RestError)
	StreamFileService(c *gin.Context, fileName string, f *os.File)
}

type Handler struct {
	audio  AudioServiceInterface
	logger *logs.Logger
}

func NewAudioHandler(audio AudioServiceInterface, logger *logs.Logger) *Handler {
	return &Handler{audio, logger}
}

// Audio godoc
// @Summary Stream audio files.
// @Description Streams audio files in the specified directory as MP3 or FLAC.
// @Tags audio-controller
// @Accept */*
// @Produce application/x-mpegURL
// @Param playlist_id path string false "Playlist ID"
// @Param control path string false "Control operation playlist play"
// @Success 200 {array} model.Track "OK"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Security ApiKeyAuth
// @Router /audio/{playlist_id} [get]
func (h *Handler) Audio(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "Audio")
	defer span.End()
	// Assuming you have a function that retrieves or generates the M3U8 playlist
	playlistID := c.Param("playlist_id")
	tracks, err := h.audio.PlayPlaylist(c.Request.Context(), playlistID)

	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, model.ErrorResponse{Message: err.Error()})
		return
	}

	if len(*tracks) == 0 {
		c.JSON(http.StatusOK, "No tracks to play")
		return
	}

	playlist := h.audio.GenerateM3U8Playlist(tracks)
	h.audio.PlayM3UPlaylist(playlist, c)
}

// StreamM3U godoc
// @Summary Stream audio files.
// @Description Streams audio files in the specified directory as MP3 or FLAC.
// @Tags audio-controller
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
func (h *Handler) StreamM3U(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "StreamM3U")
	defer span.End()
	segmentPath := c.Param("segment")
	var track *model.Track

	findObject, fileName, f, track, errValidateOTP := h.audio.StreamM3UReadFileService(c, segmentPath)
	if errValidateOTP != nil {
		c.JSON(errValidateOTP.Code, errValidateOTP.Err)
		return
	}
	defer f.Close()

	c.Header("Content-Type", findObject.Metadata.Get("Content-Type"))
	c.Header("Content-Disposition", "inline; filename="+findObject.Key)
	c.Header("Content-Length", fmt.Sprintf("%d", findObject.Size))
	c.Header("Cache-Control", "no-cache")
	c.Header("Content-Duration", fmt.Sprintf("%d", track.Duration)) // second

	h.audio.StreamFileService(c, fileName, f)
	// Wait for client disconnect notification
	<-c.Writer.CloseNotify()
	// Client disconnected, clean up and return
	h.logger.Info("Client disconnected, stopping streaming.")
}
