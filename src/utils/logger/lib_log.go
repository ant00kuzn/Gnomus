package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func SetupLogger() error {
	// Removing latest log if exists
	if err := os.Remove("latest.log"); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Setting logrus formatter
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "01-02 15:04:05",
	})

	// Setting log level
	logrus.SetLevel(logrus.InfoLevel)

	// Creating a file to log
	file, err := os.OpenFile("latest.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Setting log output to file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(multiWriter)

	// Custom log levels colors
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "01-02 15:04:05",
	})

	return nil
}
