package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

type RabbitMQPublisher struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

// NewRabbitMQPublisher initializes a new RabbitMQPublisher.
func NewRabbitMQPublisher(conn *amqp091.Connection) (*RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}

	return &RabbitMQPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// Publish sends a message to the specified topic (queue).
func (r *RabbitMQPublisher) Publish(ctx context.Context, topic string, message []byte) error {
	err := r.channel.PublishWithContext(
		ctx,
		"",    // exchange
		topic, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}
	return nil
}

// Close gracefully closes the channel and connection.
func (r *RabbitMQPublisher) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return nil
}
