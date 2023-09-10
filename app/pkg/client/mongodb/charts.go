package mongodb

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (c *MongoClient) GetAlbumsForLearn() ([]model.Album, error) {
	collection, err := c.FindCollections(config.CollectionAlbum)

	filter := bson.M{}
	ctx := context.TODO()
	cursor, errFind := collection.Find(ctx, filter)

	if errFind != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var albums []model.Album

	for cursor.Next(ctx) {
		var album model.Album
		if err = cursor.Decode(&album); err != nil {
			return nil, err
		}
		albums = append(albums, album)
		if len(albums) == ChunkSize {
			break
		}
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}
	return albums, nil
}

func (c *MongoClient) CreateTops(list []model.Tops) error {
	collection, err := c.FindCollections(config.CollectionAlbum)
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
