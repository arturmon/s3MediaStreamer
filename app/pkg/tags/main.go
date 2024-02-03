package tags

import (
	"fmt"
	"io"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"time"

	"github.com/dhowden/tag"
	"github.com/google/uuid"
)

func ReadTags(reader io.ReadSeeker, cfg *config.Config) (*model.Track, error) {
	var track model.Track
	creatorUserUUID, err := uuid.Parse(cfg.AppConfig.Jobs.JobIDUserRun)
	if err != nil {
		return nil, err
	}

	tags, errTag := tag.ReadFrom(reader)
	if errTag != nil {
		return nil, errTag
	}

	// Convert the year to a time.Time value
	createdAt := time.Date(tags.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)

	if tags.Title() == "" || tags.Artist() == "" {
		return nil, fmt.Errorf("failed to read tags: empty title or artist")
	}

	track = model.Track{
		ID:          uuid.New(),
		CreatedAt:   createdAt,
		UpdatedAt:   time.Now(),
		Title:       tags.Title(),
		Artist:      tags.Artist(),
		Description: tags.Comment(),
		Sender:      "",
		CreatorUser: creatorUserUUID,
		Likes:       false,
		S3Version:   "",
	}

	return &track, nil
}
