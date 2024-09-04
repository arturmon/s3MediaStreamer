package jobs

import (
	"s3MediaStreamer/app/internal/app"
	"time"

	"github.com/bamzi/jobrunner"
)

// InitJob initializes and schedules jobs using configuration from Consul.
func InitJob(app *app.App) error {
	jobrunner.Start()

	// Fetch initial schedules from Consul
	if err := scheduleJobsFromConsul(app); err != nil {
		app.Logger.Error("Failed to schedule jobs from Consul:", err)
		return err
	}

	// Watch for configuration changes in Consul and reschedule jobs
	go watchConsulForChanges(app)

	return nil
}

// scheduleJobsFromConsul schedules jobs based on configuration fetched from Consul.
func scheduleJobsFromConsul(app *app.App) error {
	// Iterate over job definitions from the configuration
	for _, jobConfig := range app.Cfg.AppConfig.Jobs.Job {
		// Fetch the job schedule from Consul with a default fallback value
		key := "service/" + app.AppName + "/config/jobs/" + jobConfig.Name
		interval, err := app.Service.ConsulKV.FetchConsulConfig(key, jobConfig.StartJob)
		if err != nil {
			app.Logger.Error("Failed to fetch Consul config for job:", jobConfig.Name, err)
			return err
		}

		// Schedule the job based on the function name specified in the configuration
		switch jobConfig.Name {
		case "s3Clean":
			job := NewCleanS3Job(app)
			if err = jobrunner.Schedule(interval, job); err != nil {
				app.Logger.Error("Failed to schedule job:", jobConfig.Name, err)
				return err
			}
			app.Logger.Info("Successfully scheduled job:", jobConfig.Name, "with interval:", interval)
		case "sessionClean":
			job := NewCleanOldSessionJob(app)
			if err = jobrunner.Schedule(interval, job); err != nil {
				app.Logger.Error("Failed to schedule job:", jobConfig.Name, err)
				return err
			}
			app.Logger.Info("Successfully scheduled job:", jobConfig.Name, "with interval:", interval)
		default:
			app.Logger.Warn("Unknown job function:", jobConfig.Name)
		}
	}

	return nil
}

// watchConsulForChanges monitors Consul for changes in job scheduling configuration and reschedules jobs.
func watchConsulForChanges(app *app.App) {
	for {
		// Re-schedule jobs if configuration in Consul changes
		if err := scheduleJobsFromConsul(app); err != nil {
			app.Logger.Error("Failed to reschedule jobs from Consul:", err)
		}

		// Poll interval before checking Consul again
		interval := time.Duration(app.Cfg.AppConfig.Jobs.IntervalRescanConsul) * time.Second
		time.Sleep(interval) // Adjust the interval as needed
	}
}

// NewCleanS3Job creates a new CleanS3Job instance.
func NewCleanS3Job(app *app.App) *CleanS3Job {
	return &CleanS3Job{
		app: app,
	}
}

type CleanS3Job struct {
	app *app.App
}

// NewCleanOldSessionJob creates a new CleanOldSessionJob instance.
func NewCleanOldSessionJob(app *app.App) *CleanOldSessionJob {
	return &CleanOldSessionJob{
		app: app,
	}
}

type CleanOldSessionJob struct {
	app *app.App
}
