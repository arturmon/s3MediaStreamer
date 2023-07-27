package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"skeleton-golange-application/app/internal/config"
	"time"
)

func (c *AMQPClient) amqpGetAlbumByCode(Code string) (*config.Album, error) {
	album, err := c.storage.Operations.GetIssuesByCode(Code)
	if err != nil {
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return nil, err
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, album)
	if err != nil {
		c.logger.Printf("Error publishing message: %v", err)
	}

	return &album, nil
}

func (c *AMQPClient) amqpPostAlbums(albumsData string) error {
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

	var albumsList []config.Album
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
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "Albums have been successfully posted",
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish PostAlbumsSuccess message:", err)
	}

	return nil
}

func (c *AMQPClient) amqpGetAllAlbums() ([]config.Album, error) {
	albums, err := c.storage.Operations.GetAllIssues()
	if err != nil {
		if err != nil {
			publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
			if publishErr != nil {
				c.logger.Printf("Error publishing error message: %v", publishErr)
			}
			return nil, err
		}
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, albums)
	if err != nil {
		c.logger.Printf("Error publishing message: %v", err)
	}

	return albums, nil
}

func (c *AMQPClient) amqpGetDeleteAll() error {
	err := c.storage.Operations.DeleteAll()
	if err != nil {
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "Delete all albums request",
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, messageData)
	if err != nil {
		return err
	}

	c.logger.Println("All albums have been successfully deleted.")

	return nil
}

func (c *AMQPClient) amqpAddUser(userEmail, name, password string) error {
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
		publishErr := c.publishMessage(context.Background(), TypePublisherError, errMsg.Error())
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
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	messageData := map[string]interface{}{
		"info": "User has been successfully added",
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish AddUserSuccess message:", err)
	}

	return nil
}

func (c *AMQPClient) amqpDeleteUser(userEmail string) error {
	err := c.storage.Operations.DeleteUser(userEmail)
	if err != nil {
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
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

	err = c.publishMessage(context.Background(), TypePublisherMessage, messageData)
	if err != nil {
		c.logger.Println("Failed to publish DeleteUserSuccess message:", err)
	}

	return nil
}

func (c *AMQPClient) amqpFindUserToEmail(userEmail string) error {
	info, err := c.storage.Operations.FindUserToEmail(userEmail)
	if err != nil {
		publishErr := c.publishMessage(context.Background(), TypePublisherError, err.Error())
		if publishErr != nil {
			c.logger.Printf("Error publishing error message: %v", publishErr)
		}
		return err
	}

	err = c.publishMessage(context.Background(), TypePublisherMessage, info)
	if err != nil {
		c.logger.Println("Failed to publish AddUserSuccess message:", err)
	}

	return nil
}
