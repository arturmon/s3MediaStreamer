package track

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Repository interface {
	CreateTracks(ctx context.Context, list []model.Track) error
	GetTracks(ctx context.Context, offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Track, int, error)
	GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error)
	CleanTracks(ctx context.Context) error
	DeleteTracksAll(ctx context.Context) error
	UpdateTracks(ctx context.Context, track *model.Track) error
	GetAllTracks(ctx context.Context) ([]model.Track, error)
	AddTrackToPlaylist(ctx context.Context, playlistID, referenceID, referenceType string) error
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error
	GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error)
}

type Service struct {
	trackRepository Repository
	logger          *logs.Logger
}

func NewTrackService(trackRepository Repository) *Service {
	return &Service{trackRepository: trackRepository}
}

func (s *Service) CreateTracks(ctx context.Context, list []model.Track) error {
	return s.trackRepository.CreateTracks(ctx, list)
}

func (s *Service) GetTracks(ctx context.Context, offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Track, int, error) {
	return s.trackRepository.GetTracks(ctx, offset, limit, sortBy, sortOrder, filterArtist)
}

func (s *Service) GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error) {
	return s.trackRepository.GetTracksByColumns(ctx, code, columns)
}

func (s *Service) CleanTracks(ctx context.Context) error {
	return s.trackRepository.CleanTracks(ctx)
}

func (s *Service) DeleteTracksAll(ctx context.Context) error {
	return s.trackRepository.DeleteTracksAll(ctx)
}

func (s *Service) UpdateTracks(ctx context.Context, track *model.Track) error {
	return s.trackRepository.UpdateTracks(ctx, track)
}

func (s *Service) GetAllTracks(ctx context.Context) ([]model.Track, error) {
	return s.trackRepository.GetAllTracks(ctx)
}

func (s *Service) AddTrackToPlaylist(ctx context.Context, playlistID, referenceID, referenceType string) error {
	return s.trackRepository.AddTrackToPlaylist(ctx, playlistID, referenceID, referenceType)
}

func (s *Service) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error {
	return s.trackRepository.RemoveTrackFromPlaylist(ctx, playlistID, trackID)
}

func (s *Service) GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error) {
	return s.trackRepository.GetAllTracksByPositions(ctx, playlistID)
}

func (s *Service) GetTracksService(c *gin.Context, page, pageSize, filter string, sortBy, sortOrder string) ([]model.Track, int, int, int, *model.RestError) {
	// Convert page, pageSize, and totalPages to integers
	pageInt, errPage := strconv.Atoi(page)
	pageSizeInt, errPageSize := strconv.Atoi(pageSize)
	if errPage != nil || errPageSize != nil {
		s.logger.Error("Invalid page or page_size parameters")
		return nil, 0, 0, 0, &model.RestError{Code: http.StatusBadRequest, Err: "invalid page or page_size parameters"}
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
	tracks, countTotal, err := s.trackRepository.GetTracks(c.Request.Context(), offset, pageSizeInt, sortBy, sortOrder, filter)
	if err != nil {
		s.logger.Error(err)

		return nil, 0, 0, 0, &model.RestError{Code: http.StatusInternalServerError, Err: "Internal Server Error"}
	}

	// Calculate total pages based on total count and page size
	totalPages := int(math.Ceil(float64(countTotal) / float64(pageSizeInt)))

	res, _ := json.Marshal(tracks)
	s.logger.Debugf("Tracks response: %s", res)
	return tracks, countTotal, pageInt, totalPages, nil
}

func (s *Service) GetTrackByID(c *gin.Context, id string) (*model.Track, *model.RestError) {
	result, err := s.trackRepository.GetTracksByColumns(c.Request.Context(), id, "code")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &model.RestError{Code: http.StatusNotFound, Err: "track_handler not found"}
		}
		s.logger.Error(err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Internal Server Error"}
	}

	return result, nil
}
