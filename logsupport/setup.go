package logsupport

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/rs/zerolog"
	"log"
	"os"
)

// RecommendedDebugLogFlag is the flag passed to log.SetFlags by SetUpLogger if [configsupport.IsDebug] is true.
const RecommendedDebugLogFlag = log.Ldate | log.Lmicroseconds | log.LUTC | log.Llongfile | log.Lmsgprefix

// RecommendedLogFlag is the flag passed to log.SetFlags by SetUpLogger if [configsupport.IsDebug] is false.
const RecommendedLogFlag = RecommendedDebugLogFlag | log.Lshortfile

// SetUpGlobalLogger sets up log.Default with flags set to RecommendedLogFlag and prefix set to the AwsRequestID from
// lambdacontext.FromContext.
//
// A function is returned that should be deferred upon to reset the log flags and prefix back to the original values.
// Use SetUpLogger if you wish to modify a specific log.Logger.
//
// Usage
//
//	// notice the double ()() to make sure SetUpGlobalLogger is run first, then its returned function is deferred.
//	defer logsupport.SetUpGlobalLogger())
func SetUpGlobalLogger(ctx context.Context) func() {
	return SetUpLogger(ctx, log.Default())
}

// SetUpLogger is a variant of SetUpGlobalLogger that targets a specific log.Logger.
func SetUpLogger(ctx context.Context, logger *log.Logger) func() {
	flags := logger.Flags()
	prefix := logger.Prefix()

	if configsupport.IsDebug() {
		logger.SetFlags(RecommendedDebugLogFlag)
	} else {
		logger.SetFlags(RecommendedLogFlag)
	}

	if lc, ok := lambdacontext.FromContext(ctx); ok {
		logger.SetPrefix(lc.AwsRequestID + " ")
	}

	return func() {
		logger.SetFlags(flags)
		logger.SetPrefix(prefix)
	}
}

// SetUpZeroLogGlobalLevel sets zerolog.SetGlobalLevel according to available environment variables.
//
// If ZEROLOG_GLOBAL_LEVEL is parsable with zerolog.ParseLevel then that value will be used.
// If configsupport.IsDebug is true then the level will be set to Debug.
// Otherwise, Info will be used as the default level.
func SetUpZeroLogGlobalLevel() {
	if level, err := zerolog.ParseLevel(os.Getenv("ZEROLOG_GLOBAL_LEVEL")); err == nil && level != zerolog.NoLevel {
		zerolog.SetGlobalLevel(level)
		return
	}

	if configsupport.IsDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
