package configsupport

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go/logging"
	"github.com/rs/zerolog"
	"os"
)

// Default adds a context-aware zerolog logger (ContextZerologAdapter) as well as turn on client logging only if
// IsDebug is true.
//
// Use this if you want a sensible setup for your AWS client.
//
// Usage
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), configsupport.Default))
func Default(o *config.LoadOptions) error {
	o.Logger = ContextZerologAdapter()

	switch os.Getenv("DEBUG") {
	case "1", "true":
		clientLogMode := aws.LogRequestWithBody | aws.LogResponseWithBody
		o.ClientLogMode = &clientLogMode
	}

	return nil
}

// ContextZerologAdapter returns a logging.Logger that implements logging.ContextLogger.
//
// Use this if you are attaching a zerolog.Logger to every context passed into the AWS clients. The logger will be
// retrieved with zerolog.Ctx.
//
// Usage
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithLogger(configsupport.ContextZerologAdapter()))
//
// See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/logging/.
func ContextZerologAdapter() logging.Logger {
	return &ctxAware{ctx: context.TODO()}
}

// StaticZerologAdapter wraps a zerolog.Logger and returns a logging.Logger that does not implement logging.ContextLogger.
//
// Use this if you aren't attaching a zerolog.Logger to every context passed into the AWS clients. Because zerolog.Ctx
// never returns nil (it may return a disabled logger), having a separate variant vs. one that is context-aware
// (ContextZerologAdapter) can be useful.
//
// Usage
//
//	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithLogger(configsupport.NewContextlessZeroLogAdapter()))
//
// See https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/logging/.
func StaticZerologAdapter(logger *zerolog.Logger) logging.Logger {
	return ctxLess{logger: logger}
}

type ctxLess struct {
	logger *zerolog.Logger
}

var _ logging.Logger = ctxLess{}
var _ logging.Logger = (*ctxLess)(nil)

func (c ctxLess) Logf(classification logging.Classification, format string, v ...interface{}) {
	switch classification {
	case logging.Warn:
		c.logger.Warn().Msgf(format, v...)
	case logging.Debug:
		fallthrough
	default:
		c.logger.Debug().Msgf(format, v...)
	}
}

type ctxAware struct {
	ctx context.Context
}

var _ logging.Logger = &ctxAware{}
var _ logging.Logger = (*ctxAware)(nil)

func (c *ctxAware) Logf(classification logging.Classification, format string, v ...interface{}) {
	logger := zerolog.Ctx(c.ctx)

	switch classification {
	case logging.Warn:
		logger.Warn().Msgf(format, v...)
	case logging.Debug:
		fallthrough
	default:
		logger.Debug().Msgf(format, v...)
	}
}

var _ logging.ContextLogger = &ctxAware{}
var _ logging.ContextLogger = (*ctxAware)(nil)

func (c *ctxAware) WithContext(ctx context.Context) logging.Logger {
	c.ctx = ctx
	return c
}
