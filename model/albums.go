package model

import (
	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"time"
)

// Album represents data about a record album.
type Album struct {
	ID          uuid.UUID       `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	UpdatedAt   time.Time       `json:"updated_at" bson:"updated_at" pg:"default:now()" swaggerignore:"true"`
	Title       string          `json:"title" bson:"title" example:"Title name"`
	Artist      string          `json:"artist" bson:"artist" example:"Artist name"`
	Price       currency.Amount `json:"price" bson:"price" example:"{Number: 1.10, Currency: EUR}" swaggertype:"string,string"`
	Code        string          `json:"code" bson:"code" example:"I001"`
	Description string          `json:"description" bson:"description" example:"A short description of the application"`
	Sender      string          `json:"sender" bson:"sender" example:"amqp or rest"`
	CreatorUser uuid.UUID       `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
}
