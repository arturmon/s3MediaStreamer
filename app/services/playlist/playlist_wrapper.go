package playlist

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/model"
)

const adminPolicy = "admin"

func (s *Service) isAuthorizedForPlaylist(ctx context.Context, userRole, userID, playlistID string) *model.RestError {
	// Get the owner ID of the playlist from the repository
	ownerUUID, err := s.playlistRepository.GetPlaylistOwner(ctx, playlistID)
	if err != nil {
		s.logger.Errorf("Failed to get playlist owner for playlistID %s: %v", playlistID, err)
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to get playlist owner"}
	}

	// Find the user who owns the playlist
	user, err := s.user.FindUser(ctx, ownerUUID.String(), "_id")
	if err != nil {
		s.logger.Errorf("Failed to find user with ownerUUID %s: %v", ownerUUID.String(), err)
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to find user"}
	}
	if userRole != user.Role {
		s.logger.Warnf("Unauthorized access: user role mismatch for userID %s, expected role %s, but got %s", userID, user.Role, userRole)
		return &model.RestError{Code: http.StatusForbidden, Err: "Unauthorized access"}
	}

	// Check if the user is authorized to access or modify the playlist
	if user.Role != adminPolicy && user.ID.String() != userID {
		s.logger.Warnf("Unauthorized access: userID %s is not the owner of playlistID %s", userID, playlistID)
		return &model.RestError{Code: http.StatusForbidden, Err: "Unauthorized access"}
	}

	return nil
}

func (s *Service) ensurePlaylistExists(ctx context.Context, playlistID string) *model.RestError {
	exists, err := s.playlistRepository.CheckPlaylistExists(ctx, playlistID)
	if err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to check if playlist exists"}
	}
	if !exists {
		return &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	var playlist model.PLayList
	err = s.playlistRepository.FetchPlaylistInfo(ctx, playlistID, &playlist)
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "Failed to retrieve playlist"}
	}
	return nil
}
