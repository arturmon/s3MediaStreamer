package amqp

import (
	"context"
	"errors"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

const internalServerErrorCode = 500

func NewRabbitMQHandlerWrapper(ctx context.Context, cfg *model.Config, logger *logs.Logger, conn *amqp091.Connection, amqpService rabbitmq.Service) (*Handler, error) {
	logger.Info("Starting rabbitmq handler...")
	queueConfigs := newQueueConfigs(cfg)

	repo, err := NewAMQPHandler(amqpService, conn, queueConfigs, cfg, logger)
	if err != nil {
		return nil, err
	}
	repo.StartAMQPConsumers(ctx)
	return repo, nil
}

func newRabbitMQChanel(conn *amqp091.Connection) (*amqp091.Channel, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func newRabbitMQQueue(queueConf *model.QueueConfig, channel *amqp091.Channel) (*amqp091.Queue, *amqp091.Error) {
	queue, err := channel.QueueDeclare(
		queueConf.Name,       // queue name
		queueConf.Durable,    // durable
		queueConf.AutoDelete, // delete when unused
		queueConf.Exclusive,  // exclusive
		queueConf.NoWait,     // no-wait
		queueConf.Arguments,  // arguments
	)
	if err != nil {
		var amqpErr *amqp091.Error
		if errors.As(err, &amqpErr) {
			return nil, amqpErr
		}
		return nil, amqpErr // Return error instead of trying to cast
	}
	return &queue, nil
}

func checkQueue(queueConf *model.QueueConfig, channel *amqp091.Channel, logger *logs.Logger) *amqp091.Error {
	// Check if a queue with this name and "classic" parameters already exists
	_, err := channel.QueueDeclarePassive(
		queueConf.Name,       // queue name
		queueConf.Durable,    // durable
		queueConf.AutoDelete, // delete when unused
		queueConf.Exclusive,  // exclusive
		queueConf.NoWait,     // no-wait
		queueConf.Arguments,  // arguments
	)

	if err == nil {
		// The queue already exists with these parameters, we can continue
		logger.Infof("Queue '%s' already exists", queueConf.Name)
		return nil
	}

	// If the error is of type amqp091.Error, return it directly
	var amqpErr *amqp091.Error
	if errors.As(err, &amqpErr) {
		return amqpErr
	}

	// If the error is not amqp091.Error, create and return a new amqp091.Error with the original error
	return &amqp091.Error{
		Code:   internalServerErrorCode, // or an appropriate code for your context
		Reason: err.Error(),
	}
}

func newQueueConfigs(cfg *model.Config) []model.QueueConfig {
	return cfg.Bus.QueueConfig
}
