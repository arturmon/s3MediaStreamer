package mongodb

import (
	"context"
	"errors"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"

	"go.opentelemetry.io/otel"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *MongoClient) FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error) {
	_, span := otel.Tracer("").Start(ctx, "FindUser")
	defer span.End()
	result := model.User{}
	// Define filter query for fetching a specific document from the collection.
	filter := bson.D{{Key: columnType, Value: value}}
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return result, err
	}
	// Perform FindOne operation and validate against errors.
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}
	// Return the result without any error.
	return result, nil
}

func (c *MongoClient) CreateUser(ctx context.Context, user model.User) error {
	_, span := otel.Tracer("").Start(ctx, "CreateUser")
	defer span.End()
	collection, err := c.FindCollections(config.CollectionUser)
	if err != nil {
		return err
	}

	// Поиск пользователя с помощью email.
	existingUser, err := c.FindUser(ctx, user.Email, "email")
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		// Создание нового пользователя
		_, err = collection.InsertOne(context.TODO(), user)
		if err != nil {
			return err
		}

		return nil
	}

	// Пользователь с таким email уже существует.
	return fmt.Errorf("user with email '%s' already exists", existingUser.Email)
}

func (c *MongoClient) DeleteUser(ctx context.Context, email string) error {
	_, span := otel.Tracer("").Start(ctx, "DeleteUser")
	defer span.End()
	return fmt.Errorf("DeleteUser is not supported for MongoDB, '%s' not deleted", email)
}

func (c *MongoClient) UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error {
	_, span := otel.Tracer("").Start(ctx, "UpdateUser")
	defer span.End()
	// Define the filter to find the user by email.
	filter := bson.M{"email": email}

	// Define the update statement based on the provided fields.
	update := bson.M{}
	for key, value := range fields {
		update[key] = value
	}

	// Create an options instance to enable upsert (insert if document not found).
	options := options.Update().SetUpsert(false)

	// Perform the update operation.
	collection, err := c.FindCollections(config.CollectionTrack)
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(context.TODO(), filter, bson.M{"$set": update}, options)
	if err != nil {
		return err
	}

	return nil
}
