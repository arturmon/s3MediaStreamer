package model

import (
	"time"

	"github.com/google/uuid"
)

// Track represents data about a record track.
type Track struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" pg:"default:now()" swaggerignore:"true"`
	Album       string    `json:"album" bson:"album" example:"Album name"`
	AlbumArtist string    `json:"album_artist" bson:"album_artist" example:"Album artist name"`
	Composer    string    `json:"composer" bson:"composer" example:"Composer name"`
	Genre       string    `json:"genre" bson:"genre" example:"Genre name"`
	Lyrics      string    `json:"lyrics" bson:"lyrics" example:"Lyrics of the track"`
	Title       string    `json:"title" bson:"title" example:"Title name"`
	Artist      string    `json:"artist" bson:"artist" example:"Artist name"`
	Year        int       `json:"year" bson:"year" example:"2022"`
	Comment     string    `json:"comment" bson:"comment" example:"Additional comments"`
	Disc        int       `json:"disc" bson:"disc" example:"1"`
	DiscTotal   int       `json:"disc_total" bson:"disc_total" example:"2"`
	Track       int       `json:"track" bson:"track" example:"3"`
	TrackTotal  int       `json:"track_total" bson:"track_total" example:"10"`
	Sender      string    `json:"sender" bson:"sender" example:"sender set"`
	CreatorUser uuid.UUID `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
	Likes       bool      `json:"likes" bson:"likes"`
	S3Version   string    `json:"s3Version" bson:"s3Version"`
}

type Tops struct {
	ID          uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" pg:"default:now()" swaggerignore:"true"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" pg:"default:now()" swaggerignore:"true"`
	Title       string    `json:"title" bson:"title" example:"Title name"`
	Artist      string    `json:"artist" bson:"artist" example:"Artist name"`
	Description string    `json:"description" bson:"description" example:"A short description of the application"`
	Sender      string    `json:"sender" bson:"sender" example:"open_ai"`
	CreatorUser uuid.UUID `json:"_creator_user" bson:"_creator_user" pg:"type:uuid" swaggerignore:"true"`
}
