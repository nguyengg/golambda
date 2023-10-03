package start

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/rs/zerolog"
	"os"
)

// Options provides a base struct for customisations to starting handler.
type Options struct {
	// LoggerProvider provides a new zerolog.Logger instance on every request.
	LoggerProvider func(ctx context.Context) *zerolog.Logger

	// DisableRequestDebugLogging disables logging the JSON-encoded request in DEBUG mode.
	DisableRequestDebugLogging bool
	// DisableResponseDebugLogging disables logging the JSON-encoded response in DEBUG mode.
	DisableResponseDebugLogging bool

	// DisableMetricsLogging dictates whether the handler will automatically log metrics about the request
	// and response using metrics.Metrics.
	//
	// The metrics are written as structured JSON that contains basic information such as start and end time, status
	// codes, whether the handler returns an error or panics, etc. You can turn this feature off if you are doing your
	// own service logging.
	DisableMetricsLogging bool

	// DisableSetUpGlobalLogger dictates whether logsupport.SetUpGlobalLogger is called on every request.
	// logsupport.SetUpGlobalLogger sets up log.Default with reasonable flags as well as adding the request Id as
	// prefix. You should generally leave this feature enabled if you do a lot of logging with the default log module.
	DisableSetUpGlobalLogger bool

	// DisableSetUpZeroLogGlobalLevel dictates whether logsupport.SetUpZeroLogGlobalLevel is called once on startup.
	// logsupport.SetUpZeroLogGlobalLevel changes the global log level according to available environment variables.
	// It's very useful to leave enabled if you emit lot of logs at Debug level and would like to use the DEBUG
	// environment variable to toggle writing Debug logs.
	DisableSetUpZeroLogGlobalLevel bool

	// HandlerOptions passes along additional Lambda-runtime-specific options. See lambda.StartWithOptions.
	HandlerOptions []lambda.Option
}

type Option func(*Options)

// New creates an initial Options and applies modifiers thereto.
func New(options []Option) *Options {
	opts := &Options{
		LoggerProvider: func(ctx context.Context) *zerolog.Logger {
			l := zerolog.New(os.Stderr)
			return &l
		},
	}

	for _, opt := range options {
		opt(opts)
	}

	if !opts.DisableSetUpZeroLogGlobalLevel {
		logsupport.SetUpZeroLogGlobalLevel()
	}

	return opts
}

// DisableRequestDebugLogging disables logging the JSON-encoded request in DEBUG (configsupport.IsDebug) mode.
func DisableRequestDebugLogging() Option {
	return func(options *Options) {
		options.DisableRequestDebugLogging = true
	}
}

// DisableResponseDebugLogging disables logging the JSON-encoded response in DEBUG (configsupport.IsDebug) mode.
func DisableResponseDebugLogging() Option {
	return func(options *Options) {
		options.DisableResponseDebugLogging = true
	}
}

// DisableMetricsLogging disables logging with metrics.Metrics.
func DisableMetricsLogging() Option {
	return func(options *Options) {
		options.DisableMetricsLogging = true
	}
}

// DisableSetUpGlobalLoggerPerRequest disables calling logsupport.SetUpGlobalLogger on every request.
func DisableSetUpGlobalLoggerPerRequest() Option {
	return func(o *Options) {
		o.DisableSetUpGlobalLogger = true
	}
}

// DisableSetUpZeroLogGlobalLevel disables calling logsupport.SetUpZeroLogGlobalLevel once on startup.
func DisableSetUpZeroLogGlobalLevel() Option {
	return func(o *Options) {
		o.DisableSetUpZeroLogGlobalLevel = true
	}
}

// WithLoggerProvider allows customisation of the logger and its context on every request.
func WithLoggerProvider(loggerProvider func(ctx context.Context) *zerolog.Logger) Option {
	return func(o *Options) {
		o.LoggerProvider = loggerProvider
	}
}

// WithHandlerOptions allows additional options to be passed into the underlying Lambda runtime.
// See lambda.StartWithOptions.
func WithHandlerOptions(options ...lambda.Option) Option {
	return func(o *Options) {
		o.HandlerOptions = options
	}
}
