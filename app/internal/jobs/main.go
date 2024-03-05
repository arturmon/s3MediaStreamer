package jobs

import (
	"s3MediaStreamer/app/internal/app"

	"github.com/bamzi/jobrunner"
)

func InitJob(app *app.App) error {
	jobrunner.Start()
	cleanS3Job := NewCleanS3Job(app)
	err := jobrunner.Schedule(app.Cfg.AppConfig.Jobs.JobCleanTrackS3, cleanS3Job)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	cleanSessionJob := NewCleanOldSessionJob(app)
	err = jobrunner.Schedule(app.Cfg.Session.SessionPeriodClean, cleanSessionJob)
	if err != nil {
		app.Logger.Error("Failed to schedule job:", err)
		return err
	}
	return nil
}

func NewCleanS3Job(app *app.App) *CleanS3Job {
	return &CleanS3Job{
		app: app,
	}
}

type CleanS3Job struct {
	app *app.App
}

func NewCleanOldSessionJob(app *app.App) *CleanOldSessionJob {
	return &CleanOldSessionJob{
		app: app,
	}
}

type CleanOldSessionJob struct {
	app *app.App
}
