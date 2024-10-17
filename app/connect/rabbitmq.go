package connect

import (
	"context"
	"fmt"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"

	"github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQConnection(_ context.Context, cfg *model.Config, logger *logs.Logger) (*amqp091.Connection, error) {
	var amqpURL string
	logger.Info("Starting AMQP Connection...")
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort)
	} else {
		amqpURL = cfg.MessageQueue.Broker
	}
	protocol := "amqp" //amqp, rabbitmq
	amqpURLpriv := fmt.Sprintf("%s://%s:%s@%s", protocol, cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)
	logger.Debugf("AMQP URL: %s", amqpURLpriv)
	conn, err := amqp091.Dial(amqpURLpriv)
	if err != nil {
		logger.Errorf("(AMQP) Failed to connect rabbitmq at %s://%s:***@%s, errors: %v", protocol, cfg.MessageQueue.User, amqpURL, err)

		return nil, err
	}

	logger.Infof("(AMQP) Successfully connected to AMQP Client: %s://%s:***@%s", protocol, cfg.MessageQueue.User, amqpURL)

	return conn, nil
}
