package mongodb

import (
	"context"
	"errors"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/model"
	"strconv"

	"go.opentelemetry.io/otel"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func (c *MongoClient) CreatePlayListName(ctx context.Context, newPlaylist model.PLayList) error {
	_, span := otel.Tracer("").Start(ctx, "CreatePlayListName")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	_, err = collection.InsertOne(ctx, newPlaylist)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoClient) GetPlayListByID(ctx context.Context, playlistID string) (model.PLayList, []model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetPlayListByID")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return model.PLayList{}, nil, err
	}

	collectionTrack, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return model.PLayList{}, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return model.PLayList{}, nil, err
	}

	// Find playlist by ID
	filterPlaylist := bson.M{"_id": playlistUUID}
	var playlist model.PLayList
	err = collectionPlaylist.FindOne(ctx, filterPlaylist).Decode(&playlist)
	if err != nil {
		return model.PLayList{}, nil, err
	}

	// Find tracks related to the playlist
	filterTracks := bson.M{"_creator_user": playlistUUID}
	cursor, err := collectionTrack.Find(ctx, filterTracks)
	if err != nil {
		return model.PLayList{}, nil, err
	}
	defer cursor.Close(ctx)

	var playlistTracks []model.Track

	for cursor.Next(ctx) {
		var track model.Track
		if err = cursor.Decode(&track); err != nil {
			return model.PLayList{}, nil, err
		}
		playlistTracks = append(playlistTracks, track)
	}

	if err = cursor.Err(); err != nil {
		return model.PLayList{}, nil, err
	}

	return playlist, playlistTracks, nil
}

func (c *MongoClient) DeletePlaylist(ctx context.Context, playlistID string) error {
	_, span := otel.Tracer("").Start(ctx, "DeletePlaylist")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	// Find and delete the playlist by ID
	filter := bson.M{"_id": playlistUUID}
	_, err = collectionPlaylist.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	// Optionally, you may want to delete related tracks associated with the playlist
	collectionTrack, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}

	filterTracks := bson.M{"_creator_user": playlistUUID}
	_, err = collectionTrack.DeleteMany(ctx, filterTracks)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoClient) PlaylistExists(ctx context.Context, title string) bool {
	_, span := otel.Tracer("").Start(ctx, "PlaylistExists")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Create filter to find playlist by title
	filter := bson.M{"title": title}

	// Check if any documents match the filter
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false
	}

	return count > 0
}

func (c *MongoClient) ClearPlayList(ctx context.Context, playlistTitle string) error {
	_, span := otel.Tracer("").Start(ctx, "ClearPlayList")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return err
	}

	collectionTrack, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Find the playlist by title
	filterPlaylist := bson.M{"title": playlistTitle}
	var playlist model.PLayList
	err = collectionPlaylist.FindOne(ctx, filterPlaylist).Decode(&playlist)
	if err != nil {
		return err
	}

	// Delete tracks associated with the playlist
	filterTracks := bson.M{"_creator_user": playlist.ID}
	_, err = collectionTrack.DeleteMany(ctx, filterTracks)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePlaylistTrackOrder updates the order of tracks within a playlist based on the provided order in MongoDB.
func (c *MongoClient) UpdatePlaylistTrackOrder(ctx context.Context, playlistID string, trackOrderRequest []string) error {
	_, span := otel.Tracer("").Start(ctx, "UpdatePlaylistTrackOrder")
	defer span.End()
	collectionTrack, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	// Fetch the current tracks associated with the playlist
	filter := bson.M{"_creator_user": playlistUUID}
	cursor, err := collectionTrack.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var currentTracks []model.Track
	for cursor.Next(ctx) {
		var track model.Track
		if err = cursor.Decode(&track); err != nil {
			return err
		}
		currentTracks = append(currentTracks, track)
	}

	if err = cursor.Err(); err != nil {
		return err
	}

	// Verify if the provided trackOrderRequest is valid
	if len(trackOrderRequest) != len(currentTracks) {
		return errors.New("invalid track order request")
	}

	// Create a map to quickly look up track positions
	positionMap := make(map[string]int)
	for i, track := range currentTracks {
		positionMap[track.ID.String()] = i
	}

	// Update the position of each track in the playlist
	for _, trackID := range trackOrderRequest {
		position, exists := positionMap[trackID]
		if !exists {
			return errors.New("invalid track ID in track order request")
		}

		// Update the track's position in the playlist
		// You need to adjust this part based on your actual MongoDB update logic
		update := bson.M{"$set": bson.M{"path": strconv.Itoa(position)}}
		_, err = collectionTrack.UpdateOne(ctx, bson.M{"_id": trackID}, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *MongoClient) GetAllPlayList(ctx context.Context, creatorUserID string) ([]model.PLayList, error) {
	_, span := otel.Tracer("").Start(ctx, "GetAllPlayList")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	var playlists []model.PLayList
	filter := bson.M{"_creator_user": creatorUserID}
	cursor, err := collectionPlaylist.Find(ctx, filter)

	for cursor.Next(ctx) {
		var playlist model.PLayList
		err = cursor.Decode(&playlist)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}
