package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

func (c *AMQPClient) publishMessage(ctx context.Context, types string, data interface{}) error {
	// Ensure the channel is open before publishing
	if c.channel == nil {
		return fmt.Errorf("AMQP channel is not open")
	}
	// Convert the data to JSON
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if types == TypePublisherError {
		if errorMsg, ok := data.(string); ok {
			data = map[string]interface{}{
				"error": errorMsg,
			}
			body, err = json.Marshal(data)
			if err != nil {
				return err
			}
		}
	}
	// Prepare the AMQP message
	msg := amqp.Publishing{
		Type:        types,
		ContentType: "application/json",
		Body:        body,
	}

	// Publish the message to the AMQP server
	err = c.channel.Publish(
		c.cfg.MessageQueue.PubExchange,   // exchange
		c.cfg.MessageQueue.SubRoutingKey, // routing key
		false,                            // mandatory
		false,                            // immediate
		msg,
	)
	if err != nil {
		return err
	}

	c.logger.Printf("Published message with action %s: %s", types, body)
	return nil
}
