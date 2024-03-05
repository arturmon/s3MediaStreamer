package mongodb

import (
	"context"
	"errors"
	"fmt"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/model"

	"go.opentelemetry.io/otel"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetStoredRefreshToken retrieves the refresh token for a user by email.
func (c *MongoClient) GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "GetStoredRefreshToken")
	defer span.End()
	result := model.User{}
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return "", err
	}
	filter := bson.M{"email": userEmail}
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", fmt.Errorf("user with email '%s' not found", userEmail)
	}
	return result.RefreshToken, nil
}

// SetStoredRefreshToken sets or updates the refresh token for a user by email.
func (c *MongoClient) SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error {
	_, span := otel.Tracer("").Start(ctx, "SetStoredRefreshToken")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return err
	}
	filter := bson.M{"email": userEmail}
	update := bson.M{"$set": bson.M{"refreshtoken": refreshToken}}
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}
