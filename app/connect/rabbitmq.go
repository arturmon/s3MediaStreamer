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
	protocol := "amqp" //amqp, rabbitmq
	if cfg.MessageQueue.BrokerPort != 0 {
		amqpURL = fmt.Sprintf("%s:%d", cfg.MessageQueue.Broker, cfg.MessageQueue.BrokerPort)
	} else {
		amqpURL = cfg.MessageQueue.Broker
	}

	logFields := []model.LogField{
		{Key: "TypeConnect", Value: "Rabbitmq", Mask: ""},
		{Key: "User", Value: cfg.MessageQueue.User, Mask: ""},
		{Key: "Protocol", Value: protocol, Mask: ""},
		{Key: "Addr", Value: amqpURL, Mask: ""},
		{Key: "Password", Value: cfg.MessageQueue.Pass, Mask: "password"},
	}
	loggerMsg := logs.NewLoggerMessageConnect(logFields)

	logger.Info("Starting AMQP Connection...")

	amqpURLpriv := fmt.Sprintf("%s://%s:%s@%s", protocol, cfg.MessageQueue.User, cfg.MessageQueue.Pass, amqpURL)
	logger.Debugf("AMQP URL: %s", amqpURLpriv)
	conn, err := amqp091.Dial(amqpURLpriv)
	if err != nil {
		logger.Slog().Error("(AMQP) Failed to connect", "connection", loggerMsg.MaskFields())
		return nil, err
	}

	logger.Slog().Error("(AMQP) Successfully to connect", "connection", loggerMsg.MaskFields())
	return conn, nil
}
