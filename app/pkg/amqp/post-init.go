package amqp

import (
	"github.com/streadway/amqp"
	"skeleton-golange-application/app/internal/config"
)

// Getter method for the channel
func (c *AMQPClient) GetChannel() *amqp.Channel {
	return c.channel
}

func PreInit(amqpClient *AMQPClient, cfg *config.Config) error {
	channel := amqpClient.channel
	err := DeclareQueue(channel, cfg.MessageQueue.PubQueueName)
	if err != nil {
		return err
	}

	err = DeclareExchange(channel, cfg.MessageQueue.PubExchange, ExchangeType)
	if err != nil {
		return err
	}

	err = BindQueue(channel, cfg.MessageQueue.PubQueueName, cfg.MessageQueue.PubExchange, cfg.MessageQueue.SubRoutingKey)
	if err != nil {
		return err
	}

	return nil
}

func DeclareQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName,       // queue name
		QueueDurable,    // durable
		QueueAutoDelete, // auto-delete
		QueueExclusive,  // exclusive
		QueueNoWait,     // no-wait
		nil,             // arguments
	)
	return err
}

func DeclareExchange(ch *amqp.Channel, exchangeName string, exchangeType string) error {
	return ch.ExchangeDeclare(
		exchangeName,       // exchange name
		exchangeType,       // exchange type (e.g., "direct", "topic", "fanout")
		ExchangeDurable,    // durable
		ExchangeAutoDelete, // auto-delete
		ExchangeInternal,   // internal
		ExchangeNoWait,     // no-wait
		nil,                // arguments
	)
}

func BindQueue(ch *amqp.Channel, queueName, exchangeName, routingKey string) error {
	return ch.QueueBind(
		queueName,       // queue name
		routingKey,      // routing key
		exchangeName,    // exchange name
		QueueBindNoWait, // no-wait
		nil,             // arguments
	)
}
