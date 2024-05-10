package yagl

import (
	"golang.org/x/time/rate"
	"log"
)

// Somtimes adapts the given logger's [log.Printf] to provide rate-limiting logging.
//
// Only SimpleLogger.Printf is subject to rate limiting.
func Somtimes(r *rate.Sometimes, logger *log.Logger) SimpleLogger {
	return &rateLogger{sometimes: rate.Sometimes{
		First:    r.First,
		Every:    r.Every,
		Interval: r.Interval,
	}, logger: logger}
}

type rateLogger struct {
	sometimes rate.Sometimes
	logger    *log.Logger
}

func (l *rateLogger) Printf(format string, v ...interface{}) {
	l.sometimes.Do(func() {
		l.logger.Printf(format, v...)
	})
}

func (l *rateLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

func (l *rateLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}
