package dilogger

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the level of logging.
type LogLevel int

const (
	Info LogLevel = iota
	Debug
	Warn
	Error
)

var defaultLogLevel LogLevel

func init() {
	envLogLevel := strings.ToUpper(strings.TrimSpace(os.Getenv("GODI_LOG_LEVEL")))
	switch envLogLevel {
	case "INFO":
		defaultLogLevel = Info
	case "DEBUG":
		defaultLogLevel = Debug
	case "WARN":
		defaultLogLevel = Warn
	case "ERROR":
		defaultLogLevel = Error
	default:
		defaultLogLevel = Error
		if envLogLevel != "" {
			log.Printf("Unknown log level: %s, defaulting to ERROR", envLogLevel)
		}
	}
}

// LoggerFunc defines the signature for logging functions used in LoggerOptions.
type LoggerFunc func(string, ...interface{})

// LoggerOptions allows users to customize the logging behavior by providing their own logging functions and setting the log level.
type LoggerOptions struct {
	LogLevel LogLevel
	Info     LoggerFunc
	Warn     LoggerFunc
	Debug    LoggerFunc
	Error    LoggerFunc
}

// Logger allows for flexible logging implementations that can be swapped out as needed.
// It defines methods for different log levels: Info, Warn, Debug, and Error.
//
// This is useful for dependency injection in applications where logging behavior may vary.
type Logger interface {
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

type loggerImpl struct {
	options *LoggerOptions
}

// isLoggerFunc checks if the provided function matches the expected signature for logging functions.
func isLoggerFunc(fn interface{}) bool {
	if fn == nil {
		return false
	}
	_, ok := fn.(LoggerFunc)
	return ok
}

// NewLogger creates a new logger instance with the provided options.
//
// The builder function allows users to customize the logger by modifying the LoggerOptions struct. If no builder is provided, it defaults to using the standard log package with the default log level.
// The log level filtering works by comparing the log level of the message being logged with the log level set in the LoggerOptions.
// If the message's log level is higher than the configured log level, it will be ignored and not logged.
//
// For example, if the log level is set to Warn, then Info and Debug messages will be ignored, while Warn and Error messages will be logged.
// This allows users to control the verbosity of their logs and only log messages that are relevant to their current needs.
//
// If no options are provided, it defaults to using the standard log package.
// The builder function allows users to customize the logger by modifying the LoggerOptions struct.
//
// Example usage:
//
//	logger := NewLogger(func(o *LoggerOptions) {
//	    o.LogLevel = Info
//	    o.Info = func(format string, v ...interface{}) {
//	        fmt.Printf("[CUSTOM INFO] "+format+"\n", v...)
//	    }
//	})
func NewLogger(builder func(*LoggerOptions)) Logger {
	options := &LoggerOptions{
		LogLevel: defaultLogLevel,
	}
	if builder != nil {
		builder(options)
	}

	opts := buildLoggerOptions(options)

	return &loggerImpl{options: opts}
}

// buildLoggerOptions merges the provided LoggerOptions with the default options.
func buildLoggerOptions(options *LoggerOptions) *LoggerOptions {
	loggerOpts := &LoggerOptions{
		LogLevel: defaultLogLevel,
	}
	if options != nil {
		if options.LogLevel >= Info && options.LogLevel <= Error {
			loggerOpts.LogLevel = options.LogLevel
		}
		if isLoggerFunc(options.Info) {
			loggerOpts.Info = options.Info
		}
		if isLoggerFunc(options.Warn) {
			loggerOpts.Warn = options.Warn
		}
		if isLoggerFunc(options.Debug) {
			loggerOpts.Debug = options.Debug
		}
		if isLoggerFunc(options.Error) {
			loggerOpts.Error = options.Error
		}
	}
	return loggerOpts
}

func (l *loggerImpl) Infof(format string, v ...interface{}) {
	if l.options.LogLevel > Info {
		return
	}
	var fn LoggerFunc = defaultInfoLogger
	if l.options.Info != nil {
		fn = l.options.Info
	}
	fn(format, v...)
}

func (l *loggerImpl) Debugf(format string, v ...interface{}) {
	if l.options.LogLevel > Debug {
		return
	}

	var fn LoggerFunc = defaultDebugLogger
	if l.options.Debug != nil {
		fn = l.options.Debug
	}
	fn(format, v...)
}

func (l *loggerImpl) Warnf(format string, v ...interface{}) {
	if l.options.LogLevel > Warn {
		return
	}
	var fn LoggerFunc = defaultWarnLogger
	if l.options.Warn != nil {
		fn = l.options.Warn
	}
	fn(format, v...)
}

func (l *loggerImpl) Errorf(format string, v ...interface{}) {
	if l.options.LogLevel > Error {
		return
	}
	var fn LoggerFunc = defaultErrorLogger
	if l.options.Error != nil {
		fn = l.options.Error
	}
	fn(format, v...)
}

func defaultInfoLogger(format string, v ...interface{}) {
	log.Printf("[GO-DI:INFO] "+format+"\n", v...)
}
func defaultDebugLogger(format string, v ...interface{}) {
	log.Printf("[GO-DI:DEBUG] "+format+"\n", v...)
}
func defaultWarnLogger(format string, v ...interface{}) {
	log.Printf("[GO-DI:WARN] "+format+"\n", v...)
}
func defaultErrorLogger(format string, v ...interface{}) {
	log.Printf("[GO-DI:ERROR] "+format+"\n", v...)
}
