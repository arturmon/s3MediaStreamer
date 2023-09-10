package postgresql

import "time"

const (
	ChunkSize             = 1000
	MaxConnectionAttempts = 5
	// MaxAttempts is the maximum number of attempts to connect to the database.
	MaxAttempts = 10
	// MaxDelay is the maximum delay between connection attempts.
	MaxDelay = 5 * time.Second
)
