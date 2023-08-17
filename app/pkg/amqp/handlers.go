package amqp

import (
	"encoding/json"
)

// handlePostAlbums handles the "PostAlbums" action by processing the incoming album data.
func (c *MessageClient) handlePostAlbums(data map[string]interface{}) {
	albumsData, ok := data["albums"].(map[string]interface{})
	if !ok {
		c.logger.Println("Invalid albums data")
		return
	}

	albumsJSON, err := json.Marshal(albumsData)
	if err != nil {
		c.logger.Printf("Error converting albums data to JSON: %v", err)
		return
	}

	err = c.amqpPostAlbums(string(albumsJSON))
	if err != nil {
		c.logger.Printf("Error handling PostAlbums: %v", err)
		return
	}

	c.logger.Println("Successfully handled PostAlbums")
}

// handleGetAllAlbums handles the "GetAllAlbums" action by fetching and logging all albums.
func (c *MessageClient) handleGetAllAlbums() {
	albums, err := c.amqpGetAllAlbums()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	albumsJSON, err := json.Marshal(albums)
	if err != nil {
		c.logger.Printf("Error marshaling albums: %v", err)
		return
	}

	c.logger.Printf("Albums: %s", albumsJSON)
}

// handleGetDeleteAll handles the "GetDeleteAll" action by deleting all albums.
func (c *MessageClient) handleGetDeleteAll() {
	err := c.amqpGetDeleteAll()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
}

// handleGetAlbumByCode handles the "GetAlbumByCode" action by fetching and logging an album by its code.
func (c *MessageClient) handleGetAlbumByCode(data map[string]interface{}) {
	albumCode, ok := data["albumCode"].(string)
	if !ok {
		c.logger.Println("Invalid albumCode")
		return
	}

	album, err := c.amqpGetAlbumByCode(albumCode)
	if err != nil {
		c.logger.Printf("Error fetching album with Code %s: %v", albumCode, err)
		return
	}

	c.logger.Printf("Album: %+v", album)
}

// handleAddUser handles the "AddUser" action by adding a new user.
func (c *MessageClient) handleAddUser(data map[string]interface{}) {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		c.logger.Println("Invalid userEmail")
		return
	}
	name, ok := data["name"].(string)
	if !ok {
		c.logger.Println("Invalid name")
		return
	}
	password, ok := data["password"].(string)
	if !ok {
		c.logger.Println("Invalid password")
		return
	}

	err := c.amqpAddUser(userEmail, name, password)
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	c.logger.Printf("userEmail: %s; name: %s", userEmail, name)
}

// handleDeleteUser handles the "DeleteUser" action by deleting a user.
func (c *MessageClient) handleDeleteUser(data map[string]interface{}) {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		c.logger.Println("Invalid userEmail")
		return
	}

	err := c.amqpDeleteUser(userEmail)
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	c.logger.Printf("userEmail: %s", userEmail)
}

// handleFindUserToEmail handles the "FindUserToEmail" action by finding a user by their email.
func (c *MessageClient) handleFindUserToEmail(data map[string]interface{}) {
	userEmail, ok := data["userEmail"].(string)
	if !ok {
		c.logger.Println("Invalid userEmail")
		return
	}

	err := c.amqpFindUserToEmail(userEmail)
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	c.logger.Printf("userEmail: %s", userEmail)
}

// HandlerUpdateAlbum handles the "UpdateAlbum" action by updating an album's data.
func (c *MessageClient) HandlerUpdateAlbum(data map[string]interface{}) {
	newAlbumsData, ok := data["album"].(map[string]interface{})
	if !ok {
		c.logger.Println("Invalid albums data")
		return
	}

	albumsJSON, err := json.Marshal(newAlbumsData)
	if err != nil {
		c.logger.Printf("Error converting albums data to JSON: %v", err)
		return
	}

	err = c.amqpUpdateAlbum(string(albumsJSON))
	if err != nil {
		c.logger.Printf("Error handling UpdateAlbum: %v", err)
		return
	}

	c.logger.Println("Successfully handled UpdateAlbum")
}
