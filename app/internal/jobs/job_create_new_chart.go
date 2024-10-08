package jobs

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

const timeFormat = "2006-01-02 15:04:05"

func (j *CreateNewMusicChartJob) Run() {
	ctx := context.Background()
	if !j.app.Service.ConsulElection.IsLeader() {
		j.app.Logger.Println("I'm not the leader.")
		return
	}

	j.app.Logger.Println("Start Job Create New Music chart...")

	page, pageSize := 0, 100
	sortBy, sortOrder := "updated_at", "DESC"

	startTime := time.Now().Add(-24 * time.Hour).Format(timeFormat)
	endTime := time.Now().Format(timeFormat)

	tracks, _, err := j.app.Service.Track.GetTracks(ctx, page, pageSize, sortBy, sortOrder, "", startTime, endTime)

	if err != nil {
		j.app.Logger.Println("Error fetching tracks:", err)
		return
	}
	if tracks == nil {
		j.app.Logger.Println("No new products appeared")
		return
	}
	playlistID := uuid.New()
	consulSaveKey := fmt.Sprintf("service/%s/state/jobs/CreateNewMusicChartJob/playlistID", j.app.AppName)

	j.app.Logger.Infof("Save playlistID consul:%s\n", playlistID.String())
	playlistIDold, err := saveConsulState(ctx, j.app.Logger, consulSaveKey, playlistID, j.app.Service.ConsulService.ConsulClient)
	if err != nil {
		j.app.Logger.Println("Error save or load consul:", err)
	}

	j.app.Logger.Infof("Delete old Playlist: %s\n", playlistIDold.String())
	err = j.app.Service.Playlist.DeletePlaylist(ctx, playlistIDold.String())
	if err != nil {
		j.app.Logger.Println("Error delete old Playlist:", err)
		return
	}
	j.app.Logger.Infof("Generate new Playlist: %s\n", playlistID.String())
	// Create new Playlist
	// Generate a unique ID for the new playlist_handler (you can use your own method)
	newPlaylist := model.PLayList{
		ID:          playlistID,
		CreatedAt:   time.Now(),
		Title:       "New Music Chart 24 hours",
		Description: "Automatically generated from the last 24 hours",
		CreatorUser: uuid.Must(uuid.Parse("cac22f72-1fa2-4a81-876d-39fcf1cc9159")),
	}

	err = j.app.Service.Playlist.CreatePlayListName(ctx, newPlaylist)
	if err != nil {
		j.app.Logger.Println("Error create new Playlist:", err)
		return
	}

	var request model.SetPlaylistTrackOrderRequest
	request = convertTracksToSetPlaylistTrackOrderRequest(tracks)
	if errRest := j.app.Service.Playlist.AddTracksToPlaylist(
		ctx,
		"admin",
		"cac22f72-1fa2-4a81-876d-39fcf1cc9159",
		playlistID.String(),
		&request,
		false,
	); errRest != nil {
		j.app.Logger.Println("Error add tracks to new Playlist:")
		return
	}

}

func convertTracksToSetPlaylistTrackOrderRequest(tracks []model.Track) model.SetPlaylistTrackOrderRequest {
	var itemIDs []string
	for _, track := range tracks {
		itemIDs = append(itemIDs, track.ID.String())
	}

	// Create a slice to hold the positions
	var positions = 0

	// Create the SetPlaylistTrackOrderRequest struct
	request := model.SetPlaylistTrackOrderRequest{
		ItemIDs:  itemIDs,
		Position: &positions, // Use the newly created positions slice
	}

	return request
}

func saveConsulState(_ context.Context, logger *logs.Logger, key string, value uuid.UUID, client *api.Client) (uuid.UUID, error) {
	valueStr := value.String()

	// Try to get the existing value for the key from Consul
	kv, _, err := client.KV().Get(key, nil)
	if err != nil {
		logger.Error("Error fetching key from Consul:", err)
		return uuid.Nil, err
	}

	// If the key exists, return the existing value and update it
	if kv != nil && len(kv.Value) > 0 {
		existingValue, err := uuid.Parse(string(kv.Value)) // Convert the stored value back to UUID
		if err != nil {
			logger.Error("Error parsing existing UUID from Consul:", err)
			return uuid.Nil, err
		}

		// Log the existing value
		logger.Infof("Key exists in Consul with value: %s, updating to new value: %s.\n", existingValue, valueStr)

		// Update the value in Consul
		kvPair := &api.KVPair{
			Key:   key,
			Value: []byte(valueStr),
		}
		_, err = client.KV().Put(kvPair, nil)
		if err != nil {
			logger.Error("Error updating key in Consul:", err)
			return uuid.Nil, err
		}

		// Return the new value after updating
		return existingValue, nil
	}

	// If the key does not exist, create it
	logger.Infof("Key %s does not exist in Consul, creating new key.\n", key)
	kvPair := &api.KVPair{
		Key:   key,
		Value: []byte(valueStr),
	}
	_, err = client.KV().Put(kvPair, nil)
	if err != nil {
		logger.Error("Error creating key in Consul:", err)
		return uuid.Nil, err
	}

	// Return the new value after creating the key
	return value, nil
}