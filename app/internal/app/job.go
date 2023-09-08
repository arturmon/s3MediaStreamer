package app

import (
	"skeleton-golange-application/app/pkg/logging"

	"github.com/bamzi/jobrunner"
)

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
