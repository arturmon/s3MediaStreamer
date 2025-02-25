package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

// PublisherImpl represents a RabbitMQ message publisher.
type PublisherImpl struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

// NewPublisher initializes a new RabbitMQ publisher.
func NewPublisher(conn *amqp091.Connection) (*PublisherImpl, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ channel: %w", err)
	}

	return &PublisherImpl{
		conn:    conn,
		channel: channel,
	}, nil
}

// Publish sends a message to the specified topic (queue).
func (p *PublisherImpl) Publish(ctx context.Context, topic string, message []byte) error {
	err := p.channel.PublishWithContext(
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
func (p *PublisherImpl) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ channel: %w", err)
	}
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	return nil
}
