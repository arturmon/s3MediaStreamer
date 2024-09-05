package playlist

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/session"
	"s3MediaStreamer/app/services/track"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Repository interface {
	CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error
	GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error)
	DeletePlaylist(ctx context.Context, playlistID string) error
	PlaylistExists(ctx context.Context, playlistID string) bool
	ClearPlayList(ctx context.Context, playlistID string) error
	UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error
	GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error)
	GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error)
	GetUserAtPlayList(ctx context.Context, playlistID string) (string, error)
}

type Service struct {
	trackRepository    track.Repository
	playlistRepository Repository
	session            session.Service
	authService        auth.Service
	logger             *logs.Logger
}

func NewPlaylistService(trackRepository track.Repository,
	playlistRepository Repository,
	session session.Service,
	authService auth.Service,
	logger *logs.Logger) *Service {
	return &Service{trackRepository,
		playlistRepository,
		session,
		authService,
		logger}
}

const adminPolicy = "admin"

func (s *Service) CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error {
	return s.playlistRepository.CreatePlayListName(ctx, newPlaylist)
}

func (s *Service) CreatePlaylist(c *gin.Context) (*model.PLayList, *model.RestError) {
	// Convert the 'Level' field from string to int64
	level, err := strconv.ParseInt(model.Request.Level, 10, 64)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusBadRequest, Err: "Invalid 'level' format"}
	}

	// Read user_id from the session
	value, err := s.session.GetSessionKey(c, "user_id")
	if err != nil {
		s.logger.Errorf("Error getting session value: %v", err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "could not get session value"}
	}

	valueUUID, err := uuid.Parse(value.(string))
	if err != nil {
		s.logger.Errorf("Error: %v", err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "error converting value"}
	}

	// Generate a unique ID for the new playlist_handler (you can use your own method)
	playlistID := uuid.New()

	// Create a new playlist_handler in the database
	newPlaylist := model.PLayList{
		ID:          playlistID,
		CreatedAt:   time.Now(),
		Level:       level,
		Title:       model.Request.Title,
		Description: model.Request.Description,
		CreatorUser: valueUUID,
	}

	// Save the new playlist_handler in the database
	if err = s.CreatePlayListName(c.Request.Context(), newPlaylist); err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to create playlist_handler"}
	}
	return &newPlaylist, nil
}

func (s *Service) GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error) {
	return s.playlistRepository.GetPlayListByID(ctx, playlistID)
}

func (s *Service) DeletePlaylist(ctx context.Context, playlistID string) error {
	return s.playlistRepository.DeletePlaylist(ctx, playlistID)
}

func (s *Service) DeletePlaylistService(ctx context.Context, userRole, userID, playlistID string) *model.RestError {
	// Check if the playlist_handler exists in the database
	if !s.PlaylistExists(ctx, playlistID) {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	// Check if the playlist_handler is not empty
	if err = s.ClearPlayList(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to clear playlist"}
	}

	// Delete the playlist_handler from the database
	if err = s.DeletePlaylist(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to delete playlist"}
	}

	return nil
}

func (s *Service) PlaylistExists(ctx context.Context, playlistID string) bool {
	return s.playlistRepository.PlaylistExists(ctx, playlistID)
}

func (s *Service) ClearPlayList(ctx context.Context, playlistID string) error {
	return s.playlistRepository.ClearPlayList(ctx, playlistID)
}

func (s *Service) UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error {
	return s.playlistRepository.UpdatePlaylistTrackOrder(ctx, playlistID, trackOrderRequest)
}

func (s *Service) GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error) {
	return s.playlistRepository.GetTracksByPlaylist(ctx, playlistID)
}

func (s *Service) GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error) {
	return s.playlistRepository.GetAllPlayList(ctx, creatorUserID)
}

func (s *Service) GetUserAtPlayList(ctx context.Context, playlistID string) (string, error) {
	return s.playlistRepository.GetUserAtPlayList(ctx, playlistID)
}

func (s *Service) AddToPlaylist(ctx context.Context, userRole, userID, playlistID, trackID string) *model.RestError {
	// Check if the playlist_handler and track_handler exist (you should implement this)
	if !s.PlaylistExists(ctx, playlistID) {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	_, err = s.trackRepository.GetTracksByColumns(ctx, trackID, "_id")

	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "Track not found"}
	}

	// Add the track_handler to the playlist_handler (you should implement this)
	// TODO track_handler
	if err = s.trackRepository.AddTrackToPlaylist(ctx, playlistID, trackID, "track"); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to add track_handler to playlist"}
	}
	return nil
}

func (s *Service) RemoveFromPlaylist(ctx context.Context, userRole, userID, playlistID, trackID string) *model.RestError {
	// Check if the playlist exists
	_, _, err := s.GetPlayListByID(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	// Check if the track_handler exists
	_, err = s.trackRepository.GetTracksByColumns(ctx, trackID, "_id")
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "Track not found"}
	}

	// Remove the track_handler from the playlist
	if err = s.trackRepository.RemoveTrackFromPlaylist(ctx, playlistID, trackID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to remove track_handler from playlist"}
	}
	return nil
}

func (s Service) ClearPlaylistService(ctx context.Context, userRole, userID, playlistID string) *model.RestError {
	// Check if the playlist exists
	if !s.PlaylistExists(ctx, playlistID) {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	// Clear the playlist by removing all tracks (you should implement this logic)
	if err = s.playlistRepository.ClearPlayList(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to clear playlist"}
	}

	return nil
}

func (s Service) SetFromPlaylistService(ctx context.Context, userRole, userID, playlistID string, request *model.SetPlaylistTrackOrderRequest) *model.RestError {
	// Check if the playlist exists
	if !s.PlaylistExists(ctx, playlistID) {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	// Check if the provided track_handler order contains valid track_handler IDs (you should implement this logic)
	for _, trackID := range request.TrackOrder {
		_, err = s.trackRepository.GetTracksByColumns(ctx, trackID, "_id")
		if err != nil {
			return &model.RestError{Code: http.StatusNotFound, Err: "Track not found"}
		}
	}

	// Update the track_handler order in the playlist (you should implement this logic)
	if err = s.UpdatePlaylistTrackOrder(ctx, playlistID, request.TrackOrder); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to update track_handler order"}
	}

	return nil
}

func (s Service) ListTracksFromPlaylistService(ctx context.Context, userRole, userID, playlistID string) (*model.PlaylistTracksResponse, *model.RestError) {
	// Check if the playlist exists (you should implement this logic)
	if !s.PlaylistExists(ctx, playlistID) {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	playlistCreateUser, err := s.GetUserAtPlayList(ctx, playlistID)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "get user at playlist"}
	}

	if userRole != adminPolicy && userID != playlistCreateUser {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "you are not an administrator or this is not your playlist"}
	}

	// Get playlist and tracks
	playlist, tracks, err := s.GetPlayListByID(ctx, playlistID)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
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
		}
	}

	return &response, nil
}

func (s Service) ListPlaylistsService(ctx context.Context, userRole, userID string) (*model.PlaylistsResponse, *model.RestError) {
	var playlists []model.PLayList
	var err error

	if userRole == adminPolicy {
		playlists, err = s.GetAllPlayList(ctx, adminPolicy)
	} else {
		playlists, err = s.GetAllPlayList(ctx, userID)
	}
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "Playlists not found"}
	}

	response := &model.PlaylistsResponse{
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

	return response, nil
}
