package logs

import (
	"fmt"
	"log/slog"
	"os"
)

func (l *Logger) Infof(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Info(msg) // Log the formatted message (make sure you have an Info method)
}

func (l *Logger) Debugf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Debug(msg) // Log the formatted message (make sure you have a Debug method)
}

func (l *Logger) Warnf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Warn(msg) // Log the formatted message (make sure you have a Warn method)
}

func (l *Logger) Errorf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Error(msg) // Log the formatted message
}

// Fatal logs a message at the error level and then exits.
func (l *Logger) Fatal(msg string, attrs ...slog.Attr) {
	// Convert slog.Attr slice to []any
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	l.Error(msg, args...) // Log error
	os.Exit(1)            // Exit the program
}

func (l *Logger) Fatalf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Error(msg) // Log error
	os.Exit(1)   // Exit the program
}
