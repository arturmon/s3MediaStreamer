package amqp

import (
	"context"
	"fmt"
	"net"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/streadway/amqp"
)

// MessageClient represents an AMQP message client.
type MessageClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	cfg     *config.Config
	logger  *logging.Logger
	storage *model.DBConfig
}

// NewAMQPClient creates a new instance of the MessageClient.
func NewAMQPClient(queueName string, cfg *config.Config, logger *logging.Logger) (*MessageClient, error) {
	amqpURL := net.JoinHostPort(cfg.MessageQueue.User, cfg.MessageQueue.Pass)

	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = net.JoinHostPort(amqpURL, fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort))
	} else {
		amqpURL = net.JoinHostPort(amqpURL, cfg.MessageQueue.Broker)
	}
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = net.JoinHostPort(amqpURL, fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort))
	} else {
		amqpURL = net.JoinHostPort(amqpURL, cfg.MessageQueue.Broker)
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := channel.QueueDeclare(
		queueName,       // queue name
		QueueDurable,    // durable
		QueueAutoDelete, // delete when unused
		QueueExclusive,  // exclusive
		QueueNoWait,     // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, err
	}

	storage, err := model.NewDBConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &MessageClient{
		conn:    conn,
		channel: channel,
		queue:   queue,
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}, nil
}

// Consume starts consuming messages from the queue.
func (c *MessageClient) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	messages, err := c.channel.Consume(
		c.queue.Name,      // queue
		"",                // consumer
		SubscribeAutoAck,  // auto-ack
		SubscribeExlusive, // exclusive
		SubscribeNoLocal,  // no-local
		SubscribeNoWait,   // no-wait
		nil,               // args
	)
	if err != nil {
		return nil, err
	}

	go c.consumeMessages(ctx, messages) // Start processing the messages in a separate goroutine

	return messages, nil
}

// Close closes the AMQP channel.
func (c *MessageClient) Close() error {
	return c.channel.Close()
}
