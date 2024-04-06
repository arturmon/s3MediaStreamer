package amqp

import (
	"context"
	"fmt"
	"regexp"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/pkg/tags"
)

func (c *MessageClient) deleteEvent(ctx context.Context, s3event *MessageBody) error {
	err := c.storage.Operations.DeleteS3Version(ctx, s3event.Records[0].S3.Object.VersionID)
	if err != nil {
		return fmt.Errorf("error add s3 track: %w", err)
	}
	err = c.storage.Operations.CleanTracks(ctx)
	if err != nil {
		c.logger.Printf("Error deleting filename: %v\n", err)
		return err
	}
	return nil
}

func (c *MessageClient) putEvent(ctx context.Context, s3event *MessageBody) error {
	err := checkObjectS3(ctx, s3event, c)
	if err != nil {
		return err
	}
	return nil
}

func checkObjectS3(ctx context.Context, object *MessageBody, c *MessageClient) error {
	// Download file data from S3
	fileName, err := c.s3Handler.DownloadFilesS3(ctx, object.Key)
	if err != nil {
		c.logger.Printf("Error downloading file %s from S3: %v\n", object.Key, err)
		return err
	}

	// Create a Track from the file data
	objectTags, errReadTags := tags.ReadTags(fileName, c.cfg)
	err = c.s3Handler.CleanTemplateFile(fileName)
	if err != nil {
		return err
	}
	if errReadTags != nil {
		c.logger.Printf("Error processing file: %s Error: %v\n", object.Records[0].S3.Object.Key, errReadTags)
		return err
	}
	//TODO
	//objectTags.S3Version = object.Records[0].S3.Object.VersionID

	err = checkIfTrackExists(ctx, objectTags, object.Records[0].S3.Object.VersionID, c)
	if err != nil {
		c.logger.Printf("%v\n", err)
	}
	return nil
}

func checkIfTrackExists(ctx context.Context, track *model.Track, s3id string, c *MessageClient) error {
	_, err := c.storage.Operations.GetTracksByColumns(ctx, track.Title, "title")
	if err != nil {
		if isNoRecordsFound(err.Error()) {
			return handleNonexistentTrack(ctx, track, s3id, c)
		}
		return fmt.Errorf("error getting existing albums for Artist %s: %w", track.Title, err)
	}
	return nil
}

func handleNonexistentTrack(ctx context.Context, track *model.Track, s3id string, c *MessageClient) error {
	c.logger.Printf("Track Artist: %s not found in the database.\n", track.Title)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := c.storage.Operations.CreateTracks(ctx, existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	} else {
		c.logger.Printf("Track with Artist %s already exists\n", track.Artist)
	}

	err := c.storage.Operations.AddS3Version(ctx, existingTracksSlice[0].ID.String(), s3id)
	if err != nil {
		return fmt.Errorf("error add s3 track: %w", err)
	}
	c.logger.Printf("Track Artist: %s save to database.\n", track.Artist)
	return nil
}

func isNoRecordsFound(errMsg string) bool {
	// Use a regex pattern to check for the "no records found" message
	// Adjust the pattern based on the specific structure of your error messages
	pattern := `no records found (.+)`
	matched, _ := regexp.MatchString(pattern, errMsg)
	return matched
}
