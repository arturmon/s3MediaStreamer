package amqp

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

// Interface defines methods for working with RabbitMQ.
type Interface interface {
	HandleMessage(ctx context.Context, messageBody map[string]interface{})
}

// Handler represents a repository for working with RabbitMQ.
type Handler struct {
	amqpService rabbitmq.Service
	Conn        *amqp091.Connection
	channel     *amqp091.Channel
	queue       *amqp091.Queue
	logger      *logs.Logger
	autoAck     bool
}

// NewAMQPHandler creates a new RabbitMQRepository instance.
func NewAMQPHandler(amqpService rabbitmq.Service, conn *amqp091.Connection, channel *amqp091.Channel, queue *amqp091.Queue, cfg *model.Config, logger *logs.Logger) *Handler {
	return &Handler{
		amqpService: amqpService,
		Conn:        conn,
		channel:     channel,
		queue:       queue,
		logger:      logger,
		autoAck:     cfg.MessageQueue.SubscribeAutoAck,
	}
}

func (c *Handler) Ping(_ context.Context) bool {
	return !c.Conn.IsClosed()
}

func (c *Handler) StartAMQPConsumers(ctx context.Context) {
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err := c.ConsumeMessagesWithPool(ctx, c.logger, c, numWorkers, workerDone); err != nil {
			// Handle error
			c.logger.Fatal(err.Error())
		}
	}()
}

func (c *Handler) HandleMessage(ctx context.Context, messageBody map[string]interface{}) {
	c.amqpService.HandleMessage(ctx, messageBody)
}
