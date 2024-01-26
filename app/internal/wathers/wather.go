package wathers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"skeleton-golange-application/app/internal/app"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/tags"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

/*
func MonitorDirectory(ctx context.Context, app *app.App, wg *sync.WaitGroup, fileQueue chan string) {
	defer wg.Done()

	// Create a new FS watcher.
	watcher, errWatcher := fsnotify.NewWatcher()
	if errWatcher != nil {
		app.Logger.Println("Error creating watcher:", errWatcher)
		return
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			return
		}
	}(watcher)

	// Add the directory to the watcher.
	servedDirectory := FindLocalDir(app.Cfg)
	err := watcher.Add(servedDirectory)
	if err != nil {
		app.Logger.Println("Error adding directory to watcher:", err)
		return
	}

	app.Logger.Println("Monitoring directory:", servedDirectory)

	// Start the event handling loop.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return // Watcher closed
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				app.Logger.Printf("File Created: %v\n", event.Name)
				// Add the file to the queue for processing
				fileQueue <- event.Name
				go ProcessFiles(ctx, app, fileQueue)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				app.Logger.Printf("File Removed: %v\n", event.Name)
				_, err = app.Storage.Operations.GetTracksByColumns(event.Name, "path")
				if err == nil {
					err = app.Storage.Operations.DeleteTracks(event.Name, "path")
					if err != nil {
						return
					}
				}
			}

		case errEvent, ok := <-watcher.Errors:
			if !ok {
				return // Watcher closed
			}
			app.Logger.Printf("Error: %v\n", errEvent)

		case <-ctx.Done():
			return // Context canceled, stop monitoring
		}
	}
}

*/

func MonitorDirectory(ctx context.Context, app *app.App, wg *sync.WaitGroup, fileQueue chan string) {
	defer wg.Done()

	watcher, err := createWatcher(app)
	if err != nil {
		app.Logger.Println("Error creating watcher:", err)
		return
	}
	defer closeWatcher(watcher)

	servedDirectory := FindLocalDir(app.Cfg)
	if err = addDirectoryToWatcher(watcher, servedDirectory); err != nil {
		app.Logger.Println("Error adding directory to watcher:", err)
		return
	}

	app.Logger.Println("Monitoring directory:", servedDirectory)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return // Watcher closed
			}
			handleFileEvent(ctx, app, event, fileQueue)

		case errEvent, ok := <-watcher.Errors:
			if !ok {
				return // Watcher closed
			}
			handleErrorEvent(app, errEvent)

		case <-ctx.Done():
			return // Context canceled, stop monitoring
		}
	}
}

func createWatcher(_ *app.App) (*fsnotify.Watcher, error) {
	return fsnotify.NewWatcher()
}

func closeWatcher(watcher *fsnotify.Watcher) {
	_ = watcher.Close()
}

func addDirectoryToWatcher(watcher *fsnotify.Watcher, directory string) error {
	return watcher.Add(directory)
}

func handleFileEvent(ctx context.Context, app *app.App, event fsnotify.Event, fileQueue chan string) {
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		app.Logger.Printf("File Created: %v\n", event.Name)
		fileQueue <- event.Name
		go ProcessFiles(ctx, app, fileQueue)

	case event.Op&fsnotify.Remove == fsnotify.Remove:
		app.Logger.Printf("File Removed: %v\n", event.Name)
		handleFileRemoval(app, event.Name)
	}
}

func handleFileRemoval(app *app.App, fileName string) {
	if _, err := app.Storage.Operations.GetTracksByColumns(fileName, "path"); err == nil {
		if err = app.Storage.Operations.DeleteTracks(fileName, "path"); err != nil {
			return
		}
	}
}

func handleErrorEvent(app *app.App, errEvent error) {
	app.Logger.Printf("Error: %v\n", errEvent)
}

func FindLocalDir(cfg *config.Config) string {
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	servedDirectory := filepath.Join(currentDir, cfg.AppConfig.MusicPath)
	return servedDirectory
}

func ProcessFiles(ctx context.Context, app *app.App, fileQueue chan string) {
	for {
		select {
		case fileName := <-fileQueue:
			if err := processFile(ctx, app, fileName); err != nil {
				app.Logger.Printf("Error processing file %v: %v\n", fileName, err)
			}

		case <-ctx.Done():
			return // Stop the goroutine when the context is canceled
		}
	}
}

func processFile(ctx context.Context, app *app.App, fileName string) error {
	track, err := processFileWithRetries(ctx, fileName, app)
	if err != nil {
		return fmt.Errorf("error reading tags for %v: %w", fileName, err)
	}

	err = handleExistingTrack(app, track)
	if err != nil {
		return err
	}

	return nil
}

func handleExistingTrack(app *app.App, track *model.Track) error {
	return checkIfTrackExists(app, track)
}

func checkIfTrackExists(app *app.App, track *model.Track) error {
	_, err := app.Storage.Operations.GetTracksByColumns(track.Path, "path")
	if err != nil {
		if strings.Contains(err.Error(), "no records found") {
			return handleNonexistentTrack(app, track)
		}
		return fmt.Errorf("error getting existing albums for code %s: %w", track.Code, err)
	}

	return nil
}

func handleNonexistentTrack(app *app.App, track *model.Track) error {
	app.Logger.Printf("Track code:%s not found in the database. Continuing...\n", track.Code)

	existingTracksSlice := []model.Track{*track}
	if len(existingTracksSlice) == 1 {
		if err := app.Storage.Operations.CreateTracks(existingTracksSlice); err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	} else {
		app.Logger.Printf("Track with code %s already exists\n", track.Code)
	}

	return nil
}

func processFileWithRetries(ctx context.Context, fileName string, app *app.App) (*model.Track, error) {
	retryInterval := initialInterval * time.Second

	for retries := 0; retries < maxRetries; retries++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err() // Context canceled, stop retrying
		default:
			// Try to process the file
			track, errReadTags := tags.ReadTags(fileName, app.Cfg, app.Logger)
			if errReadTags != nil {
				app.Logger.Printf("Error reading tags for %v: %v", fileName, errReadTags)
			} else {
				// File processed successfully, exit the retry loop
				return track, nil
			}

			// Wait for a while before retrying
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("file %s could not be processed after %d retries", fileName, maxRetries)
}
