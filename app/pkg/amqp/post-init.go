package amqp

import (
	"errors"
	"github.com/streadway/amqp"
	"skeleton-golange-application/app/internal/config"
)

// Getter method for retrieving the channel.
func (c *MessageClient) GetChannel() *amqp.Channel {
	return c.channel
}

// PostInit performs post-initialization setup for AMQP configuration.
func PostInit(amqpClient *MessageClient, cfg *config.Config) error {
	if amqpClient == nil {
		return errors.New("AMQP client is nil")
	}

	channel := amqpClient.channel

	// Declare the publisher queue
	err := DeclareQueue(channel, cfg.MessageQueue.PubQueueName)
	if err != nil {
		return err
	}

	// Declare the publisher exchange
	err = DeclareExchange(channel, cfg.MessageQueue.PubExchange, ExchangeType)
	if err != nil {
		return err
	}

	// Bind the publisher queue to the publisher exchange
	err = BindQueue(channel, cfg.MessageQueue.PubQueueName, cfg.MessageQueue.PubExchange, cfg.MessageQueue.SubRoutingKey)
	if err != nil {
		return err
	}

	return nil
}

// DeclareQueue declares an AMQP queue.
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

// DeclareExchange declares an AMQP exchange.
func DeclareExchange(ch *amqp.Channel, exchangeName, exchangeType string) error {
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

// BindQueue binds an AMQP queue to an exchange with a routing key.
func BindQueue(ch *amqp.Channel, queueName, exchangeName, routingKey string) error {
	return ch.QueueBind(
		queueName,       // queue name
		routingKey,      // routing key
		exchangeName,    // exchange name
		QueueBindNoWait, // no-wait
		nil,             // arguments
	)
}
