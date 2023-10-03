package auth

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/nguyengg/golambda/metrics"
	"github.com/nguyengg/golambda/start"
	"github.com/rs/zerolog"
	"log"
	"os"
)

// HandlerV2 for API Gateway HTTP Lambda authorizer requests using V2 payload request and response format.
type HandlerV2 func(context.Context, events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error)

// StartV2 the Lambda runtime loop.
func StartV2(handler HandlerV2, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.APIGatewayV2CustomAuthorizerV2Request) (response events.APIGatewayV2CustomAuthorizerSimpleResponse, err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			zerolog.New(os.Stderr).WithContext(ctx),
			request.RequestContext.RequestID,
			request.RequestContext.TimeEpoch)

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

				if response.IsAuthorized {
					m.SetCount("isAuthorized", 1).Log()
				} else {
					m.SetCount("isAuthorized", 2).Log()
				}
			}()
		}

		response, err = handler(ctx, request)
		panicked = false
		return
	})
}
