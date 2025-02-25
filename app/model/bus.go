package model

// QueueConfig holds the configuration for a RabbitMQ queue.
type QueueConfig struct {
	Name       string                 // Name of the queue
	Durable    bool                   // Whether the queue is durable
	AutoDelete bool                   // Whether the queue is auto-deleted when no consumers are left
	Exclusive  bool                   // Whether the queue is exclusive to the connection
	NoWait     bool                   // Whether the queue declaration will wait for acknowledgment from the server
	Arguments  map[string]interface{} // Additional arguments for the queue declaration (optional)
}
