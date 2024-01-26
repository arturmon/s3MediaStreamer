package jobs

import (
	"skeleton-golange-application/app/internal/app"

	"github.com/bamzi/jobrunner"
)

func InitJob(app *app.App) error {
	jobrunner.Start()
	openAIJob := NewOpenAIJob(app)
	err := jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JonRun, openAIJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	cleanJob := NewCleanJob(app)
	err = jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JobCleanChart, cleanJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	cleanTracksJob := NewCleanTracksJob(app)
	err = jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JobCleanTrackPathNull, cleanTracksJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	return nil
}

func NewOpenAIJob(app *app.App) *OpenAIJob {
	return &OpenAIJob{
		app: app,
	}
}

type OpenAIJob struct {
	app *app.App
}

func NewCleanJob(app *app.App) *CleanChartJob {
	return &CleanChartJob{
		app: app,
	}
}

type CleanChartJob struct {
	app *app.App
}

func NewCleanTracksJob(app *app.App) *CleanTracksJob {
	return &CleanTracksJob{
		app: app,
	}
}

type CleanTracksJob struct {
	app *app.App
}
