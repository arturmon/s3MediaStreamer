package playlist

import (
	"context"
	"fmt"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/session"
	"s3MediaStreamer/app/services/track"
	"s3MediaStreamer/app/services/tree"
	"s3MediaStreamer/app/services/user"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type Repository interface {
	CheckPlaylistExists(ctx context.Context, playlistID string) (bool, error)
	CreatePlayListName(ctx context.Context, playlist model.PLayList) error
	FetchPlaylistInfo(ctx context.Context, playlistID string, playlist *model.PLayList) error
	GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error)
	ClearPlaylistContents(ctx context.Context, playlistID string) error
	DeletePlaylist(ctx context.Context, playlistID string) error
	UpdatePlaylistDetails(ctx context.Context, playlistID, title, description string) error
	GetPlaylistOwner(ctx context.Context, playlistID string) (uuid.UUID, error)
	GetPlaylists(ctx context.Context, userID string) ([]model.PLayList, error)
	GetPlaylistAllTracks(ctx context.Context, playlistID string) ([]model.TrackRequest, error)
	GetPlaylistPath(ctx context.Context, playlistID string) (string, error)
}

type Service struct {
	trackRepository    track.Repository
	playlistRepository Repository
	session            session.Service
	authService        auth.Service
	logger             *logs.Logger
	user               user.Service
	tree               *tree.Service
}

func (s *Service) CheckPlaylistExists(ctx context.Context, playlistID string) (bool, error) {
	return s.playlistRepository.CheckPlaylistExists(ctx, playlistID)
}

func (s *Service) CreatePlayListName(ctx context.Context, playlist model.PLayList) error {
	return s.playlistRepository.CreatePlayListName(ctx, playlist)
}

func (s *Service) FetchPlaylistInfo(ctx context.Context, playlistID string, playlist *model.PLayList) error {
	return s.playlistRepository.FetchPlaylistInfo(ctx, playlistID, playlist)
}

func (s *Service) GetTracksByPlaylist(ctx context.Context, playlistID string) ([]model.Track, error) {
	return s.playlistRepository.GetTracksByPlaylist(ctx, playlistID)
}

func (s *Service) ClearPlaylistContents(ctx context.Context, playlistID string) error {
	return s.playlistRepository.ClearPlaylistContents(ctx, playlistID)
}

func (s *Service) DeletePlaylist(ctx context.Context, playlistID string) error {
	return s.playlistRepository.DeletePlaylist(ctx, playlistID)
}

func (s *Service) UpdatePlaylistDetails(ctx context.Context, playlistID, title, description string) error {
	return s.playlistRepository.UpdatePlaylistDetails(ctx, playlistID, title, description)
}

func (s *Service) GetPlaylistOwner(ctx context.Context, playlistID string) (uuid.UUID, error) {
	return s.playlistRepository.GetPlaylistOwner(ctx, playlistID)
}

func (s *Service) GetPlaylists(ctx context.Context, userID string) ([]model.PLayList, error) {
	return s.playlistRepository.GetPlaylists(ctx, userID)
}

func (s *Service) GetPlaylistAllTracks(ctx context.Context, playlistID string) ([]model.TrackRequest, error) {
	return s.playlistRepository.GetPlaylistAllTracks(ctx, playlistID)
}

func (s *Service) GetPlaylistPath(ctx context.Context, playlistID string) (string, error) {
	return s.playlistRepository.GetPlaylistPath(ctx, playlistID)
}

func NewPlaylistService(trackRepository track.Repository,
	playlistRepository Repository,
	session session.Service,
	authService auth.Service,
	user user.Service,
	logger *logs.Logger,
	tree *tree.Service) *Service {
	return &Service{
		trackRepository,
		playlistRepository,
		session,
		authService,
		logger,
		user,
		tree}
}

