package apigatewayhttpapi

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

// Handler for API Gateway HTTP API requests using V2 payload request and response format.
type Handler func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.StartHandlerFunc(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (response events.APIGatewayV2HTTPResponse, err error) {
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
				SetProperty("method", request.RequestContext.HTTP.Method).
				SetProperty("stage", request.RequestContext.Stage).
				SetProperty("routeKey", request.RequestContext.RouteKey)
			if len(request.PathParameters) != 0 {
				m.SetJSONProperty("pathParameters", request.PathParameters)
			}
			if len(request.StageVariables) != 0 {
				m.SetJSONProperty("stageVariables", request.StageVariables)
			}

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
