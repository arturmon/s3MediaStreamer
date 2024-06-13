package amqp

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/streadway/amqp"
)

// AmqpServiceInterface определяет методы для работы с RabbitMQ.
type Interface interface {
	HandleMessage(ctx context.Context, messageBody map[string]interface{})
}

// AmqpHandler представляет репозиторий для работы с RabbitMQ.
type Handler struct {
	amqpService rabbitmq.Service
	Conn        *amqp.Connection
	channel     *amqp.Channel
	queue       *amqp.Queue
}

// NewAmqpHandler создает новый экземпляр RabbitMQRepository.
func NewAMQPHandler(amqpService rabbitmq.Service, conn *amqp.Connection, channel *amqp.Channel, queue *amqp.Queue) *Handler {
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
			logger.Fatal(err)
		}
	}()
}

func (c *Handler) HandleMessage(ctx context.Context, messageBody map[string]interface{}) {
	c.amqpService.HandleMessage(ctx, messageBody)
}
