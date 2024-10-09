package jobs

import (
	"fmt"
	"s3MediaStreamer/app/internal/app"
	"time"

	"github.com/bamzi/jobrunner"
	"github.com/robfig/cron/v3"
)

// JobScheduler encapsulates job scheduling and management.
type JobScheduler struct {
	jobEntryMap  map[string]cron.EntryID
	jobConfigMap map[string]string
	keyPrefix    string
}

// NewJobScheduler creates a new JobScheduler instance.
func NewJobScheduler(appName string) *JobScheduler {
	return &JobScheduler{
		jobEntryMap:  make(map[string]cron.EntryID),
		jobConfigMap: make(map[string]string),
		keyPrefix:    fmt.Sprintf("service/%s/config/jobs/", appName),
	}
}

// InitJob initializes and schedules jobs using configuration from Consul.
func InitJob(app *app.App) error {
	jobScheduler := NewJobScheduler(app.AppName)
	jobrunner.Start()

	// Fetch initial schedules from Consul
	if err := jobScheduler.scheduleJobsFromConsul(app); err != nil {
		app.Logger.Error("Failed to schedule jobs from Consul:", err)
		return err
	}

	// Watch for configuration changes in Consul and reschedule jobs
	go jobScheduler.watchConsulForChanges(app)

	return nil
}

// scheduleJobsFromConsul schedules jobs based on configuration fetched from Consul.
func (js *JobScheduler) scheduleJobsFromConsul(app *app.App) error {
	// Iterate over job definitions from the configuration
	for _, jobConfig := range app.Cfg.AppConfig.Jobs.Job {
		// Fetch the job schedule from Consul with a default fallback value
		key := js.keyPrefix + jobConfig.Name
		interval, err := app.Service.ConsulKV.FetchConsulConfig(key, jobConfig.StartJob)
		if err != nil {
			app.Logger.Error("Failed to fetch Consul config for job:", jobConfig.Name, err)
			return err
		}

		// Check if the job configuration has changed
		lastConfig, configExists := js.jobConfigMap[jobConfig.Name]
		if !configExists || lastConfig != interval {
			// Log the configuration change
			app.Logger.Infof("Job configuration changed: %s from: %s to: %s", jobConfig.Name, lastConfig, interval)

			// Update the stored configuration
			js.jobConfigMap[jobConfig.Name] = interval
		} else {
			// Skip scheduling if the configuration hasn't changed
			app.Logger.Debugf("No change in job configuration for: %s", jobConfig.Name)
			continue
		}

		// Stop existing job if it exists
		if entryID, entryExists := js.jobEntryMap[jobConfig.Name]; entryExists {
			jobrunner.Remove(entryID)
			app.Logger.Infof("Removed existing job: %s", jobConfig.Name)
		}

		// Schedule the job based on the function name specified in the configuration
		var entryID cron.EntryID
		switch jobConfig.Name {
		case "s3Clean":
			job := NewCleanS3Job(app)
			err = jobrunner.Schedule(interval, job)
		case "sessionClean":
			job := NewCleanOldSessionJob(app)
			err = jobrunner.Schedule(interval, job)
		case "createNewMusicChart":
			job := NewCreateNewMusicChartJob(app)
			err = jobrunner.Schedule(interval, job)
		default:
			app.Logger.Warnf("Unknown job function: %s", jobConfig.Name)
			continue
		}

		if err != nil {
			app.Logger.Error("Failed to schedule job:", jobConfig.Name, err)
			return err
		}

		// Store the new EntryID in the map
		js.jobEntryMap[jobConfig.Name] = entryID
		app.Logger.Infof("Successfully scheduled job: %s with interval: %s", jobConfig.Name, interval)
	}

	return nil
}

// watchConsulForChanges monitors Consul for changes in job scheduling configuration and reschedules jobs.
func (js *JobScheduler) watchConsulForChanges(app *app.App) {
	for {
		// Re-schedule jobs if configuration in Consul changes
		if err := js.scheduleJobsFromConsul(app); err != nil {
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

// NewCreateNewMusicChartJob creates a new CreateNewMusicChartJob instance.
func NewCreateNewMusicChartJob(app *app.App) *CreateNewMusicChartJob {
	return &CreateNewMusicChartJob{
		app: app,
	}
}

type CreateNewMusicChartJob struct {
	app *app.App
}
