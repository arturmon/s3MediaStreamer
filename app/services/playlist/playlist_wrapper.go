package playlist

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/model"
)

const adminPolicy = "admin"

func (c *Service) isAuthorizedForPlaylist(ctx context.Context, userRole, userID, playlistID string) (*model.User, *model.RestError) {
	// Get the owner ID of the playlist from the repository
	ownerUUID, err := c.playlistRepository.GetPlaylistOwner(ctx, playlistID)
	if err != nil {
		c.logger.Errorf("Failed to get playlist owner for playlistID %s: %v", playlistID, err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to get playlist owner"}
	}

	// Find the user who owns the playlist
	user, err := c.user.FindUser(ctx, ownerUUID.String(), "_id")
	if err != nil {
		c.logger.Errorf("Failed to find user with ownerUUID %s: %v", ownerUUID.String(), err)
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to find user"}
	}
	if userRole != user.Role {
		c.logger.Warnf("Unauthorized access: user role mismatch for userID %s, expected role %s, but got %s", userID, user.Role, userRole)
		return nil, &model.RestError{Code: http.StatusForbidden, Err: "Unauthorized access"}
	}

	// Check if the user is authorized to access or modify the playlist
	if user.Role != adminPolicy && user.ID.String() != userID {
		c.logger.Warnf("Unauthorized access: userID %s is not the owner of playlistID %s", userID, playlistID)
		return nil, &model.RestError{Code: http.StatusForbidden, Err: "Unauthorized access"}
	}

	return &user, nil
}

func (c *Service) ensurePlaylistExists(ctx context.Context, playlistID string) (*model.PLayList, *model.RestError) {
	exists, err := c.playlistRepository.CheckPlaylistExists(ctx, playlistID)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to check if playlist exists"}
	}
	if !exists {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "Playlist not found"}
	}

	var playlist model.PLayList
	err = c.playlistRepository.FetchPlaylistInfo(ctx, playlistID, &playlist)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "Failed to retrieve playlist"}
	}
	return &playlist, nil
}
