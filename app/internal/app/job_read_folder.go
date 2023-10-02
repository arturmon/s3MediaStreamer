package app

import (
	"os"
	"path/filepath"
	"reflect"
	"skeleton-golange-application/model"
	"time"

	"github.com/bojanz/currency"
	"github.com/dhowden/tag"
	"github.com/google/uuid"
)

func (j *ReadFolderJob) Run() {
	j.app.Logger.Println("init read music folders...")

	err := InitReadFolders(j.app)
	if err != nil {
		j.app.Logger.Fatal(err)
	}
	j.app.Logger.Println("complete read music folders")
}

func InitReadFolders(app *App) error {
	currentDir, err := os.Getwd()
	if err != nil {
		app.Logger.Printf("Error getting current directory: %v\n", err)
		return err
	}

	app.Logger.Printf("Current directory: %s\n", currentDir)
	albums, err := ReadFolders(currentDir+"/"+app.GetCfg().AppConfig.Jobs.JobReadPath, app)
	diff, errDiff := AlbumsStore(app, albums)
	if errDiff != nil {
		return err
	}
	app.Logger.Println(diff)
	return nil
}

func ReadFolders(folderPath string, app *App) ([]model.Album, error) {
	var albums []model.Album

	creatorUserUUID, err := uuid.Parse(app.Cfg.AppConfig.Jobs.JobIDUserRun)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(folderPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			app.GetLogger().Errorf("Error opening file: %v\n", err)
			return err
		}

		tags, errTag := tag.ReadFrom(file)
		if errTag != nil {
			app.GetLogger().Errorf("Error reading tags %s: %v\n", filePath, errTag)
			return errTag
		}

		price, errPrice := currency.NewAmount("0", "EUR")
		if errPrice != nil {
			return errPrice
		}

		createdAt := fileInfo.ModTime()

		defer file.Close()
		album := model.Album{
			ID:          uuid.New(),
			CreatedAt:   createdAt,
			UpdatedAt:   time.Now(),
			Title:       tags.Title(),
			Artist:      tags.Artist(),
			Price:       price,
			Code:        randomString(lengthRandomGenerateCode),
			Description: tags.Comment(),
			Sender:      app.Cfg.AppConfig.Jobs.SystemWriteUser,
			CreatorUser: creatorUserUUID,
			Likes:       false,
			Path:        filePath,
		}
		albums = append(albums, album)

		return nil
	})

	if err != nil {
		app.GetLogger().Errorf("Error traversing folder: %v\n", err)
	}

	return albums, nil
}

func AlbumsStore(app *App, diskAlbums []model.Album) ([]model.Album, error) {
	// Retrieve all albums from the database

	dbAlbums, err := app.Storage.Operations.GetAllAlbums()
	if err != nil {
		return nil, err
	}

	// Create a map of album IDs for efficient lookup
	dbAlbumsMap := make(map[uuid.UUID]model.Album)
	for _, dbAlbum := range dbAlbums {
		dbAlbumsMap[dbAlbum.ID] = dbAlbum
	}

	// Create a slice to hold the differences
	var differences []model.Album

	// Iterate over the disk albums and compare with the database albums
	for _, diskAlbum := range diskAlbums {
		dbAlbum, exists := dbAlbumsMap[diskAlbum.ID]
		if !exists {
			// Disk album doesn't exist in the database, add it to differences
			differences = append(differences, diskAlbum)
		} else if !reflect.DeepEqual(diskAlbum, dbAlbum) {
			// Disk album differs from the database album, add it to differences
			differences = append(differences, diskAlbum)
		}
	}

	if len(differences) == 0 {
		app.Logger.Println("Albums are equal.")
	} else {
		app.Logger.Println("Albums are not equal.")
		app.Logger.Debugf("Differences: %+v\n", differences)
		// Store the differences
		err = app.Storage.Operations.CreateAlbums(differences)
		if err != nil {
			return nil, err
		}
	}

	// Store the differences if needed
	if len(differences) > 0 {
		err = app.Storage.Operations.CreateAlbums(differences)
		if err != nil {
			return nil, err
		}
	}
	return differences, nil
}

func randomString(length int) string {
	return uuid.NewString()[:length]
}
