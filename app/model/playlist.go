package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type PLayList struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	Title       string    `json:"title" bson:"title" example:"Title name"`
	Description string    `json:"description" bson:"description" example:"A short description of the application"`
	CreatorUser uuid.UUID `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
}

type PlaylistsResponse struct {
	PLayLists []PLayList `json:"playlists"`
}

// Request Define a struct to parse the request body.
var Request struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// SetPlaylistTrackOrderRequest represents a request to set track order or add tracks to a playlist.
type SetPlaylistTrackOrderRequest struct {
	ItemIDs  []string `json:"item_ids" binding:"required"` // List of track IDs to add
	Position *int     `json:"position,omitempty"`          // Optional position where tracks will be added. If not provided, tracks will be added to the end.
}

type PlaylistStruct struct {
	PlaylistID uuid.UUID    `json:"playlist_id"` // UUID
	Path       pgtype.Ltree `json:"path"`        // type LTREE
}

// NestedPlaylist represents the structure of a nested playlist with its own tracks and further nested playlists.
type NestedPlaylist struct {
	ID           uuid.UUID        `json:"_id"`
	CreatedAt    time.Time        `json:"created_at"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	CreatorUser  uuid.UUID        `json:"_creator_user"`
	Tracks       []Track          `json:"tracks,omitempty"`        // Tracks directly under this playlist
	SubPlaylists []NestedPlaylist `json:"sub_playlists,omitempty"` // Further nested playlists
}

// PlaylistTracksResponse represents the complete response structure for a playlist, including its nested sub-playlists and tracks.
type PlaylistTracksResponse struct {
	Playlist     NestedPlaylist   `json:"playlist"`
	SubPlaylists []NestedPlaylist `json:"sub_playlists,omitempty"`
	Tracks       []Track          `json:"tracks,omitempty"`
}
