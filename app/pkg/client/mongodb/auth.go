package mongodb

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetStoredRefreshToken retrieves the refresh token for a user by email.
func (c *MongoClient) GetStoredRefreshToken(userEmail string) (string, error) {
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
func (c *MongoClient) SetStoredRefreshToken(userEmail, refreshToken string) error {
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
