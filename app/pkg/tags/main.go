package tags

import (
	"os"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/logging"
	"time"

	"github.com/bojanz/currency"
	"github.com/dhowden/tag"
	"github.com/google/uuid"
)

func ReadTags(filePath string, cfg *config.Config, logger *logging.Logger) (*model.Track, error) {
	var track model.Track
	creatorUserUUID, err := uuid.Parse(cfg.AppConfig.Jobs.JobIDUserRun)
	if err != nil {
		return nil, err
	}

	file, errOpen := os.Open(filePath)
	if errOpen != nil {
		logger.Errorf("Error opening file: %v\n", err)
		return nil, errOpen
	}

	tags, errTag := tag.ReadFrom(file)
	if errTag != nil {
		logger.Errorf("Error reading tags %s: %v\n", filePath, errTag)
		return nil, errTag
	}

	price, errPrice := currency.NewAmount("0", "EUR")
	if errPrice != nil {
		return nil, errPrice
	}

	fileInfo, _ := file.Stat()
	createdAt := fileInfo.ModTime()
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			return
		}
	}(file)
	track = model.Track{
		ID:          uuid.New(),
		CreatedAt:   createdAt,
		UpdatedAt:   time.Now(),
		Title:       tags.Title(),
		Artist:      tags.Artist(),
		Price:       price,
		Code:        randomString(lengthRandomGenerateCode),
		Description: tags.Comment(),
		Sender:      cfg.AppConfig.Jobs.SystemWriteUser,
		CreatorUser: creatorUserUUID,
		Likes:       false,
		Path:        filePath,
	}

	return &track, nil
}

func randomString(length int) string {
	return uuid.NewString()[:length]
}
