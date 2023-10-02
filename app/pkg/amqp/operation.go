package amqp

import (
	"encoding/json"
	"fmt"
	"skeleton-golange-application/model"
	"time"

	"github.com/bojanz/currency"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 14

// amqpGetAlbumByCode retrieves an album by its code using AMQP.
func (c *MessageClient) amqpGetAlbumByCode(code string) (*model.Album, error) {
	album, getErr := c.storage.Operations.GetAlbumsByCode(code)
	if getErr != nil {
		publishErr := c.publishMessage(TypePublisherError, getErr.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return nil, getErr
	}

	err := c.publishMessage(TypePublisherMessage, album)
	if err != nil {
		c.logger.Printf("Error publishing message: %v", err)
	}

	return &album, nil
}

// amqpPostAlbums posts albums data using AMQP.
func (c *MessageClient) amqpPostAlbums(albumsData string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(albumsData), &data)
	if err != nil {
		return err
	}

	albumArray, ok := data["album"].([]interface{})
	if !ok {
		c.logger.Println("Invalid albums data")
		return fmt.Errorf("invalid albums data")
	}
	var newPrice currency.Amount
	albumsList := make([]model.Album, 0, len(albumArray)) // Pre-allocate with the expected length
	for _, albumObj := range albumArray {
		albumData, castOk := albumObj.(map[string]interface{})
		if !castOk {
			c.logger.Println("Invalid album data")
			continue
		}
		priceData, castOk := albumData["Price"].(map[string]interface{})
		if !castOk {
			c.logger.Println("Invalid price data")
			continue
		}

		numberStr, numberOk := priceData["Number"].(string)
		currencyCode, currencyOk := priceData["Currency"].(string)

		if !numberOk || !currencyOk {
			c.logger.Println("Invalid price components")
			continue
		}

		newPrice, err = currency.NewAmount(numberStr, currencyCode)
		if err != nil {
			c.logger.Println("Error creating currency.Amount")
			continue
		}

		systemUser, errSystemUser := c.storage.Operations.FindUser(c.cfg.MessageQueue.SystemWriteUser, "email")
		if errSystemUser != nil {
			c.logger.Println("Error find system user")
			return errSystemUser
		}

		parsedUUID, errUUID := uuid.Parse(systemUser.ID.String())
		if errUUID != nil {
			c.logger.Println("Error parsing system user uuid")
			return errUUID
		}

		album := model.Album{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       albumData["Title"].(string),
			Artist:      albumData["Artist"].(string),
			Price:       newPrice,
			Code:        albumData["Code"].(string),
			Description: albumData["Description"].(string),
			Sender:      systemUser.Name,
			CreatorUser: parsedUUID,
		}

		albumsList = append(albumsList, album)
	}

	err = c.storage.Operations.CreateAlbums(albumsList)
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "Albums have been successfully posted",
	}

	err = c.publishMessage(TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish PostAlbumsSuccess message:", err)
	}

	return nil
}

// amqpGetAllAlbums retrieves paginated albums using AMQP.
func (c *MessageClient) amqpGetAllAlbums(page, pageSize int, sortBy, sortOrder, filter string) ([]model.Album, int, error) {
	albums, totalRows, err := c.storage.Operations.GetAlbums(page, pageSize, sortBy, sortOrder, filter)
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return nil, 0, err
	}

	err = c.publishMessage(TypePublisherMessage, albums)
	if err != nil {
		c.logger.Printf("Error publishing message: %v", err)
	}

	return albums, totalRows, nil
}

// amqpGetDeleteAll deletes all albums using AMQP.
func (c *MessageClient) amqpGetDeleteAll() error {
	err := c.storage.Operations.DeleteAlbumsAll()
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "Delete all albums request",
	}

	err = c.publishMessage(TypePublisherMessage, messageData)
	if err != nil {
		return err
	}

	c.logger.Println("All albums have been successfully deleted.")

	return nil
}

// amqpAddUser adds a user using AMQP.
func (c *MessageClient) amqpAddUser(userEmail, name, password, role string) error {
	user := model.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    userEmail,
		Password: []byte(password),
		Role:     role,
	}

	// Check if user already exists
	_, err := c.storage.Operations.FindUser(userEmail, "email")
	if err == nil {
		// User with this email already exists
		errMsg := fmt.Errorf("user %s with this email already exists", userEmail)
		publishErr := c.publishMessage(TypePublisherError, errMsg.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return errMsg
	}

	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword(user.Password, bcryptCost)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	err = c.storage.Operations.CreateUser(user)
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "User has been successfully added",
	}

	err = c.publishMessage(TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish AddUserSuccess message:", err)
	}

	return nil
}

// amqpDeleteUser deletes a user using AMQP.
func (c *MessageClient) amqpDeleteUser(userEmail string) error {
	err := c.storage.Operations.DeleteUser(userEmail)
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}
	// TODO нужно еще подумать
	// {"info":"User has been successfully deleted","userEmail":"a@a.com"}
	messageData := map[string]interface{}{
		"info":      "User has been successfully deleted",
		"userEmail": userEmail,
	}

	err = c.publishMessage(TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish DeleteUserSuccess message:", err)
	}

	return nil
}

// amqpFindUserToEmail finds a user by email using AMQP.
func (c *MessageClient) amqpFindUserToEmail(userEmail string) error {
	info, err := c.storage.Operations.FindUser(userEmail, "email")
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	err = c.publishMessage(TypePublisherMessage, info)
	if err != nil {
		c.logger.Println("Failed to publish AddUserSuccess message:", err)
	}

	return nil
}

// amqpUpdateAlbum updates an album using AMQP.
func (c *MessageClient) amqpUpdateAlbum(albumsData string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(albumsData), &data)
	if err != nil {
		return err
	}

	// Fetch the album code from the data
	code, codeOk := data["Code"].(string)
	if !codeOk {
		return fmt.Errorf("invalid code field")
	}

	// Fetch the album from the database based on the provided code
	existingAlbum, getErr := c.storage.Operations.GetAlbumsByCode(code)
	if getErr != nil {
		c.logger.Printf("Error fetching album with code %s: %v", code, getErr)
		return getErr
	}

	// Update the album fields based on the data received
	if title, titleOk := data["Title"].(string); titleOk {
		existingAlbum.Title = title
	}
	if artist, artistOk := data["Artist"].(string); artistOk {
		existingAlbum.Artist = artist
	}
	var newPrice currency.Amount
	// Handle price data
	if priceData, priceOk := data["Price"].(map[string]interface{}); priceOk {
		numberStr, numberOk := priceData["Number"].(string)
		currencyCode, currencyOk := priceData["Currency"].(string)

		if !numberOk || !currencyOk {
			return fmt.Errorf("invalid price components")
		}

		newPrice, err = currency.NewAmount(numberStr, currencyCode)
		if err != nil {
			return err
		}

		existingAlbum.Price = newPrice
	}

	if description, descOk := data["Description"].(string); descOk {
		existingAlbum.Description = description
	}
	if sender, senderOk := data["Sender"].(string); senderOk {
		existingAlbum.Sender = sender
	}

	// Update the album in the database
	err = c.storage.Operations.UpdateAlbums(&existingAlbum)
	if err != nil {
		c.logger.Printf("Error updating album with code %s: %v", code, err)
		return err
	}

	c.logger.Printf("Album with code %s updated successfully", code)
	return nil
}
