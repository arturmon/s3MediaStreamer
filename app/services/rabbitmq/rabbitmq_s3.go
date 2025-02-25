package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"s3MediaStreamer/app/model"
)

func (s *Service) S3BucketActionEventQueue(ctx context.Context, messageBody map[string]interface{}) {
	// Extract the event from the message
	s3event, errExtract := s.extractRecordsEvent(messageBody)
	if errExtract != nil {
		s.logger.Infof("Error extracting message: %v", errExtract)
		return
	}

	action, ok := messageBody["EventName"].(string)
	if !ok {
		s.logger.Error("Invalid action field")
		return
	}

	// Process based on the action
	switch action {
	case "s3:ObjectRemoved:Delete":
		err := s.deleteEvent(ctx, s3event)
		if err != nil {
			s.logger.Errorf("Error handling deleteEvent: %v", err)
			return
		}
	case "s3:ObjectCreated:Put":
		err := s.putEvent(ctx, s3event)
		if err != nil {
			s.logger.Errorf("Error handling putEvent: %v", err)
			return
		}
	default:
		s.logger.Debugf("Event: %s not processed", action)
	}
}

// extractRecordsEvent extracts event data from the message
func (s *Service) extractRecordsEvent(data map[string]interface{}) (*model.MessageBody, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &model.MessageBody{}, err
	}

	// Unmarshal the JSON data into a Records struct
	var messageBody model.MessageBody
	err = json.Unmarshal(jsonData, &messageBody)
	if err != nil {
		return &model.MessageBody{}, err
	}

	// Check if Records array is not empty
	if len(messageBody.Records) == 0 {
		return &model.MessageBody{}, fmt.Errorf("empty records array")
	}

	return &messageBody, nil
}

// deleteEvent handles the event where an object is removed from S3
func (s *Service) deleteEvent(ctx context.Context, s3event *model.MessageBody) error {
	err := s.s3.DeleteS3Version(ctx, s3event.Records[0].S3.Object.VersionID)
	if err != nil {
		return fmt.Errorf("error deleting from S3: %w", err)
	}
	err = s.track.CleanTracks(ctx)
	if err != nil {
		s.logger.Errorf("Error deleting filename: %v\n", err)
		return err
	}
	return nil
}

// putEvent handles the event where an object is created or updated in S3
func (s *Service) putEvent(ctx context.Context, s3event *model.MessageBody) error {
	err := s.checkObjectS3(ctx, s3event)
	if err != nil {
		return err
	}
	return nil
}

// checkObjectS3 checks the object in S3 and processes it
func (s *Service) checkObjectS3(ctx context.Context, object *model.MessageBody) error {
	// Download file data from S3
	fileName, err := s.s3.DownloadFilesS3(ctx, object.Key)
	if err != nil {
		s.logger.Errorf("Error downloading file %s from S3: %v\n", object.Key, err)
		return err
	}

	// Create a Track from the file data
	objectTags, errReadTags := s.tags.ReadTags(fileName)
	err = s.s3.CleanTemplateFile(fileName)
	if err != nil {
		return err
	}
	if errReadTags != nil {
		s.logger.Errorf("Error processing file: %s Error: %v\n", object.Records[0].S3.Object.Key, errReadTags)
		return err
	}
	err = s.checkIfTrackExists(ctx, objectTags, object.Records[0].S3.Object.VersionID)
	if err != nil {
		s.logger.Errorf("%v\n", err)
	}
	return nil
}

// checkIfTrackExists checks if the track already exists in the database
func (s *Service) checkIfTrackExists(ctx context.Context, track *model.Track, s3id string) error {
	_, err := s.track.GetTracksByColumns(ctx, track.Title, "title")
	if err != nil {
		if s.isNoRecordsFound(err.Error()) {
			return s.handleNonexistentTrack(ctx, track, s3id)
		}
		return fmt.Errorf("error getting existing tracks: %w", err)
	}
	return nil
}

// handleNonexistentTrack handles the case where a track is not found in the database
func (s *Service) handleNonexistentTrack(ctx context.Context, track *model.Track, s3id string) error {
	s.logger.Infof("Track '%s' not found in the database.\n", track.Title)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := s.track.CreateTracks(ctx, existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	} else {
		s.logger.Errorf("Track '%s' already exists\n", track.Artist)
	}

	err := s.s3.AddS3Version(ctx, existingTracksSlice[0].ID.String(), s3id)
	if err != nil {
		return fmt.Errorf("error adding S3 version: %w", err)
	}
	s.logger.Infof("Track '%s' saved to the database.\n", track.Artist)
	return nil
}

// isNoRecordsFound checks if the error message indicates no records were found
func (s *Service) isNoRecordsFound(errMsg string) bool {
	// Use a regex pattern to check for the "no records found" message
	pattern := `no records found (.+)`
	matched, _ := regexp.MatchString(pattern, errMsg)
	return matched
}
