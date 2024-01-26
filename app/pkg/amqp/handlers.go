package amqp

import (
	"encoding/json"
	"fmt"
)

// handlePostTracks handles the "PostTracks" action by processing the incoming track data.
func (c *MessageClient) handlePostTracks(data map[string]interface{}) error {
	albumsData, ok := data["tracks"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid tracks data")
	}

	albumsJSON, err := json.Marshal(albumsData)
	if err != nil {
		return err
	}

	return c.amqpPostTracks(string(albumsJSON))
}

// handleGetAllTracks handles the "GetAllTracks" action by fetching and logging all tracks.
func (c *MessageClient) handleGetAllTracks(page, pageSize int, sortBy, sortOrder, filter string) error {
	tracks, totalRows, err := c.amqpGetAllTracks(page, pageSize, sortBy, sortOrder, filter)
	if err != nil {
		return err
	}

	albumsJSON, err := json.Marshal(tracks)
	if err != nil {
		return err
	}
	c.logger.Debugf("TotalRows: %d", totalRows)
	c.logger.Debugf("Tracks: %s", albumsJSON)
	return nil
}

// handleGetDeleteAll handles the "GetDeleteAll" action by deleting all tracks.
func (c *MessageClient) handleGetDeleteAll() error {
	err := c.amqpGetDeleteAll()
	if err != nil {
		return err
	}

	return nil
}

// handleGetTrackByCode handles the "GetTrackByCode" action by fetching and logging an track by its code.
func (c *MessageClient) handleGetTrackByCode(data map[string]interface{}) error {
	albumCode, ok := data["albumCode"].(string)
	if !ok {
		return fmt.Errorf("invalid albumCode")
	}

	track, err := c.amqpGetTrackByCode(albumCode)
	if err != nil {
		return err
	}

	c.logger.Printf("Track: %+v", track)
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

// HandlerUpdateTrack handles the "UpdateTrack" action by updating an track's data.
func (c *MessageClient) handleUpdateTrack(data map[string]interface{}) error {
	newTracksData, ok := data["track"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid tracks data")
	}

	albumsJSON, err := json.Marshal(newTracksData)
	if err != nil {
		return err
	}

	return c.amqpUpdateTrack(string(albumsJSON))
}
