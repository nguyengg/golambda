package cloudwatchevent

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/nguyengg/golambda/metrics"
	"github.com/nguyengg/golambda/start"
	"log"
)

// Handler for CloudWatch events.
type Handler func(ctx context.Context, request events.CloudWatchEvent) error

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.CloudWatchEvent) (err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			"",
			0)

		if !opts.DisableSetUpGlobalLogger {
			defer logsupport.SetUpGlobalLogger(ctx)()
		}

		if !opts.DisableRequestDebugLogging && configsupport.IsDebug() {
			data, err := json.Marshal(request)
			if err != nil {
				log.Printf("ERROR marshal request: %v\n", err)
			} else {
				log.Printf("INFO request: %s\n", data)
			}
		}

		panicked := true

		if !opts.DisableMetricsLogging {
			m.
				SetProperty("cloudWatchEventId", request.ID).
				SetProperty("detailType", request.DetailType).
				SetProperty("source", request.Source).
				SetProperty("accountId", request.AccountID).
				SetProperty("region", request.Region).
				SetJSONProperty("resources", request.Resources)

			defer func() {
				if panicked {
					m.Panicked()
				}
				if err != nil {
					m.Faulted()
				}

				m.Log()
			}()
		}

		err = handler(ctx, request)
		panicked = false
		return
	})
}
