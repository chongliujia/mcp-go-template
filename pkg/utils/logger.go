package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is a global logger instance
var Logger *logrus.Logger

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// init initializes the global logger
func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
	})
	Logger.SetLevel(logrus.InfoLevel)
}

// SetLogLevel sets the logging level
func SetLogLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// SetFormatter sets the log formatter
func SetFormatter(formatter logrus.Formatter) {
	Logger.SetFormatter(formatter)
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	return Logger
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Logger.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Logger.Fatalf(format, args...)
}

// WithField creates a logger with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithFields creates a logger with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}