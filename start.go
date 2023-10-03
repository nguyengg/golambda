package golambda

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/nguyengg/golambda/metrics"
	"github.com/nguyengg/golambda/start"
	"log"
)

// StartHandlerFunc calls lambda.StartHandlerFunc passing the given handler after wrapping the context with a metrics
// instance that is used to populate basis statistics about the invocation.
//
// Use this wrapper if there isn't one created for specific events.
func StartHandlerFunc[TIn any, TOut any, H lambda.HandlerFunc[TIn, TOut]](handler H, options ...start.Option) {
	opts := start.New(options)

	lambda.StartHandlerFunc(func(ctx context.Context, in TIn) (out TOut, err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			"",
			0)

		if !opts.DisableSetUpGlobalLogger {
			defer logsupport.SetUpGlobalLogger(ctx)()
		}

		if !opts.DisableRequestDebugLogging && configsupport.IsDebug() {
			data, err := json.Marshal(in)
			if err != nil {
				log.Printf("ERROR marshal request: %v\n", err)
			} else {
				log.Printf("INFO request: %s\n", data)
			}
		}

		if !opts.DisableResponseDebugLogging && configsupport.IsDebug() {
			defer func() {
				data, err := json.Marshal(out)
				if err != nil {
					log.Printf("ERROR marshal response: %v\n", err)
				} else {
					log.Printf("INFO response: %s\n", data)
				}
			}()
		}

		defer func() {
			switch r := recover(); {
			case r != nil:
				log.Printf("ERROR handler panicked with error: %#v", r)
				m.Panicked()
			case err != nil:
				log.Printf("ERROR handler failed with error: %#v", err)
				m.Faulted()
			}

			m.Log()
		}()

		return handler(ctx, in)
	}, opts.HandlerOptions...)
}
