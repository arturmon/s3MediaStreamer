package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (c *MongoClient) CleanSessions() error {
	filter := bson.D{
		{Key: "expires_on", Value: bson.D{{Key: "$lt", Value: time.Now()}}},
	}

	collection, err := c.FindCollections("http_sessions")
	if err != nil {
		return err
	}

	_, err = collection.DeleteMany(context.TODO(), filter)
	return err
}
