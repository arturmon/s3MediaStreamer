package amqp

import (
	"encoding/json"
)

func (c *AMQPClient) handlePostAlbums(data map[string]interface{}) {
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

func (c *AMQPClient) handleGetAllAlbums() {
	albums, err := c.amqpGetAllAlbums()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	c.logger.Printf("Albums: %s", albums)
}

func (c *AMQPClient) handleGetDeleteAll() {
	err := c.amqpGetDeleteAll()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
}

func (c *AMQPClient) handleGetAlbumByCode(data map[string]interface{}) {
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

func (c *AMQPClient) handleAddUser(data map[string]interface{}) {
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

func (c *AMQPClient) handleDeleteUser(data map[string]interface{}) {
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

func (c *AMQPClient) handleFindUserToEmail(data map[string]interface{}) {
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
