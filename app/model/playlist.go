package model

import (
	"time"

	"github.com/google/uuid"
)

type PLayList struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	Level       int64     `json:"level" bson:"level" example:"34"`
	Title       string    `json:"title" bson:"title" example:"Title name"`
	Description string    `json:"description" bson:"description" example:"A short description of the application"`
	CreatorUser uuid.UUID `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
}
