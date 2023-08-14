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

			// Decode the incoming JSON message body.
			var data map[string]interface{}
			err := json.Unmarshal(message.Body, &data)
			if err != nil {
				c.logger.Printf("Error decoding message: %v", err)
				continue // Continue processing other messages.
			}

			// Extract the action field from the message data.
			action, ok := data["action"].(string)
			if !ok {
				c.logger.Println("Invalid action field")
				continue // Continue processing other messages.
			}

			// Based on the action, handle different types of messages.
			switch action {
			case "PostAlbums":
				c.handlePostAlbums(data)

			case "GetAllAlbums":
				c.handleGetAllAlbums()

			case "GetDeleteAll":
				c.handleGetDeleteAll()

			case "GetAlbumByCode":
				c.handleGetAlbumByCode(data)

			case "AddUser":
				c.handleAddUser(data)

			case "DeleteUser":
				c.handleDeleteUser(data)

			case "FindUserToEmail":
				c.handleFindUserToEmail(data)

			case "UpdateAlbum":
				c.HandlerUpdateAlbum(data)

			default:
				c.logger.Printf("Unknown action: %s", action)
				continue // Continue processing other messages with unknown actions.
			}
		}
	}
}
