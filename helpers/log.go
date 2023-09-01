package helpers

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func SetupLogger(debug bool, logFilePath string) {
	// Set log level based on debug flag
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Create the log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Info("Failed to open log file, using default stderr")
		return
	}

	// For debugging purposes, we can add a debug log here
	logrus.Debug("Successfully opened the log file.")

	// Set logrus output to both console and file
	mw := io.MultiWriter(os.Stderr, file)
	logrus.SetOutput(mw)

	// Debug log after setting the output
	logrus.Debug("Logrus output set to both console and log file.")

	// Set the format
	customFormatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	logrus.SetFormatter(customFormatter)

	// Debug log after setting the formatter
	logrus.Debug("Custom formatter set for the logger.")
}
