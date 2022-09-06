package web

import "time"

type DateTime time.Time

type ResponseRequest struct {
	Begin DateTime `json:"message,omitempty"`
}
