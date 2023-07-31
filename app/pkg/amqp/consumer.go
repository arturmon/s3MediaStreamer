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
			return

		case message, ok := <-messages:
			if !ok {
				return
			}

			var data map[string]interface{}
			err := json.Unmarshal(message.Body, &data)
			if err != nil {
				c.logger.Printf("Error decoding message: %v", err)
				continue
			}

			action, ok := data["action"].(string)
			if !ok {
				c.logger.Println("Invalid action field")
				continue
			}

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

			default:
				c.logger.Printf("Unknown action: %s", action)
				continue
			}
		}
	}
}
