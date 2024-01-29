package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7/pkg/notification"
	"regexp"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/tags"
	"strings"
	"sync"
)

type eventS3 struct {
	Name      string
	Type      string
	VersionID string
}

func HandlersWatherS3(ctx context.Context, wg *sync.WaitGroup, app *app.App) {
	defer wg.Done()
	notificationChannel, err := app.S3.MonitoringDirectoryS3(ctx)
	if err != nil {
		app.Logger.Printf("Error getting objectInfo: %v\n", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case notificationInfo, ok := <-notificationChannel:
			if !ok {
				// Channel closed, return or take appropriate action
				return
			}

			if notificationInfo.Err != nil {
				app.Logger.Printf("Error receiving notification: %v\n", notificationInfo.Err)
				continue
			}

			app.Logger.Println(notificationInfo)

			structS3, errStruct := fillStruct(notificationInfo) // Adjust this to use notificationInfo if needed
			if errStruct != nil {
				app.Logger.Printf("Error extracting filename: %v\n", errStruct)
				continue
			}

			app.Logger.Printf("Track name: %v\n", structS3.Name)

			switch structS3.Type {
			case "Put":
				err := checkObjectS3(ctx, structS3, app)
				if err != nil {
					app.Logger.Printf("Error handling object: %v\n", err)
					continue
				}
			case "Delete":
				err := app.Storage.Operations.DeleteTracks(structS3.VersionID, "s3Version")
				if err != nil {
					app.Logger.Printf("Error deleting filename: %v\n", err)
				}
				app.Logger.Printf("Delete track: %s\n", structS3.Name)
			default:
				app.Logger.Printf("Unsupported event type: %s\n", structS3.Type)
			}
		}
	}
}

func fillStruct(event notification.Info) (*eventS3, error) {
	if len(event.Records) == 0 {
		return nil, fmt.Errorf("no records in the event")
	}
	// Assuming the file name is in the first record
	firstRecord := event.Records[0]

	// Check if S3 information is present
	if firstRecord.S3.Object.Key == "" {
		return nil, fmt.Errorf("no S3 key in the event")
	}
	parts := strings.Split(firstRecord.EventName, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid event type format")
	}

	if firstRecord.S3.Object.VersionID == "" {
		return nil, fmt.Errorf("no S3 Etag in the event")
	}
	versionID := firstRecord.S3.Object.VersionID

	return &eventS3{
		Name:      firstRecord.S3.Object.Key,
		Type:      parts[2],
		VersionID: versionID,
	}, nil
}

// checkObjectS3 checks and processes an object in S3.
func checkObjectS3(ctx context.Context, object *eventS3, app *app.App) error {
	// Download file data from S3
	fileData, err := app.S3.DownloadFilesS3(ctx, object.Name)
	if err != nil {
		app.Logger.Printf("Error downloading file %s from S3: %v\n", object.Name, err)
		return err
	}

	// Create a Track from the file data
	objectTags, errReadTags := tags.ReadTags(bytes.NewReader(fileData), app.Cfg, app.Logger)
	if errReadTags != nil {
		app.Logger.Printf("Error processing file: %s Error: %v\n", object.Name, errReadTags)
		return err
	}
	objectTags.S3Version = object.VersionID
	objectTags.Sender = "Event"

	err = checkIfTrackExists(objectTags, app)
	if err != nil {
		app.Logger.Printf("%v\n", err)
	}
	return nil
}

func checkIfTrackExists(track *model.Track, app *app.App) error {
	_, err := app.Storage.Operations.GetTracksByColumns(track.Title, "title")
	if err != nil {
		if isNoRecordsFound(err.Error()) {
			return handleNonexistentTrack(track, app)
		}
		return fmt.Errorf("error getting existing albums for Artist %s: %w", track.Title, err)
	}
	return nil
}

func handleNonexistentTrack(track *model.Track, app *app.App) error {
	app.Logger.Printf("Track Artist: %s not found in the database. Continuing...\n", track.Title)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := app.Storage.Operations.CreateTracks(existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	} else {
		app.Logger.Printf("Track with Artist %s already exists\n", track.Artist)
	}
	app.Logger.Printf("Track Artist: %s added\n", track.Artist)
	return nil
}

func isNoRecordsFound(errMsg string) bool {
	// Use a regex pattern to check for the "no records found" message
	// Adjust the pattern based on the specific structure of your error messages
	pattern := `no records found (.+)`
	matched, _ := regexp.MatchString(pattern, errMsg)
	return matched
}
