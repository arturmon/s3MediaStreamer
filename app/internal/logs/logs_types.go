package logs

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"s3MediaStreamer/app/model"
)

// ToLogFields converts any texture into a LogField array taking into account masking.
// ToLogFields converts any struct into a LogField array, considering masking.
func (l *Logger) ToLogFields(v interface{}) *LoggerMessageConnect {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return nil // Return nil if v is not a structure
	}

	fields := make([]model.LogField, 0)

	// Iterate through all fields of the structure
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldValue := val.Field(i)

		// Retrieve the mask tag
		mask := field.Tag.Get("mask")

		// Add a field to the LogField array
		fields = append(fields, model.LogField{
			Key:   field.Name,
			Value: fieldValue.Interface(),
			Mask:  mask,
		})
	}

	// Pass fields to NewLoggerMessageConnect
	loggerMsg := NewLoggerMessageConnect(fields)

	return loggerMsg
}

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