func (s *Service) CreateNewPlaylist(c *gin.Context) (*model.PLayList, *model.RestError) {
	// Read user_id from the session
	value, err := s.session.GetSessionKey(c, "user_id")
	if err != nil {
		s.logger.Errorf("Error getting session value: %v", err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "could not get session value"}
	}

	valueUUID, err := uuid.Parse(value.(string))
	if err != nil {
		s.logger.Errorf("Error parsing user ID: %v", err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "error converting value"}
	}

	// Create a new playlist_handler in the database
	newPlaylist := model.PLayList{
		ID:          uuid.New(),
		CreatedAt:   time.Now(),
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

func (s *Service) DeletePlaylistForUser(ctx context.Context, userRole, userID, playlistID string) *model.RestError {
	// Check if the playlist exists in the repository
	_, restErr := s.ensurePlaylistExists(ctx, playlistID)
	if restErr != nil {
		return restErr
	}

	// Validate user authorization
	restErr = s.isAuthorizedForPlaylist(ctx, userRole, userID, playlistID)
	if restErr != nil {
		return restErr
	}

	// Clear playlist contents
	if err := s.ClearPlaylistContents(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to clear playlist contents"}
	}

	// Delete the playlist
	if err := s.DeletePlaylist(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to delete playlist"}
	}

	return nil
}

func (s *Service) AddTrackToPlaylist(ctx context.Context, userRole, userID, playlistID, referenceID, parentID string) *model.RestError {
	// Check if the playlist exists
	_, restErr := s.ensurePlaylistExists(ctx, playlistID)
	if restErr != nil {
		return restErr
	}

	// Validate user authorization
	restErr = s.isAuthorizedForPlaylist(ctx, userRole, userID, playlistID)
	if restErr != nil {
		return restErr
	}

	// Determine if the reference is a track or playlist
	var referenceType string
	_, errTrack := s.trackRepository.GetTracksByColumns(ctx, referenceID, "_id")
	isPlaylist, errPlaylist := s.CheckPlaylistExists(ctx, referenceID)

	switch {
	case errTrack == nil:
		referenceType = "track"
	case isPlaylist && errPlaylist == nil:
		referenceType = "playlist"
	default:
		// Neither a track nor a playlist was found
		return &model.RestError{Code: http.StatusNotFound, Err: "Reference (track or playlist) not found"}
	}

	// Get the parent path
	var parentPath string
	if parentID != "" {
		// If the parent is provided, retrieve its path
		_, err := s.GetPlaylistPath(ctx, parentID)
		if err != nil {
			return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to retrieve parent playlist path"}
		}
	} else {
		// If no parent is provided, this is a root-level item
		parentPath = playlistID
	}

	// Delegate the task of adding the track or playlist to the repository layer
	err := s.trackRepository.AddTrackToPlaylist(ctx, playlistID, referenceType, referenceID, parentPath) // Pass parentPath to the repository
	if err != nil {
		s.logger.Error(err)
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to add reference to playlist"}
	}

	return nil
}

func (s Service) GetTracksInPlaylist(ctx context.Context, userRole, userID, playlistID string) ([]model.TrackRequest, *model.RestError) {
	// Validate user authorization
	restErr := s.isAuthorizedForPlaylist(ctx, userRole, userID, playlistID)
	if restErr != nil {
		return nil, restErr
	}

	playlistContents, err := s.GetPlaylistAllTracks(ctx, playlistID)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: fmt.Sprintf("Error GetTracksInPlaylist")}
	}

	return playlistContents, nil
}

func (s Service) GetUserPlaylists(ctx context.Context, userRole, userID string) (*model.PlaylistsResponse, *model.RestError) {
	var playlists []model.PLayList
	var err error

	if userRole == adminPolicy {
		playlists, err = s.GetPlaylists(ctx, "")
	} else {
		playlists, err = s.GetPlaylists(ctx, userID)
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
			Title:       playlist.Title,
			Description: playlist.Description,
			CreatorUser: playlist.CreatorUser,
		}
	}

	return response, nil
}

