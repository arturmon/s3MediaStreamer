package amqp

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

// AmqpServiceInterface defines methods for working with RabbitMQ.
type Interface interface {
	HandleMessage(ctx context.Context, messageBody map[string]interface{})
}

// AmqpHandler represents a repository for working with RabbitMQ.
type Handler struct {
	amqpService rabbitmq.Service
	Conn        *amqp091.Connection
	channel     *amqp091.Channel
	queue       *amqp091.Queue
}

// NewAmqpHandler creates a new RabbitMQRepository instance.
func NewAMQPHandler(amqpService rabbitmq.Service, conn *amqp091.Connection, channel *amqp091.Channel, queue *amqp091.Queue) *Handler {
	return &Handler{
		amqpService: amqpService,
		Conn:        conn,
		channel:     channel,
		queue:       queue,
	}
}

func (c *Handler) Ping(_ context.Context) bool {
	return !c.Conn.IsClosed()
}

// TODO
func (c *Handler) StartAMQPConsumers(ctx context.Context, logger *logs.Logger, messageClient *Handler) {
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err := c.ConsumeMessagesWithPool(ctx, logger, messageClient, numWorkers, workerDone); err != nil {
			// Handle error
			logger.Fatal(err.Error())
		}
	}()
}

func (c *Handler) HandleMessage(ctx context.Context, messageBody map[string]interface{}) {
	c.amqpService.HandleMessage(ctx, messageBody)
}
