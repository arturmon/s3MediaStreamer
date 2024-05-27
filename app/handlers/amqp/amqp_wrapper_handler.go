package amqp

import (
	"context"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/streadway/amqp"
)

const (
	QueueDurable    = true
	QueueAutoDelete = false
	QueueExclusive  = false
	QueueNoWait     = false
)

func NewRabbitMQHandlerWrapper(ctx context.Context, cfg *model.Config, logger *logs.Logger, conn *amqp.Connection, amqpService rabbitmq.MessageService) (*AmqpHandler, error) {
	logger.Info("Starting rabbitmq handler...")
	rabbitChannel, err := newRabbitMQChanel(conn)
	if err != nil {
		return nil, err
	}
	rabbitQueue, err := newRabbitMQQueue(
		cfg.MessageQueue.SubQueueName,
		conn)
	if err != nil {
		return nil, err
	}
	repo := NewAmqpHandler(amqpService, conn, rabbitChannel, rabbitQueue)
	repo.StartAMQPConsumers(ctx, logger, repo)
	return repo, nil
}

func newRabbitMQChanel(conn *amqp.Connection) (*amqp.Channel, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func newRabbitMQQueue(queueName string, conn *amqp.Connection) (*amqp.Queue, error) {
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
	return &queue, nil
}
