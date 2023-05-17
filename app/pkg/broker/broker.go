package broker

import (
	"github.com/streadway/amqp"
)

type MessageBrokerOperations interface {
	MessageBroker
}

type MessageBroker interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Subscribe(exchange, key, kind string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type RabbitMQBroker struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQBroker(uri string) (MessageBroker, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQBroker{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQBroker) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return r.channel.Publish(exchange, key, mandatory, immediate, msg)
}

func (r *RabbitMQBroker) Subscribe(exchange, key, kind string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	err := r.channel.ExchangeDeclare(exchange, kind, true, false, false, noWait, nil)
	if err != nil {
		return nil, err
	}

	queue, err := r.channel.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		return nil, err
	}

	err = r.channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, err
	}

	return r.channel.Consume(queue.Name, "", autoAck, exclusive, noLocal, noWait, args)
}

func (r *RabbitMQBroker) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}

	if err := r.conn.Close(); err != nil {
		return err
	}

	return nil
}
