package amqp

import (
	"github.com/streadway/amqp"
	"skeleton-golange-application/app/pkg/broker"
)

type AmqpBroker struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func New(uri string) (broker.MessageBroker, error) {
	connection, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}

	return &AmqpBroker{connection: connection, channel: channel}, nil
}

func (a *AmqpBroker) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return a.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (a *AmqpBroker) Subscribe(exchange, key, kind string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	err := a.channel.ExchangeDeclare(exchange, kind, true, false, false, noWait, nil)
	if err != nil {
		return nil, err
	}

	queue, err := a.channel.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	err = a.channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, err
	}

	return a.channel.Consume(queue.Name, "", autoAck, exclusive, noLocal, noWait, args)
}

func (a *AmqpBroker) Close() error {
	if err := a.channel.Close(); err != nil {
		return err
	}

	if err := a.connection.Close(); err != nil {
		return err
	}

	return nil
}
