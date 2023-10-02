package app

import (
	"github.com/bamzi/jobrunner"
)

func InitJob(app *App) error {
	jobrunner.Start()
	openAIJob := NewOpenAIJob(app)
	err := jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JonRun, openAIJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	cleanJob := NewCleanJob(app)
	err = jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JonRun, cleanJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	readFolderJob := NewReadFolderJob(app)
	err = jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JobReadFolder, readFolderJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	return nil
}

func NewOpenAIJob(app *App) *OpenAIJob {
	return &OpenAIJob{
		app: app,
	}
}

type OpenAIJob struct {
	app *App
}

func NewCleanJob(app *App) *CleanJob {
	return &CleanJob{
		app: app,
	}
}

type CleanJob struct {
	app *App
}

func NewReadFolderJob(app *App) *ReadFolderJob {
	return &ReadFolderJob{
		app: app,
	}
}

type ReadFolderJob struct {
	app *App
}
