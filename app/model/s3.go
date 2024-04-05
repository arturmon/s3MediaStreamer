package model

import "github.com/google/uuid"

type S3 struct {
	ID        uuid.UUID `json:"track_id" bson:"track_id" pg:"type:uuid" swaggerignore:"true"`
	S3Version string    `json:"s3Version" bson:"s3Version"`
}
