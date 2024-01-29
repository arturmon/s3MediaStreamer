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
	cleanS3Job := NewCleanS3Job(app)
	err = jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JobCleanTrackS3, cleanS3Job)
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

func NewCleanS3Job(app *app.App) *CleanS3Job {
	return &CleanS3Job{
		app: app,
	}
}

type CleanS3Job struct {
	app *app.App
}
