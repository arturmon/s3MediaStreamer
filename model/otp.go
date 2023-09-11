package model

type OTPInput struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}
