package playlist

import (
	"context"
	"fmt"
	"net/http"
	"s3MediaStreamer/app/model"

	"github.com/emirpasic/gods/maps/treemap"
)

// Helper: Validates user authorization for a playlist
func (s *Service) validateUserForPlaylist(ctx context.Context, userContext *model.UserContext, playlistID string) *model.RestError {
	// Validate user authorization
	restErr := s.isAuthorizedForPlaylist(ctx, userContext, playlistID)
	if restErr != nil {
		return restErr
	}

	restErr = s.ensurePlaylistExists(ctx, playlistID)
	if restErr != nil {
		return restErr
	}

	return nil
}

// Helper: Validates the track or playlist and determines its type
func (s *Service) validateTrackOrPlaylist(ctx context.Context, referenceID string) (string, *model.RestError) {
	_, errTrack := s.trackRepository.GetTracksByColumns(ctx, referenceID, "_id")
	isPlaylist, err := s.CheckPlaylistExists(ctx, referenceID)

	if err != nil && errTrack != nil {
		return "", &model.RestError{
			Code: http.StatusNotFound,
			Err:  fmt.Sprintf("item with ID %s not found", referenceID),
		}
	}

	if errTrack == nil {
		return "track", nil
	}
	if isPlaylist {
		return "playlist", nil
	}

	return "", &model.RestError{Code: http.StatusNotFound, Err: fmt.Sprintf("Reference (track or playlist) with ID %s not found", referenceID)}
}

// Helper: Generates the position string based on the request and index
func generatePositionStr(request *model.SetPlaylistTrackOrderRequest, i int) string {
	if len(request.ItemIDs) == 1 {
		if request.Position != nil {
			return fmt.Sprintf("%d", *request.Position)
		}
		return "0"
	}
	return fmt.Sprintf("%d", i)
}

// Helper: Updates the playlist tree structure
func (s *Service) updatePlaylistTree(ctx context.Context, playlistID string, addPlaylistStructs []model.PlaylistStruct, rebalance bool) (*treemap.Map, *model.RestError) {
	// Retrieve existing playlist items
	structPlaylist, err := s.trackRepository.GetPlaylistItems(ctx, playlistID)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to retrieve playlist tracks"}
	}

	// Initialize the tree and fill with existing tracks
	treeAddItems := treemap.NewWithStringComparator()
	err = s.tree.FillTree(treeAddItems, structPlaylist)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to load playlist into tree"}
	}

	// Add new tracks to the tree
	err = s.tree.AddToTree(treeAddItems, addPlaylistStructs, rebalance)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to add tracks to playlist"}
	}

	return treeAddItems, nil
}

// Helper: Updates the database with the modified tree
func (s *Service) updateDatabaseWithTree(ctx context.Context, tree *treemap.Map) *model.RestError {
	err := s.trackRepository.InsertPositionInDB(ctx, tree)
	if err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "Failed to update playlist in database"}
	}
	return nil
}
