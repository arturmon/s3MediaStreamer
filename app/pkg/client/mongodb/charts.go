package mongodb

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (c *MongoClient) GetTracksForLearn() ([]model.Track, error) {
	collection, err := c.FindCollections(config.CollectionTrack)

	filter := bson.M{}
	ctx := context.TODO()
	cursor, errFind := collection.Find(ctx, filter)

	if errFind != nil {
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
		if len(tracks) == ChunkSize {
			break
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

func (c *MongoClient) CreateTops(list []model.Tops) error {
	collection, err := c.FindCollections(config.CollectionTrack)
	insertableList := make([]interface{}, len(list))
	if err != nil {
		return err
	}
	_, err = collection.InsertMany(context.TODO(), insertableList)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) CleanupRecords(retentionPeriod time.Duration) error {
	return fmt.Errorf("retention Period is not supported for MongoDB, %s not finded", retentionPeriod)
}
