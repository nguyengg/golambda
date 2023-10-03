package dynamodbevent

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

// Handler for DynamoDB events that doesn't return a response, and as a result cannot report batch item failure.
type Handler func(ctx context.Context, request events.DynamoDBEvent) error

// HandlerWithResponse for DynamoDB events and returns a response to report batch item failure.
type HandlerWithResponse func(ctx context.Context, request events.DynamoDBEvent) (events.DynamoDBEventResponse, error)

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.DynamoDBEvent) (err error) {
		m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			"",
			0)
		ctx = m.WithContext(ctx)

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
			m.AddCount("recordCount", int64(len(request.Records)))

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

// StartHandlerWithResponse starts the Lambda runtime loop with the specified HandlerWithResponse.
func StartHandlerWithResponse(handler HandlerWithResponse, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.DynamoDBEvent) (response events.DynamoDBEventResponse, err error) {
		m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			"",
			0)
		ctx = m.WithContext(ctx)

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
			m.AddCount("recordCount", int64(len(request.Records)))
			m.AddCount("batchItemFailureCount", int64(len(response.BatchItemFailures)))

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

		response, err = handler(ctx, request)
		panicked = false
		return
	})
}
