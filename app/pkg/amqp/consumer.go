package amqp

import (
	"context"
	"encoding/json"

	"github.com/streadway/amqp"
)

func (c *MessageClient) consumeMessages(ctx context.Context, messages <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			// If the context is canceled, return and stop processing messages.
			return

		case message, ok := <-messages:
			// Check if the channel is closed (no more messages).
			if !ok {
				return
			}

			// Handle the message based on its action
			go c.handleMessage(ctx, message)
		}
	}
}

func (c *MessageClient) handleMessage(ctx context.Context, message amqp.Delivery) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Decode the incoming JSON message body.
	var data map[string]interface{}
	err := json.Unmarshal(message.Body, &data)
	if err != nil {
		c.logger.Printf("Error decoding message: %v", err)
		return
	}

	// Extract the action field from the message data.
	action, ok := data["EventName"].(string)
	if !ok {
		c.logger.Println("Invalid action field")
		return
	}

	s3event, errExtract := extractRecordsEvent(data)
	c.logger.Debugf("%v\n", s3event)
	if errExtract != nil {
		c.logger.Printf("Error extract message: %v", errExtract)
		return
	}
	// Based on the action, handle different types of messages.
	switch action {
	case "s3:ObjectRemoved:Delete":
		err = c.deleteEvent(ctx, s3event)
		if err != nil {
			c.logger.Printf("Error handling deleteEvent: %v", err)
			return
		}
	case "s3:ObjectCreated:Put":
		err = c.putEvent(ctx, s3event)
		if err != nil {
			c.logger.Printf("Error handling putEvent: %v", err)
			return
		}
	default:
		c.logger.Debugf("Event: %s not processed", action)
	}
}

func extractRecordsEvent(data map[string]interface{}) (*MessageBody, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &MessageBody{}, err
	}

	// Unmarshal the JSON data into a Records struct
	var messageBody MessageBody
	err = json.Unmarshal(jsonData, &messageBody)
	if err != nil {
		return &MessageBody{}, err
	}

	// Check if Records array is not empty
	if messageBody.Records == nil || len(messageBody.Records) == 0 {
		return &MessageBody{}, err
	}

	return &messageBody, nil
}
