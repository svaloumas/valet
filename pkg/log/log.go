package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

func NewLogger(key, format string) *logrus.Logger {

	logger := logrus.New()
	if format == "text" {
		customFormatter := &logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		}
		customFormatter.TimestampFormat = time.RFC3339Nano
		logger.SetFormatter(customFormatter)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{DataKey: key, TimestampFormat: time.RFC3339Nano})
	}

	return logger
}
