package tags

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3MediaStreamer/app/internal/config"
	"s3MediaStreamer/app/model"
	"time"

	"github.com/dhowden/tag"
	"github.com/google/uuid"
	"github.com/mewkiz/flac"
	"github.com/tcolgate/mp3"
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

	fileExtension := filepath.Ext(filename)

	var (
		duration   time.Duration
		sampleRate uint32
		bitrate    uint32
	)

	switch fileExtension {
	case ".flac":
		sampleRate, duration, bitrate = getSampleRate(filename)
	case ".mp3":
		sampleRate, duration, bitrate, err = getMp3Info(f)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported audio format")
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
		Duration:    duration,
		SampleRate:  sampleRate,
		Bitrate:     bitrate,
	}, nil
}

func getSampleRate(fileName string) (uint32, time.Duration, uint32) {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		panic(err)
	}
	data := f.Info

	duration := time.Duration(float64(f.Info.NSamples) / float64(f.Info.SampleRate) * float64(time.Second))
	bitrate := uint32(float64(data.NSamples) * float64(data.BitsPerSample) / duration.Seconds() / millisecondsPerSecond)
	return data.SampleRate, duration, bitrate
}

func getMp3Info(f io.Reader) (uint32, time.Duration, uint32, error) {
	dec := mp3.NewDecoder(f)
	var frame mp3.Frame
	var duration time.Duration
	var sampleRate uint32
	var bitrate uint32

	skipped := 0
	for {
		if err := dec.Decode(&frame, &skipped); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, 0, 0, err
		}
		duration += frame.Duration()
		sampleRate = uint32(frame.Header().SampleRate())
		bitrate = uint32(frame.Header().BitRate() / millisecondsPerSecond)
	}
	return sampleRate, duration, bitrate, nil
}
