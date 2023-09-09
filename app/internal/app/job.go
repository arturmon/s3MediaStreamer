package app

import (
	"github.com/bamzi/jobrunner"
)

func InitJob(app *App) error {
	jobrunner.Start()
	openAIJob := NewOpenAIJob(app)
	err := jobrunner.Schedule(app.cfg.AppConfig.OpenAI.JonRun, openAIJob)
	if err != nil {
		app.logger.Error("Failed to schedule job:", err)
		return err
	}
	cleanJob := NewCleanJob(app)
	err = jobrunner.Schedule(app.cfg.AppConfig.OpenAI.JonRun, cleanJob)
	if err != nil {
		app.logger.Error("Failed to schedule job:", err)
		return err
	}
	return nil
}

func NewOpenAIJob(app *App) *OpenAIJob {
	return &OpenAIJob{
		app: app,
	}
}

// OpenAIJob Job Specific Functions.
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
