package amqp

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
)

type AMQPClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
	cfg     *config.Config
	logger  *logging.Logger
	storage *model.DBConfig
}

func NewAMQPClient(queueName string, config *config.Config, logger *logging.Logger) (*AMQPClient, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",
		config.MessageQueue.User,
		config.MessageQueue.Pass,
		config.MessageQueue.Broker,
		config.MessageQueue.BrokerPort))
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

	storage, err := model.NewDBConfig(config)
	if err != nil {
		return nil, err
	}

	return &AMQPClient{
		conn:    conn,
		channel: channel,
		queue:   queue,
		cfg:     config,
		logger:  logger,
		storage: storage,
	}, nil
}

func (c *AMQPClient) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
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

func (c *AMQPClient) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}
