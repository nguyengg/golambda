package lambdafunctionurl

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/nguyengg/golambda/lambdafunctionurl/buffered"
	"github.com/nguyengg/golambda/lambdafunctionurl/streaming"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/nguyengg/golambda/metrics"
	"github.com/nguyengg/golambda/start"
	"log"
)

// Handler handles requests to Lambda Function URLs in BUFFERED invoke mode.
type Handler func(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error)

// StreamingHandler for requests to Lambda Function URLs in RESPONSE_STREAM invoke mode.
type StreamingHandler func(ctx context.Context, request events.LambdaFunctionURLRequest) (*events.LambdaFunctionURLStreamingResponse, error)

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.StartHandlerFunc(func(ctx context.Context, request events.LambdaFunctionURLRequest) (response events.LambdaFunctionURLResponse, err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			request.RequestContext.RequestID,
			request.RequestContext.TimeEpoch)

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
			m.
				SetProperty("path", request.RequestContext.HTTP.Path).
				SetProperty("method", request.RequestContext.HTTP.Method)

			defer func() {
				if panicked {
					m.Panicked()
				}
				if err != nil {
					m.Faulted()
				}

				m.SetStatusCode(response.StatusCode).Log()
			}()
		}

		response, err = handler(ctx, request)
		panicked = false
		return
	}, opts.HandlerOptions...)
}

// StartStreaming starts the Lambda runtime loop with the specified StreamingHandler.
func StartStreaming(handler StreamingHandler, options ...start.Option) {
	opts := start.New(options)

	lambda.StartHandlerFunc(func(ctx context.Context, request events.LambdaFunctionURLRequest) (response *events.LambdaFunctionURLStreamingResponse, err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			request.RequestContext.RequestID,
			request.RequestContext.TimeEpoch)

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
			m.
				SetProperty("path", request.RequestContext.HTTP.Path).
				SetProperty("method", request.RequestContext.HTTP.Method)

			defer func() {
				if panicked {
					m.Panicked()
				}
				if err != nil {
					m.Faulted()
				}

				m.SetStatusCode(response.StatusCode).Log()
			}()
		}

		response, err = handler(ctx, request)
		panicked = false
		return
	}, opts.HandlerOptions...)
}

// StartWrapper starts the Lambda runtime loop with the abstract handler.
func StartWrapper(handler func(Context) error, options ...start.Option) {
	Start(func(ctx context.Context, req events.LambdaFunctionURLRequest) (response events.LambdaFunctionURLResponse, err error) {
		response = events.LambdaFunctionURLResponse{
			Headers: map[string]string{},
			Cookies: make([]string, 0),
		}
		c := newContext[events.LambdaFunctionURLResponse](ctx, &req, buffered.Wrap(&response))
		err = handler(c)
		return
	}, options...)
}

// StartStreamingWrapper starts the Lambda runtime loop with the abstract handler.
func StartStreamingWrapper(handler func(Context) error, options ...start.Option) {
	StartStreaming(func(ctx context.Context, req events.LambdaFunctionURLRequest) (response *events.LambdaFunctionURLStreamingResponse, err error) {
		response = &events.LambdaFunctionURLStreamingResponse{
			Headers: map[string]string{},
			Cookies: make([]string, 0),
		}
		c := newContext[events.LambdaFunctionURLStreamingResponse](ctx, &req, streaming.Wrap(response))
		err = handler(c)
		return
	}, options...)
}
