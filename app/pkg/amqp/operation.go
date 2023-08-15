package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"skeleton-golange-application/app/internal/config"
	"time"
)

// amqpGetAlbumByCode retrieves an album by its code using AMQP.
func (c *MessageClient) amqpGetAlbumByCode(code string) (*config.Album, error) {
	album, err := c.storage.Operations.GetIssuesByCode(code)
	if err != nil {
		publishErr := c.publishMessage(TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return nil, err
	}

	err = c.publishMessage(TypePublisherMessage, album)
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

	albumsList := make([]config.Album, 0, len(albumArray)) // Pre-allocate with the expected length
	for _, albumObj := range albumArray {
		albumData, ok := albumObj.(map[string]interface{})
		if !ok {
			c.logger.Println("Invalid album data")
			continue
		}

		album := config.Album{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       albumData["Title"].(string),
			Artist:      albumData["Artist"].(string),
			Price:       albumData["Price"].(float64),
			Code:        albumData["Code"].(string),
			Description: albumData["Description"].(string),
			Completed:   albumData["Completed"].(bool),
		}

		albumsList = append(albumsList, album)
	}

	err = c.storage.Operations.CreateMany(albumsList)
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

// amqpGetAllAlbums retrieves all albums using AMQP.
func (c *MessageClient) amqpGetAllAlbums() ([]config.Album, error) {
	albums, err := c.storage.Operations.GetAllIssues()
	if err != nil {
		if err != nil {
			publishErr := c.publishMessage(TypePublisherError, err.Error())
			if publishErr != nil {
				c.logger.Printf("Error publishing error message: %v", publishErr)
			}
			return nil, err
		}
	}

	err = c.publishMessage(TypePublisherMessage, albums)
	if err != nil {
		c.logger.Printf("Error publishing message: %v", err)
	}

	return albums, nil
}

// amqpGetDeleteAll deletes all albums using AMQP.
func (c *MessageClient) amqpGetDeleteAll() error {
	err := c.storage.Operations.DeleteAll()
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
func (c *MessageClient) amqpAddUser(userEmail, name, password string) error {
	user := config.User{
		Id:       uuid.New(),
		Name:     name,
		Email:    userEmail,
		Password: []byte(password),
	}

	// Check if user already exists
	_, err := c.storage.Operations.FindUserToEmail(userEmail)
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
	hashedPassword, err := bcrypt.GenerateFromPassword(user.Password, 14)
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
	info, err := c.storage.Operations.FindUserToEmail(userEmail)
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
	// Check if the required fields are present in the data
	var data map[string]interface{}
	err := json.Unmarshal([]byte(albumsData), &data)
	if err != nil {
		return err
	}

	// Fetch the album from the database based on the provided code
	code, ok := data["Code"].(string)
	if !ok {
		c.logger.Println("Invalid code field")
		return fmt.Errorf("invalid code field")
	}

	existingAlbum, err := c.storage.Operations.GetIssuesByCode(code)
	if err != nil {
		c.logger.Printf("Error fetching album with code %s: %v", code, err)
		return err
	}

	// Update the album fields based on the data received
	if title, ok := data["Title"].(string); ok {
		existingAlbum.Title = title
	}
	if artist, ok := data["Artist"].(string); ok {
		existingAlbum.Artist = artist
	}
	if price, ok := data["Price"].(float64); ok {
		existingAlbum.Price = price
	}
	if description, ok := data["Description"].(string); ok {
		existingAlbum.Description = description
	}
	if completed, ok := data["Completed"].(bool); ok {
		existingAlbum.Completed = completed
	}

	// Update the album in the database
	err = c.storage.Operations.UpdateIssue(&existingAlbum)
	if err != nil {
		c.logger.Printf("Error updating album with code %s: %v", code, err)
		return err
	}

	c.logger.Printf("Album with code %s updated successfully", code)
	return nil
}
