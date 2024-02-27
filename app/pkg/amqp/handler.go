package amqp

import (
	"context"
	"fmt"
	"regexp"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/tags"
)

func (c *MessageClient) deleteEvent(ctx context.Context, s3event *MessageBody) error {
	err := c.storage.Operations.DeleteTracks(ctx, s3event.Records[0].S3.Object.VersionID, "s3Version")
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
	objectTags.S3Version = object.Records[0].S3.Object.VersionID
	objectTags.Sender = "Event"

	err = checkIfTrackExists(ctx, objectTags, c)
	if err != nil {
		c.logger.Printf("%v\n", err)
	}
	return nil
}

func checkIfTrackExists(ctx context.Context, track *model.Track, c *MessageClient) error {
	_, err := c.storage.Operations.GetTracksByColumns(ctx, track.Title, "title")
	if err != nil {
		if isNoRecordsFound(err.Error()) {
			return handleNonexistentTrack(ctx, track, c)
		}
		return fmt.Errorf("error getting existing albums for Artist %s: %w", track.Title, err)
	}
	return nil
}

func handleNonexistentTrack(ctx context.Context, track *model.Track, c *MessageClient) error {
	c.logger.Printf("Track Artist: %s not found in the database.\n", track.Title)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := c.storage.Operations.CreateTracks(ctx, existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	} else {
		c.logger.Printf("Track with Artist %s already exists\n", track.Artist)
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
