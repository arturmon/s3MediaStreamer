package amqp

import (
	"context"
	"fmt"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"

	"github.com/panjf2000/ants/v2"

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
	var amqpURL string
	logger.Info("Starting AMQP Client...")
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort)
	} else {
		amqpURL = cfg.MessageQueue.Broker
	}

	amqpURLpriv := fmt.Sprintf("amqp://%s:%s@%s", cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)

	conn, err := amqp.Dial(amqpURLpriv)
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

	storage, err := model.NewDBConfig(cfg, logger)
	if err != nil {
		return nil, err
	}

	logger.Infof("Connect AMQP Client: amqp://%s:***@%s", cfg.MessageQueue.User, amqpURL)

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
	c.logger.Infof("AMQP Comsume: %s", c.queue.Name)
	return messages, nil
}

// Close closes the AMQP channel.
func (c *MessageClient) Close() error {
	return c.channel.Close()
}

func ConsumeMessages(ctx context.Context, logger logging.Logger, messageClient *MessageClient) {
	if messageClient == nil {
		return
	}

	poolSize := 10 // Set the desired pool size

	pool, poolErr := ants.NewPool(poolSize)
	if poolErr != nil {
		logger.Fatal("Failed to create ants pool:", poolErr)
	}
	defer pool.Release()

	consumeErr := pool.Submit(func() {
		messages, consumeErr := messageClient.Consume(ctx)
		if consumeErr != nil {
			logger.Fatal("Failed to start consuming messages:", consumeErr)
		}
		// Wait for the background goroutine to finish
		logger.Info("Waiting for the background goroutine to finish...")
		<-messages
		logger.Info("Finished consuming messages")
	})
	if consumeErr != nil {
		logger.Fatal("Failed to submit task to ants pool:", consumeErr)
	}
}
