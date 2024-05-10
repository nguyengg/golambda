package yagl

import (
	"log"
)

// Discard ignores all [SimpleLogger.Printf] calls but still proxies [SimpleLogger.Fatalf] and [SimpleLogger.Panicf].
//
// If you want to also discard log writing for [SimpleLogger.Fatalf] and [SimpleLogger.Panicf] (their os.Exit and panic
// invocations still take effect), set [log.Logger.SetWriter] to [io.Discard] instead.
func Discard(logger *log.Logger) SimpleLogger {
	return &discardLogger{logger: logger}
}

type discardLogger struct {
	logger *log.Logger
}

func (l *discardLogger) Printf(string, ...interface{}) {
}

func (l *discardLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatalf(format, v...)
}

func (l *discardLogger) Panicf(format string, v ...interface{}) {
	l.logger.Panicf(format, v...)
}
