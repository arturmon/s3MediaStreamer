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

// Function to publish an "AddUser" action with the provided user details
/*
func (c *AMQPClient) PublishAddUser(ctx context.Context, userEmail, name, password string) error {
	data := map[string]interface{}{
		"action":    "AddUser",
		"userEmail": userEmail,
		"name":      name,
		"password":  password,
	}

	// Use the publishMessage function to send the "AddUser" action
	return c.publishMessage(ctx, "AddUser", data)
}


*/
// TODO
// Publish an "AddUser" action
//err = amqpClient.PublishAddUser(context.Background(), "a@a.com", "a", "1")
//if err != nil {
//logger.Error("Failed to publish AddUser action:", err)
//}
