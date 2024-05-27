package logs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

type LogsRepositoryInterface interface {
	NewLogger(level string, format string, logrusWriter io.Writer) *Logger
	GetLogger(level, format, graylogAddr, graylogType, appName string) *Logger
}

const callerDepth = 5

type loggerInitializer struct {
	instance *Logger
	once     sync.Once
}

type Logger struct {
	*logrus.Entry
}

type gelfWriterWrapper struct {
	writer io.Writer
}

func (w *gelfWriterWrapper) Write(p []byte) (int, error) {
	// Parse the GELF message into a map
	var data map[string]interface{}
	err := json.Unmarshal(p, &data)
	if err != nil {
		return 0, err
	}

	// Get the runtime frame to retrieve filename, line number, and function name
	pc, file, line, ok := runtime.Caller(callerDepth) // Adjust the frame depth as needed
	if !ok {
		return 0, errors.New("failed to get runtime caller information")
	}

	// Add filename, line number, and function name to the GELF message
	data["_file"] = file
	data["_line"] = line
	data["_func"] = runtime.FuncForPC(pc).Name()

	// Marshal the modified data back to JSON
	modifiedP, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	// Write the modified GELF message to the underlying writer
	return w.writer.Write(modifiedP)
}

func GetLogger(level, format, graylogAddr, graylogType, appName string) *Logger {
	var loggerInit loggerInitializer

	loggerInit.once.Do(func() {
		var logrusWriter io.Writer

		if format == "gelf" {
			if graylogAddr != "" {
				var gelfWriter io.Writer
				var err error

				// If using UDP
				if graylogType == "udp" {
					gelfWriter, err = NewUDPWriter(graylogAddr, appName)
				}
				// If using TCP
				if graylogType == "tcp" {
					gelfWriter, err = NewTCPWriter(graylogAddr, appName)
				}

				if err != nil {
					log.Fatalf("gelf.NewWriter: %s", err)
				}

				// Customize the GELF message based on your needs
				// Directly use constructMessage function where needed

				// Create a new Logrus logger for the GELF writer
				gelfLogger := logrus.New()
				gelfLogger.SetFormatter(&logrus.JSONFormatter{}) // Use a JSON formatter for consistency
				gelfWrapper := &gelfWriterWrapper{writer: gelfWriter}
				gelfLogger.SetOutput(gelfWrapper)
				logrusLevel, err := logrus.ParseLevel(level)
				gelfLogger.SetLevel(logrusLevel)

				loggerInit.instance = &Logger{logrus.NewEntry(gelfLogger)}
			}
		}

		if format != "gelf" {
			// For other formats, use the default logger initialization logic
			loggerInit.instance = newLogger(level, format, logrusWriter)
		}
	})

	return loggerInit.instance
}

func newLogger(level string, format string, logrusWriter io.Writer) *Logger {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalln(err)
	}

	log := logrus.New()
	log.SetReportCaller(true)

	switch format {
	case "text":
		log.Formatter = &logrus.TextFormatter{}
	case "json":
		log.Formatter = &logrus.JSONFormatter{}
	default:
		log.Formatter = &logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
			},
			DisableColors: false,
			FullTimestamp: true,
		}
	}

	if logrusWriter != nil {
		log.SetOutput(logrusWriter)
	}
	log.SetLevel(logrusLevel)

	return &Logger{logrus.NewEntry(log)}
}
