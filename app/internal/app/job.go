package app

import (
	"skeleton-golange-application/app/pkg/logging"

	"github.com/bamzi/jobrunner"
)

// jobrunner.Schedule("* */5 * * * *", DoSomething{}) // every 5min do something
// jobrunner.Schedule("@every 1h30m10s", ReminderEmails{})
// jobrunner.Schedule("@midnight", DataStats{}) // every midnight do this..
// jobrunner.Every(16*time.Minute, CleanS3{}) // evey 16 min clean...
// jobrunner.In(10*time.Second, WelcomeEmail{}) // one time job. starts after 10sec
// jobrunner.Now(NowDo{}) // do the job as soon as it's triggered
// https://github.com/robfig/cron/blob/v2/doc.go

func InitJob(logger *logging.Logger) error {
	jobrunner.Start()
	reminderEmails := NewReminderEmails(logger)
	err := jobrunner.Schedule("@every 5s", reminderEmails)
	if err != nil {
		logger.Error("Failed to schedule job:", err)
		return err
	}
	return nil
}

func NewReminderEmails(logger *logging.Logger) *ReminderEmails {
	return &ReminderEmails{
		logger: logger,
	}
}

// ReminderEmails Job Specific Functions.
type ReminderEmails struct {
	logger *logging.Logger
}

// Run ReminderEmails.Run() will get triggered automatically.
func (e ReminderEmails) Run() {
	// Queries the DB
	// Sends some email
	e.logger.Info("Every 5 seconds send reminder emails")
}
