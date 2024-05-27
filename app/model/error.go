package model

type RestError struct {
	Code int    `json:"code"`
	Err  string `json:"error"`
}
