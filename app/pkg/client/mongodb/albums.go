package mongodb

import (
	"context"
	"errors"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *MongoClient) CreateAlbums(list []model.Album) error {
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

	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.InsertMany(context.TODO(), insertableList)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) GetAlbums(offset, limit int, sortBy, sortOrder, filterArtist string) ([]model.Album, int, error) {
	if offset < 1 || limit < 1 {
		return nil, 0, errors.New("invalid pagination parameters")
	}

	filter := bson.D{}
	if filterArtist != "" {
		filter = append(filter, bson.E{Key: "artist", Value: bson.M{"$regex": filterArtist, "$options": "i"}})
	}

	var albums []model.Album

	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return albums, 0, err
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
		return albums, 0, findError
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var album model.Album
		decodeErr := cur.Decode(&album)
		if decodeErr != nil {
			return albums, 0, decodeErr
		}
		albums = append(albums, album)
	}

	totalCount, countErr := collection.CountDocuments(context.TODO(), filter)
	if countErr != nil {
		return albums, 0, countErr
	}

	if len(albums) == 0 {
		return albums, 0, mongo.ErrNoDocuments
	}
	return albums, int(totalCount), nil
}

func (c *MongoClient) GetAlbumsByCode(code string) (model.Album, error) {
	result := model.Album{}
	filter := bson.D{{Key: "code", Value: code}} // Fix the linting issue here
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return result, err
	}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (c *MongoClient) DeleteAlbums(code string) error {
	filter := bson.D{primitive.E{Key: "code", Value: code}}
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) DeleteAlbumsAll() error {
	selector := bson.D{{}}
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}
	_, err = collection.DeleteMany(context.TODO(), selector)
	if err != nil {
		return err
	}
	return nil
}

func (c *MongoClient) UpdateAlbums(album *model.Album) error {
	collection, err := c.FindCollections(config.CollectionAlbum)
	if err != nil {
		return err
	}

	// Define the filter to find the album by its code.
	filter := bson.D{{Key: "code", Value: album.Code}}

	// Define the update fields using the $set operator to update only the specified fields.
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "title", Value: album.Title},
			{Key: "artist", Value: album.Artist},
			{Key: "price", Value: album.Price},
			{Key: "description", Value: album.Description},
			{Key: "sender", Value: album.Sender},
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
