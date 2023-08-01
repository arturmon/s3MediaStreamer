package amqp

import (
	"encoding/json"
)

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

func (c *MessageClient) handleGetAllAlbums() {
	albums, err := c.amqpGetAllAlbums()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
	c.logger.Printf("Albums: %s", albums)
}

func (c *MessageClient) handleGetDeleteAll() {
	err := c.amqpGetDeleteAll()
	if err != nil {
		c.logger.Printf("Error: %v", err)
		return
	}
}

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
