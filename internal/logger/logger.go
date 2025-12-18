package logger

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the global logger
func Init() *logrus.Logger {
	log = logrus.New()

	// Set log level
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Create logs directory if it doesn't exist
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "./logs/automation.log"
	}

	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err == nil {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	}

	return log
}

// Get returns the global logger instance
func Get() *logrus.Logger {
	if log == nil {
		return Init()
	}
	return log
}
