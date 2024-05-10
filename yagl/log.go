// Package yagl provides Yet-Another-Go-Logger interface, because there isn't enough of them in Go ecosystem already.
package yagl

import (
	"log"
)

type SimpleLogger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
	Panicf(format string, v ...interface{})
}

// Simplifies adapts the specified [log.Logger] as a SimpleLogger.
func Simplifies(logger *log.Logger) SimpleLogger {
	return doLogger{
		logger: logger,
	}
}

type doLogger struct {
	logger *log.Logger
}

func (l doLogger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

func (l doLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

func (l doLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}
