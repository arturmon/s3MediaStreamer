package audio

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/playlist"
	"s3MediaStreamer/app/services/s3"
	"s3MediaStreamer/app/services/track"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel"
)

type Repository interface {
}

type Service struct {
	track    track.Service
	s3       s3.Service
	playlist playlist.Service
	logger   *logs.Logger
}

func NewAudioService(track track.Service, s3 s3.Service, playlist playlist.Service, logger *logs.Logger) *Service {
	return &Service{track, s3, playlist, logger}
}

func (h Service) StreamFileService(c *gin.Context, fileName string, f *os.File) {
	// Use a channel to signal completion
	done := make(chan struct{})
	defer close(done)

	// Use a goroutine to stream the file and check for client disconnection
	go func() {
		defer func() {
			// Clean up after streaming is done
			err := h.s3.CleanTemplateFile(fileName)
			if err != nil {
				h.logger.Errorf("Error cleaning up file: %v", err)
				return
			}
		}()

		_, err := io.Copy(c.Writer, f)
		if err != nil {
			// Log the error, but don't treat it as a critical error
			h.logger.Errorf("Error streaming audio_handler: %v", err)
			return
		}
	}()
}

func (h Service) StreamM3UReadFileService(ctx context.Context, segmentPath string) (*minio.ObjectInfo, string, *os.File, *model.Track, *model.RestError) {
	track, err := h.track.GetTracksByColumns(ctx, segmentPath, "_id")
	if err != nil {
		return nil, "", nil, nil, &model.RestError{Code: http.StatusNotFound, Err: "Segment not found"}
	}

	trackID, err := h.s3.GetS3VersionByTrackID(ctx, track.ID.String())
	if err != nil {
		return nil, "", nil, nil, &model.RestError{Code: http.StatusNotFound, Err: "Segment not found"}
	}
	findObject, err := h.s3.FindObjectFromVersion(context.Background(), trackID)
	if err != nil {
		return nil, "", nil, nil, &model.RestError{Code: http.StatusNotFound, Err: "Segment not found"}
	}

	fileName, err := h.s3.DownloadFilesS3(context.Background(), findObject.Key)
	if err != nil {
		return nil, "", nil, nil, &model.RestError{Code: http.StatusNotAcceptable, Err: "Error downloading file"}
	}
	// Open the file
	f, err := h.s3.OpenTemplateFile(fileName)
	if err != nil {
		return nil, "", nil, nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Error reading file data"}
	}
	return &findObject, fileName, f, track, nil
}

func (h *Service) GenerateM3U8Playlist(filePaths *[]model.TrackRequest) []*model.PlaylistM3U {
	var generatePlaylist []*model.PlaylistM3U

	var prefixURI = "stream/"
	for _, trackRequest := range *filePaths {
		item := trackRequest.Track // Access the embedded Track struct
		segment := &model.PlaylistM3U{
			URI:      prefixURI + item.ID.String(),
			Title:    filepath.Base(item.Artist) + " - " + filepath.Base(item.Title),
			Duration: item.Duration.Seconds(),
		}
		generatePlaylist = append(generatePlaylist, segment)
	}
	return generatePlaylist
}
func (h *Service) PlayM3UPlaylist(playlist []*model.PlaylistM3U, c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "PlayM3UPlaylist")
	defer span.End()
	c.Header("Content-Type", "application/x-mpegURL")
	// c.Header("Content-Type", "application/json")

	_, err := fmt.Fprintf(c.Writer, "#EXTM3U\n")
	if err != nil {
		h.logger.Errorf("Error writing ENDLIST information: %v", err)
		return
	}

	// Write each segment information
	for _, segment := range playlist {
		_, err = fmt.Fprintf(c.Writer, "#EXTINF:%d,%s\n%s\n", int(segment.Duration), segment.Title, segment.URI)
		if err != nil {
			h.logger.Errorf("Error writing segment information: %v", err)
			return
		}
	}
}

func (h *Service) PlayPlaylist(ctx context.Context, playlistID string) (*[]model.TrackRequest, error) {
	// Simulate the playlist retrieval

	isExist, err := h.playlist.CheckPlaylistExists(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	if !isExist {
		h.logger.Infof("Playlist not found")
		return nil, fmt.Errorf("playlist not found")
	}
	allTracks, err := h.playlist.GetPlaylistAllTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	// Return the stubbed data without any real logic
	return &allTracks, nil
}
