package model

type Node struct {
	ID       string
	ParentID string
	Type     string // 'track' or 'playlist'
	Position int
	Data     interface{}
}
