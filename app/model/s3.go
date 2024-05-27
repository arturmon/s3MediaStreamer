package model

import "github.com/google/uuid"

type S3 struct {
	ID        uuid.UUID `json:"track_id" bson:"track_id" pg:"type:uuid" swaggerignore:"true"`
	S3Version string    `json:"s3Version" bson:"s3Version"`
}

type UploadS3 struct {
	ObjectName  string `json:"object_name" example:"Title name"`
	FilePath    string `json:"file_path" example:"File path"`
	ContentType string `json:"content_type" example:"Content Type"`
}
