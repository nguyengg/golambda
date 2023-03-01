package auth

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/golambda/metrics"
	"log"
	"time"
)

// Handler for API Gateway HTTP Lambda authorizer requests using V2 payload request and response format.
type HandlerV2 func(context.Context, APIGatewayHTTPLambdaAuthorizerV2Request) (APIGatewayHTTPLambdaAuthorizerV2SimpleResponse, error)

// Starts the Lambda runtime loop.
func StartV2(handler HandlerV2) {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	lambda.Start(func(ctx context.Context, request APIGatewayHTTPLambdaAuthorizerV2Request) (response APIGatewayHTTPLambdaAuthorizerV2SimpleResponse, err error) {
		startTime := time.Now().UTC()

		lc, ok := lambdacontext.FromContext(ctx)
		if !ok {
			return APIGatewayHTTPLambdaAuthorizerV2SimpleResponse{IsAuthorized: false}, fmt.Errorf("no lambda context")
		}
		log.SetPrefix(lc.AwsRequestID + " ")

		var m *metrics.Metrics

		defer func() {
			switch r := recover(); {
			case r != nil:
				if e, ok := r.(error); ok {
					log.Printf("panicked with error: %v", e)
					err = e
					break
				}
				log.Printf("panicked due to: %v", r)
				err = fmt.Errorf("recover: %v", r)
				if m != nil {
					_ = m.Faulted()
					_ = m.Panicked()
				}
				response = APIGatewayHTTPLambdaAuthorizerV2SimpleResponse{IsAuthorized: false}
			case err != nil:
				log.Printf("failed with error: %v", err)
				if m != nil {
					_ = m.Faulted()
				}
				response = APIGatewayHTTPLambdaAuthorizerV2SimpleResponse{IsAuthorized: false}
			}

			if m == nil {
				log.Printf("ERROR cannot emit metrics since it wasn't created properly")
				return
			}

			var isAuthorized int64
			if response.IsAuthorized {
				isAuthorized = 1
			}
			_ = m.AddCount("isAuthorized", isAuthorized)
			if len(response.Context) != 0 {
				_ = m.SetProperty("responseContext", response.Context)
			}

			fmt.Printf("%s\n", m.JSONString())
		}()

		m = metrics.NewWithStartTime(startTime)
		_ = m.SetProperty("lambdaRequestId", lc.AwsRequestID)
		_ = m.SetProperty("apiGatewayRequestId", request.RequestContext.RequestID)
		_ = m.SetProperty("path", request.RequestContext.HTTP.Path)
		_ = m.SetProperty("method", request.RequestContext.HTTP.Method)
		_ = m.SetProperty("stage", request.RequestContext.Stage)
		_ = m.SetProperty("routeKey", request.RequestContext.RouteKey)
		if len(request.PathParameters) != 0 {
			_ = m.SetProperty("pathParameters", request.PathParameters)
		}
		if len(request.StageVariables) != 0 {
			_ = m.SetProperty("stageVariables", request.StageVariables)
		}

		response, err = handler(metrics.NewContext(ctx, m), request)
		return
	})
}
