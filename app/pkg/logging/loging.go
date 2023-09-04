package logging

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

type loggerInitializer struct {
	instance Logger
	once     sync.Once
}

type Logger struct {
	*logrus.Entry
}

func (s *Logger) ExtraFields(fields map[string]interface{}) *Logger {
	return &Logger{s.WithFields(fields)}
}

func newLogger(level string) Logger {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Fatalln(err)
	}

	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}
	l.SetOutput(os.Stdout)
	l.SetLevel(logrusLevel)

	return Logger{logrus.NewEntry(l)}
}

func GetLogger(level string) Logger {
	var loggerInit loggerInitializer

	loggerInit.once.Do(func() {
		loggerInit.instance = newLogger(level)
	})

	return loggerInit.instance
}
