package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// LoggerOption describes the functions necessary to modify a logger at creation
type LoggerOption interface {
	apply(*Logger)
}

// WithInfoLog allows the user to set the log to which information messages will be sent
type WithInfoLog log.Logger

// Apply the info log to the logger
func (wl WithInfoLog) apply(logger *Logger) {
	logger.infoLog = (*log.Logger)(&wl)
}

// WithErrorLog allows the user to set the log to which error messages will be sent
type WithErrorLog log.Logger

// Apply the error log to the logger
func (wl WithErrorLog) apply(logger *Logger) {
	logger.errLog = (*log.Logger)(&wl)
}

// WithErrorProvider allows the user to set the error provider which defines the
// behavior the logger should exhibit when an error message is generated
type WithErrorProvider ErrorProvider

// Apply the error provider to the logger
func (ep WithErrorProvider) apply(logger *Logger) {
	logger.errProvider = ErrorProvider(ep)
}

// WithPrefix allows the user to set the prefix prepended to all log messages
type WithPrefix string

// Apply the prefix to the logger
func (p WithPrefix) apply(logger *Logger) {
	logger.Prefix = string(p)
}

// Logger contains the data necessary to log status and error messages in a standard way
type Logger struct {
	Environment string
	Prefix      string
	infoLog     *log.Logger
	errLog      *log.Logger
	errProvider ErrorProvider
}

// NewLogger creates a new logger from the service name and environment name
func NewLogger(service string, environment string, opts ...LoggerOption) *Logger {

	// First, create a logger with base values for its fields
	logger := Logger{
		infoLog:     log.New(os.Stdout, "", log.LstdFlags),
		errLog:      log.New(os.Stderr, "", log.LstdFlags),
		errProvider: ErrorProvider{SkipFrames: 2, PackageBase: "goutils"},
		Environment: environment,
		Prefix:      fmt.Sprintf("[%s][%s] ", environment, service),
	}

	// Next, if we have any options then apply them now
	for _, opt := range opts {
		opt.apply(&logger)
	}

	// Finally, return our logger
	return &logger
}

// ChangeFrame creates a new logger from an existing logger with a different
// number of frames to skip, allowing for errors to be referenced from a different
// part of the call stack than the default logger
func (logger *Logger) ChangeFrame(skipFrames int) *Logger {
	return &Logger{
		infoLog:     logger.infoLog,
		errLog:      logger.errLog,
		Environment: logger.Environment,
		Prefix:      logger.Prefix,
		errProvider: ErrorProvider{
			SkipFrames:  skipFrames,
			PackageBase: logger.errProvider.PackageBase,
		},
	}
}

// Log a message to the standard output
func (logger *Logger) Log(message string, args ...interface{}) {
	msg := "[Info]" + logger.Prefix + fmt.Sprintf(message, args...)
	logger.infoLog.Println(msg)
}

// Generate and log an error from the inner error and message. The
// resulting error will be returned for use by the caller
func (logger *Logger) Error(inner error, message string, args ...interface{}) *GError {
	err := logger.errProvider.GenerateError(logger.Environment, inner, message, args...)
	msg := "[Error]" + logger.Prefix + err.Error()
	logger.errLog.Println(msg)
	return err
}

// Discard is primarily used for testing. It operates by setting the
// output from all the loggers to a discard stream so that no logging
// actually appears on the screen
func (logger *Logger) Discard() {
	logger.infoLog.SetOutput(ioutil.Discard)
	logger.errLog.SetOutput(ioutil.Discard)
}
