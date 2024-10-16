package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"s3MediaStreamer/app/model"
)

func (s *Service) HandleMessage(ctx context.Context, messageBody map[string]interface{}) {
	s3event, errExtract := s.extractRecordsEvent(messageBody)
	if errExtract != nil {
		s.logger.Infof("Error extract message: %v", errExtract)
		return
	}

	action, ok := messageBody["EventName"].(string)
	if !ok {
		s.logger.Error("Invalid action field")
		return
	}

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

func (s *Service) deleteEvent(ctx context.Context, s3event *model.MessageBody) error {
	err := s.s3.DeleteS3Version(ctx, s3event.Records[0].S3.Object.VersionID)
	if err != nil {
		return fmt.Errorf("error add s3 track_handler: %w", err)
	}
	err = s.track.CleanTracks(ctx)
	if err != nil {
		s.logger.Errorf("Error deleting filename: %v\n", err)
		return err
	}
	return nil
}

func (s *Service) putEvent(ctx context.Context, s3event *model.MessageBody) error {
	err := s.checkObjectS3(ctx, s3event, s)
	if err != nil {
		return err
	}
	return nil
}

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
		return &model.MessageBody{}, err
	}

	return &messageBody, nil
}

func (s *Service) checkObjectS3(ctx context.Context, object *model.MessageBody, c *Service) error {
	// Download file data from S3
	fileName, err := s.s3.DownloadFilesS3(ctx, object.Key)
	if err != nil {
		c.logger.Errorf("Error downloading file %s from S3: %v\n", object.Key, err)
		return err
	}

	// Create a Track from the file data
	objectTags, errReadTags := s.tags.ReadTags(fileName)
	err = s.s3.CleanTemplateFile(fileName)
	if err != nil {
		return err
	}
	if errReadTags != nil {
		c.logger.Errorf("Error processing file: %s Error: %v\n", object.Records[0].S3.Object.Key, errReadTags)
		return err
	}
	err = s.checkIfTrackExists(ctx, objectTags, object.Records[0].S3.Object.VersionID, c)
	if err != nil {
		c.logger.Errorf("%v\n", err)
	}
	return nil
}

func (s *Service) checkIfTrackExists(ctx context.Context, track *model.Track, s3id string, c *Service) error {
	_, err := s.track.GetTracksByColumns(ctx, track.Title, "title")
	if err != nil {
		if s.isNoRecordsFound(err.Error()) {
			return s.handleNonexistentTrack(ctx, track, s3id, c)
		}
		return fmt.Errorf("error getting existing albums for Artist %s: %w", track.Title, err)
	}
	return nil
}

func (s *Service) handleNonexistentTrack(ctx context.Context, track *model.Track, s3id string, c *Service) error {
	c.logger.Infof("Track Artist: %s not found in the database.\n", track.Title)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := s.track.CreateTracks(ctx, existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track_handler: %w", err)
		}
	} else {
		c.logger.Errorf("Track with Artist %s already exists\n", track.Artist)
	}

	err := s.s3.AddS3Version(ctx, existingTracksSlice[0].ID.String(), s3id)
	if err != nil {
		return fmt.Errorf("error add s3 track_handler: %w", err)
	}
	c.logger.Infof("Track Artist: %s save to database.\n", track.Artist)
	return nil
}

func (s *Service) isNoRecordsFound(errMsg string) bool {
	// Use a regex pattern to check for the "no records found" message
	// Adjust the pattern based on the specific structure of your error messages
	pattern := `no records found (.+)`
	matched, _ := regexp.MatchString(pattern, errMsg)
	return matched
}