func (s *Service) RemoveTrackFromPlaylist(ctx context.Context, userRole, userID, playlistID, trackID string) *model.RestError {
	// Check if the playlist exists
	_, restErr := s.ensurePlaylistExists(ctx, playlistID)
	if restErr != nil {
		return restErr
	}

	// Validate user authorization
	restErr = s.isAuthorizedForPlaylist(ctx, userRole, userID, playlistID)
	if restErr != nil {
		return restErr
	}

	// Check if the track_handler exists
	_, err := s.trackRepository.GetTracksByColumns(ctx, trackID, "_id")
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "Track not found"}
	}

	// Remove the track_handler from the playlist
	if err = s.trackRepository.RemoveTrackFromPlaylist(ctx, playlistID, trackID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to remove track_handler from playlist"}
	}

	return nil
}

func (s Service) ClearAllTracksInPlaylist(ctx context.Context, userRole, userID, playlistID string) *model.RestError {
	// Check if the playlist exists
	_, restErr := s.ensurePlaylistExists(ctx, playlistID)
	if restErr != nil {
		return restErr
	}

	// Validate user authorization
	restErr = s.isAuthorizedForPlaylist(ctx, userRole, userID, playlistID)
	if restErr != nil {
		return restErr
	}

	// Clear the playlist by removing all tracks
	if err := s.ClearPlaylistContents(ctx, playlistID); err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to clear playlist contents"}
	}

	return nil
}

// AddTracksToPlaylist handles the addition of tracks to an existing playlist.
// This function first reads the playlist from the DB into a tree structure,
// adds the new tracks from the request into the tree, and then updates the
// playlist in the database with the new structure.
//
// Parameters:
//   - ctx: context.Context
//     The context that carries deadlines, cancellation signals, and other request-scoped values.
//   - userRole: string
//     The role of the user performing the operation (e.g., admin, user).
//   - userID: string
//     The ID of the user performing the operation.
//   - playlistID: string
//     The ID of the playlist to update.
//   - request: *model.SetPlaylistTrackOrderRequest
//     The request containing track IDs and their positions.
//
// Return Values:
//   - *model.RestError: An error response if something goes wrong.
//     Returns nil if successful.
func (s *Service) AddTracksToPlaylist(ctx context.Context, userRole, userID, playlistID string, request *model.SetPlaylistTrackOrderRequest, rebalance bool) *model.RestError {
	// Step 1: Validate the user and playlist existence
	if restErr := s.validateUserForPlaylist(ctx, userRole, userID, playlistID); restErr != nil {
		return restErr
	}

	// Step 2: Generate paths and validate tracks or playlists
	stringPlaylistID, err := uuid.Parse(playlistID)
	if err != nil {
		s.logger.Printf("Invalid PlaylistID %s: %v\n", playlistID, err)
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Invalid parse PlaylistID string to UUID"}
	}

	var addPlaylistStructs []model.PlaylistStruct
	for i, referenceID := range request.ItemIDs {
		referenceType, restErr := s.validateTrackOrPlaylist(ctx, referenceID)
		if restErr != nil {
			return restErr
		}

		positionStr := generatePositionStr(request, i)

		addPlaylistStructs = append(addPlaylistStructs, model.PlaylistStruct{
			PlaylistID: stringPlaylistID,
			Path:       pgtype.Ltree{String: fmt.Sprintf("%s.%s.%s.%s", playlistID, referenceType, referenceID, positionStr)},
		})
	}

	// Step 3: Update the playlist tree with new tracks
	treeAddItems, restErr := s.updatePlaylistTree(ctx, playlistID, addPlaylistStructs, rebalance)
	if restErr != nil {
		return restErr
	}

	// Step 4: Update the playlist in the database
	if restErr = s.updateDatabaseWithTree(ctx, treeAddItems); restErr != nil {
		return restErr
	}

	return nil
}
