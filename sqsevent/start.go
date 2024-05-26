package sqsevent

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

// Handler is the handler for SQS events that can report individual record processing failure.
type Handler func(context.Context, events.SQSEvent) (events.SQSEventResponse, error)

// MessageHandler is the handler for individual message handlers. See StartMessageHandler.
type MessageHandler func(context.Context, events.SQSMessage) error

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.SQSEvent) (response events.SQSEventResponse, err error) {
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

		if !opts.DisableResponseDebugLogging && configsupport.IsDebug() {
			defer func() {
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("ERROR marshal response: %v\n", err)
				} else {
					log.Printf("INFO response: %s\n", data)
				}
			}()
		}

		panicked := true

		if !opts.DisableMetricsLogging {
			m.AddCount("recordCount", int64(len(request.Records)))
			m.AddCount("failureCount", int64(len(response.BatchItemFailures)))

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

		response, err = handler(m.WithContext(ctx), request)
		panicked = false
		return
	})
}

// StartMessageHandler handles the generation of events.SQSEventResponse for caller.
//
// When MessageHandler returns a non-nil error for a specific message, an events.SQSBatchItemFailure will be created for
// it. The main handler will always return a non-nil error unless panic happens.
func StartMessageHandler(handler MessageHandler, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.SQSEvent) (response events.SQSEventResponse, err error) {
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

		if !opts.DisableResponseDebugLogging && configsupport.IsDebug() {
			defer func() {
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("ERROR marshal response: %v\n", err)
				} else {
					log.Printf("INFO response: %s\n", data)
				}
			}()
		}

		panicked := true

		if !opts.DisableMetricsLogging {
			m.AddCount("recordCount", int64(len(request.Records)))
			m.AddCount("failureCount", int64(len(response.BatchItemFailures)))

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

		ctx = m.WithContext(ctx)

		for _, record := range request.Records {
			if err := handler(ctx, record); err != nil {
				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{ItemIdentifier: record.MessageId})
			}
		}

		panicked = false
		return
	})
}
