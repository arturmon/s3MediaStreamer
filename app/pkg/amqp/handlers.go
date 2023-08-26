package amqp

import (
	"encoding/json"
	"fmt"
)

// handlePostAlbums handles the "PostAlbums" action by processing the incoming album data.
func (c *MessageClient) handlePostAlbums(data map[string]interface{}) error {
	albumsData, ok := data["albums"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid albums data")
	}

	albumsJSON, err := json.Marshal(albumsData)
	if err != nil {
		return err
	}

	return c.amqpPostAlbums(string(albumsJSON))
}

// handleGetAllAlbums handles the "GetAllAlbums" action by fetching and logging all albums.
func (c *MessageClient) handleGetAllAlbums() error {
	albums, err := c.amqpGetAllAlbums()
	if err != nil {
		return err
	}

	albumsJSON, err := json.Marshal(albums)
	if err != nil {
		return err
	}

	c.logger.Printf("Albums: %s", albumsJSON)
	return nil
}

// handleGetDeleteAll handles the "GetDeleteAll" action by deleting all albums.
func (c *MessageClient) handleGetDeleteAll() error {
	err := c.amqpGetDeleteAll()
	if err != nil {
		return err
	}

	return nil
}

// handleGetAlbumByCode handles the "GetAlbumByCode" action by fetching and logging an album by its code.
func (c *MessageClient) handleGetAlbumByCode(data map[string]interface{}) error {
	albumCode, ok := data["albumCode"].(string)
	if !ok {
		return fmt.Errorf("invalid albumCode")
	}

	album, err := c.amqpGetAlbumByCode(albumCode)
	if err != nil {
		return err
	}

	c.logger.Printf("Album: %+v", album)
	return nil
}

// handleAddUser handles the "AddUser" action by adding a new user.
func (c *MessageClient) handleAddUser(data map[string]interface{}) error {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		return fmt.Errorf("invalid userEmail")
	}
	name, ok := data["name"].(string)
	if !ok {
		return fmt.Errorf("invalid name")
	}
	password, ok := data["password"].(string)
	if !ok {
		return fmt.Errorf("invalid password")
	}
	role, ok := data["role"].(string)
	if !ok {
		return fmt.Errorf("invalid role")
	}

	return c.amqpAddUser(userEmail, name, password, role)
}

// handleDeleteUser handles the "DeleteUser" action by deleting a user.
func (c *MessageClient) handleDeleteUser(data map[string]interface{}) error {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		return fmt.Errorf("invalid userEmail")
	}

	return c.amqpDeleteUser(userEmail)
}

// handleFindUserToEmail handles the "FindUserToEmail" action by finding a user by their email.
func (c *MessageClient) handleFindUserToEmail(data map[string]interface{}) error {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		return fmt.Errorf("invalid userEmail")
	}

	return c.amqpFindUserToEmail(userEmail)
}

// HandlerUpdateAlbum handles the "UpdateAlbum" action by updating an album's data.
func (c *MessageClient) handleUpdateAlbum(data map[string]interface{}) error {
	newAlbumsData, ok := data["album"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid albums data")
	}

	albumsJSON, err := json.Marshal(newAlbumsData)
	if err != nil {
		return err
	}

	return c.amqpUpdateAlbum(string(albumsJSON))
}
