package model

import "github.com/google/uuid"

// User represents user account information.
// @Description User account information
// @Description with: user _id, name, email, password
type User struct {
	ID           uuid.UUID `json:"_id" bson:"_id" pg:"type:uuid" swaggerignore:"true"`
	Name         string    `json:"-" bson:"name" example:"Artur"`
	Email        string    `json:"email" bson:"email" example:"aaaa@aaaa.com"`
	Password     []byte    `json:"password" bson:"password"  example:"1111" swaggertype:"string"`
	Role         string    `json:"role" bson:"role"  example:"-" swaggerignore:"true"`
	RefreshToken string    `json:"refreshtoken" bson:"refreshtoken" swaggerignore:"true"`
	Otp_enabled  bool      `json:"otp_enabled" bson:"otp_enabled"`
	Otp_verified bool      `json:"otp_verified" bson:"otp_verified"`

	Otp_secret   string `bson:"otp_secret"`
	Otp_auth_url string `bson:"otp_auth_url"`
}
