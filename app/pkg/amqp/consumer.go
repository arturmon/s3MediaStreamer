package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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
			c.handleMessage(message)
		}
	}
}

func (c *MessageClient) handleMessage(message amqp.Delivery) {
	// Decode the incoming JSON message body.
	var data map[string]interface{}
	err := json.Unmarshal(message.Body, &data)
	if err != nil {
		c.logger.Printf("Error decoding message: %v", err)
		return
	}

	// Extract the action field from the message data.
	action, ok := data["action"].(string)
	if !ok {
		c.logger.Println("Invalid action field")
		return
	}

	// Based on the action, handle different types of messages.
	switch action {
	case "PostAlbums":
		c.handleActionPostAlbums(data)

	case "GetAllAlbums":
		c.handleActionGetAllAlbums(data)

	case "GetDeleteAll":
		c.handleActionGetDeleteAll()

	case "GetAlbumByCode":
		c.handleActionGetAlbumByCode(data)

	case "AddUser":
		c.handleActionAddUser(data)

	case "DeleteUser":
		c.handleActionDeleteUser(data)

	case "FindUserToEmail":
		c.handleActionFindUserToEmail(data)

	case "UpdateAlbum":
		c.handleActionUpdateAlbum(data)

	default:
		c.logger.Printf("Unknown action: %s", action)
		return
	}
}

func (c *MessageClient) handleActionPostAlbums(data map[string]interface{}) {
	resultErr := c.handlePostAlbums(data)
	c.handleResult(resultErr, "PostAlbums")
}

func (c *MessageClient) handleActionGetAllAlbums(data map[string]interface{}) {
	pageRaw, pageExists := data["page"]
	pageSizeRaw, pageSizeExists := data["page_size"]
	sortByRaw, sortByExists := data["sort_by"]
	sortOrderRaw, sortOrderExists := data["sort_order"]
	filterRaw, filterExists := data["filter"]

	if !pageExists || !pageSizeExists || !sortByExists || !sortOrderExists || !filterExists {
		c.logger.Println("Missing required parameters in message data")
		return
	}

	page, pageErr := strconv.Atoi(fmt.Sprint(pageRaw))
	pageSize, pageSizeErr := strconv.Atoi(fmt.Sprint(pageSizeRaw))
	sortBy := fmt.Sprint(sortByRaw)
	sortOrder := fmt.Sprint(sortOrderRaw)
	filter := fmt.Sprint(filterRaw)

	if pageErr != nil || pageSizeErr != nil {
		c.logger.Println("Invalid parameter values in message data")
		return
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	err := c.handleGetAllAlbums(offset, limit, sortBy, sortOrder, filter)
	if err != nil {
		c.logger.Printf("Error handling GetAllAlbums action: %v", err)
	}
}

func (c *MessageClient) handleActionGetDeleteAll() {
	resultErr := c.handleGetDeleteAll()
	c.handleResult(resultErr, "GetDeleteAll")
}

func (c *MessageClient) handleActionGetAlbumByCode(data map[string]interface{}) {
	resultErr := c.handleGetAlbumByCode(data)
	c.handleResult(resultErr, "GetAlbumByCode")
}

func (c *MessageClient) handleActionAddUser(data map[string]interface{}) {
	resultErr := c.handleAddUser(data)
	c.handleResult(resultErr, "AddUser")
}

func (c *MessageClient) handleActionDeleteUser(data map[string]interface{}) {
	resultErr := c.handleDeleteUser(data)
	c.handleResult(resultErr, "DeleteUser")
}

func (c *MessageClient) handleActionFindUserToEmail(data map[string]interface{}) {
	resultErr := c.handleFindUserToEmail(data)
	c.handleResult(resultErr, "FindUserToEmail")
}

func (c *MessageClient) handleActionUpdateAlbum(data map[string]interface{}) {
	resultErr := c.handleUpdateAlbum(data)
	c.handleResult(resultErr, "UpdateAlbum")
}

func (c *MessageClient) handleResult(resultErr error, action string) {
	if resultErr != nil {
		errorData := map[string]interface{}{
			"error": resultErr.Error(),
		}
		c.publishAndLogResult(TypePublisherError, errorData)
	} else {
		successData := map[string]interface{}{
			"info": fmt.Sprintf("Successfully handled %s", action),
		}
		c.publishAndLogResult(TypePublisherStatus, successData)
	}
}

func (c *MessageClient) publishAndLogResult(resultType string, data map[string]interface{}) {
	publishErr := c.publishMessage(resultType, data)
	if publishErr != nil {
		c.logger.Printf("Error publishing %s message: %v", resultType, publishErr)
	}
}
