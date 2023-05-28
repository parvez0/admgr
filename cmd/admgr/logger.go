package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
)

// Logger provides an interface to convert
// logger to custom logger type, it will have
// all the basic functionalities of a logger
type Logger interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Debug(args ...interface{})
	Error(args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}

var logger *logrus.Logger

// InitializeLogger returns a logrus custom_logger object with prefilled options
func InitializeLogger(config *Config, writer io.Writer) Logger {
	if logger != nil {
		return logger
	}

	baseLogger := logrus.New()

	// set REQUESTS_LOGLEVEL for custom_logger level, defaults to info
	level, err := logrus.ParseLevel(config.Logger.Level)
	if err != nil {
		panic(fmt.Sprintf("failed to parse log level : %s", err.Error()))
	}

	// setting custom_logger format to string
	baseLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: config.Logger.FullTimestamp,
	})

	baseLogger.SetOutput(writer)

	// set to true for showing filename and line number from where custom_logger being called
	baseLogger.SetReportCaller(false)
	baseLogger.SetLevel(level)

	logger = baseLogger
	return logger
}
