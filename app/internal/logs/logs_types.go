package logs

import (
	"fmt"
	"log/slog"
	"os"
)

// Grouped logging helper that adds a group with a name and logs the message.

func (l *Logger) Infof(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Info(msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Debug(msg)
}

func (l *Logger) Warnf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Warn(msg)
}

func (l *Logger) Errorf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Error(msg)
}

// Fatal logs a message at the error level and then exits.
func (l *Logger) Fatal(msg string, attrs ...slog.Attr) {
	// Convert slog.Attr slice to []any
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	l.Error(msg, args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...any) {
	// Format the message using fmt.Sprintf
	msg := fmt.Sprintf(format, args...)

	l.Error(msg)
	os.Exit(1)
}
