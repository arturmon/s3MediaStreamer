package tags

import (
	"fmt"
	"github.com/dhowden/tag"
	"github.com/google/uuid"
	"os"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"time"
)

func ReadTags(filename string, cfg *config.Config) (*model.Track, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tags, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	creatorUserUUID, err := uuid.Parse(cfg.AppConfig.Jobs.JobIDUserRun)
	if err != nil {
		return nil, err
	}

	// Convert the year to a time.Time value
	createdAt := time.Date(tags.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)

	if title, artist := tags.Title(), tags.Artist(); title == "" || artist == "" {
		return nil, fmt.Errorf("failed to read tags: empty title or artist")
	}
	discNumber, discTotal := tags.Disc()
	trackNumber, trackTotal := tags.Track()

	// Create and return the Track
	return &model.Track{
		ID:          uuid.New(),
		CreatedAt:   createdAt,
		UpdatedAt:   time.Now(),
		Album:       tags.Album(),
		AlbumArtist: tags.AlbumArtist(),
		Composer:    tags.Composer(),
		Genre:       tags.Genre(),
		Lyrics:      tags.Lyrics(),
		Title:       tags.Title(),
		Artist:      tags.Artist(),
		Year:        tags.Year(),
		Comment:     tags.Comment(),
		Disc:        discNumber,
		DiscTotal:   discTotal,
		Track:       trackNumber,
		TrackTotal:  trackTotal,
		Sender:      "",
		CreatorUser: creatorUserUUID,
		Likes:       false,
		S3Version:   "",
	}, nil
}
