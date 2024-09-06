package connect

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/streadway/amqp"
)

func NewRabbitMQConnection(_ context.Context, cfg *model.Config, logger *logs.Logger) (*amqp.Connection, error) {
	var amqpURL string
	logger.Info("Starting AMQP Connection...")
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort)
	} else {
		amqpURL = cfg.MessageQueue.Broker
	}

	// amqpURLpriv := fmt.Sprintf("rabbitmq://%s:%s@%s", cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)
	amqpURLpriv := fmt.Sprintf("amqp://%s:%s@%s", cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)
	logger.Debugf("AMQP URL: %s", amqpURLpriv)
	conn, err := amqp.Dial(amqpURLpriv)
	if err != nil {
		logger.Errorf("(AMQP) Failed to connect rabbitmq at rabbitmq://%s:***@%s, errors: %v", cfg.MessageQueue.User, amqpURL, err)

		return nil, err
	}

	logger.Infof("(AMQP) Successfully connected to AMQP Client: rabbitmq://%s:***@%s", cfg.MessageQueue.User, amqpURL)

	return conn, nil
}
