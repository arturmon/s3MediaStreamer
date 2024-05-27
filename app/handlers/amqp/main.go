package amqp

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/streadway/amqp"
)

// AmqpServiceInterface определяет методы для работы с RabbitMQ.
type AmqpServiceInterface interface {
	HandleMessage(ctx context.Context, messageBody map[string]interface{})
}

// AmqpHandler представляет репозиторий для работы с RabbitMQ.
type AmqpHandler struct {
	amqpService rabbitmq.MessageService
	Conn        *amqp.Connection
	channel     *amqp.Channel
	queue       *amqp.Queue
}

// NewAmqpHandler создает новый экземпляр RabbitMQRepository.
func NewAmqpHandler(amqpService rabbitmq.MessageService, conn *amqp.Connection, channel *amqp.Channel, queue *amqp.Queue) *AmqpHandler {
	return &AmqpHandler{
		amqpService: amqpService,
		Conn:        conn,
		channel:     channel,
		queue:       queue,
	}
}

func (c *AmqpHandler) Ping(_ context.Context) bool {
	if c.Conn.IsClosed() {
		return false
	}
	return true
}

// TODO
func (c *AmqpHandler) StartAMQPConsumers(ctx context.Context, logger *logs.Logger, messageClient *AmqpHandler) {
	numWorkers := 5
	workerDone := make(chan struct{})
	go func() {
		if err := c.ConsumeMessagesWithPool(ctx, logger, messageClient, numWorkers, workerDone); err != nil {
			// Handle error
			logger.Fatal(err)
		}
	}()
}

func (c *AmqpHandler) HandleMessage(ctx context.Context, messageBody map[string]interface{}) {
	c.amqpService.HandleMessage(ctx, messageBody)
}
