package mongodb

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *MongoClient) CreateTracks(ctx context.Context, list []model.Track) error {
	_, span := otel.Tracer("").Start(ctx, "CreateTracks")
	defer span.End()
	insertableList := make([]interface{}, len(list))
	for i := range list {
		v := &list[i] // Use a pointer to the current issue.
		insertableList[i] = v
		/*
			if v.Completed {
				log.Infof("INFO: Completed %d: %f    %s\n", i+1, v.Price, v.Title)
			} else {
				log.Infof("INFO: No Completed %d: %f    %s\n", i+1, v.Price, v.Title)
			}

		*/
	}

	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}
	_, err = collection.InsertMany(context.TODO(), insertableList)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) GetTracks(ctx context.Context, offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Track, int, error) {
	_, span := otel.Tracer("").Start(ctx, "GetTracks")
	defer span.End()
	if offset < 1 || limit < 1 {
		return nil, 0, errors.New("invalid pagination parameters")
	}

	filter := bson.D{}
	if filterArtist != "" {
		filter = append(filter, bson.E{Key: "artist", Value: bson.M{"$regex": filterArtist, "$options": "i"}})
	}

	var tracks []model.Track

	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return tracks, 0, err
	}

	findOptions := options.Find()
	findOptions.SetSkip(int64((offset - 1) * limit))
	findOptions.SetLimit(int64(limit))

	sortOptions := options.Find()
	sortOrderValue := 1
	if sortOrder == "desc" {
		sortOrderValue = -1
	}
	sortOptions.SetSort(bson.D{{Key: sortBy, Value: sortOrderValue}})
	findOptions.Sort = sortOptions.Sort

	cur, findError := collection.Find(context.TODO(), filter, findOptions)
	if findError != nil {
		return tracks, 0, findError
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var track model.Track
		decodeErr := cur.Decode(&track)
		if decodeErr != nil {
			return tracks, 0, decodeErr
		}
		tracks = append(tracks, track)
	}

	totalCount, countErr := collection.CountDocuments(context.TODO(), filter)
	if countErr != nil {
		return tracks, 0, countErr
	}

	if len(tracks) == 0 {
		return tracks, 0, mongo.ErrNoDocuments
	}
	return tracks, int(totalCount), nil
}

func (c *MongoClient) GetTracksByColumns(ctx context.Context, code, columns string) (*model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetTracksByColumns")
	defer span.End()
	result := model.Track{}
	filter := bson.D{{Key: columns, Value: code}} // Fix the linting issue here
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return &result, err
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return &result, err
	}
	return &result, nil
}

func (c *MongoClient) DeleteTracks(ctx context.Context, code, columns string) error {
	_, span := otel.Tracer("").Start(ctx, "DeleteTracks")
	defer span.End()
	filter := bson.D{primitive.E{Key: columns, Value: code}}
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) DeleteTracksAll(ctx context.Context) error {
	_, span := otel.Tracer("").Start(ctx, "DeleteTracksAll")
	defer span.End()
	selector := bson.D{{}}
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}
	_, err = collection.DeleteMany(context.TODO(), selector)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) UpdateTracks(ctx context.Context, track *model.Track) error {
	_, span := otel.Tracer("").Start(ctx, "UpdateTracks")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}

	// Define the filter to find the track by its code.
	filter := bson.D{{Key: "_id", Value: track.ID}}

	// Define the update fields using the $set operator to update only the specified fields.
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "title", Value: track.Title},
			{Key: "artist", Value: track.Artist},
			{Key: "album", Value: track.Album},              // Adding album field
			{Key: "album_artist", Value: track.AlbumArtist}, // Adding album_artist field
			{Key: "composer", Value: track.Composer},        // Adding composer field
			{Key: "genre", Value: track.Genre},              // Adding genre field
			{Key: "lyrics", Value: track.Lyrics},            // Adding lyrics field
			{Key: "year", Value: track.Year},                // Adding year field
			{Key: "comment", Value: track.Comment},          // Adding comment field
			{Key: "disc", Value: track.Disc},                // Adding disc field
			{Key: "track", Value: track.Track},              // Adding track field
			{Key: "sender", Value: track.Sender},
		}},
		{Key: "$currentDate", Value: bson.D{
			{Key: "updatedat", Value: true}, // Set the "updatedat" field to the current date.
		}},
	}

	// Perform the update operation.
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoClient) GetAllTracks(ctx context.Context) ([]model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetAllTracks")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	filter := bson.M{} // Add your filter criteria if needed

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tracks []model.Track

	for cursor.Next(ctx) {
		var track model.Track
		if err = cursor.Decode(&track); err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return tracks, nil
}

func (c *MongoClient) AddTrackToPlaylist(ctx context.Context, playlistID, trackID string) error {
	_, span := otel.Tracer("").Start(ctx, "AddTrackToPlaylist")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID and trackID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	trackUUID, err := uuid.Parse(trackID)
	if err != nil {
		return err
	}

	// Update the playlist document to include the new track
	filterPlaylist := bson.M{"_id": playlistUUID}
	update := bson.M{"$push": bson.M{"tracks": trackUUID}}
	_, err = collectionPlaylist.UpdateOne(ctx, filterPlaylist, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoClient) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error {
	_, span := otel.Tracer("").Start(ctx, "RemoveTrackFromPlaylist")
	defer span.End()
	collectionPlaylist, err := c.FindCollections(config.CollectionPlaylist)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID and trackID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return err
	}

	trackUUID, err := uuid.Parse(trackID)
	if err != nil {
		return err
	}

	// Update the playlist document to remove the specified track
	filterPlaylist := bson.M{"_id": playlistUUID}
	update := bson.M{"$pull": bson.M{"tracks": trackUUID}}
	_, err = collectionPlaylist.UpdateOne(ctx, filterPlaylist, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoClient) GetAllTracksByPositions(ctx context.Context, playlistID string) ([]model.Track, error) {
	_, span := otel.Tracer("").Start(ctx, "GetAllTracksByPositions")
	defer span.End()
	collectionTrack, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	// Convert playlistID to UUID
	playlistUUID, err := uuid.Parse(playlistID)
	if err != nil {
		return nil, err
	}

	// Find tracks related to the playlist and sort them by position
	filterTracks := bson.M{"_creator_user": playlistUUID}
	options := options.Find().SetSort(primitive.D{{Key: "position", Value: 1}}) // Assuming "position" is the field indicating the order
	cursor, err := collectionTrack.Find(ctx, filterTracks, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var playlistTracks []model.Track

	for cursor.Next(ctx) {
		var track model.Track
		if err = cursor.Decode(&track); err != nil {
			return nil, err
		}
		playlistTracks = append(playlistTracks, track)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return playlistTracks, nil
}
