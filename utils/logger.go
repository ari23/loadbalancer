package utils

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func NewLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()

	switch strings.ToLower(logLevel) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
