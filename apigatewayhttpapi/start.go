package apigatewayhttpapi

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/golambda/metrics"
	"log"
	"net/http"
	"time"
)

// Handler for API Gateway HTTP API requests using V2 payload request and response format.
type Handler func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error)

// Start starts the Lambda runtime loop.
func Start(handler Handler) {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	lambda.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (response events.APIGatewayV2HTTPResponse, err error) {
		startTime := time.Now().UTC()

		lc, ok := lambdacontext.FromContext(ctx)
		if !ok {
			return JSONError(http.StatusInternalServerError), fmt.Errorf("no lambda context")
		}
		log.SetPrefix(lc.AwsRequestID + " ")

		var m *metrics.Metrics

		defer func() {
			switch r := recover(); {
			case r != nil:
				if e, ok := r.(error); ok {
					log.Printf("ERROR panicked with error: %v", e)
					err = e
					break
				}
				log.Printf("ERROR panicked due to: %v", r)
				err = fmt.Errorf("recover: %v", r)
				if m != nil {
					_ = m.Faulted()
					_ = m.Panicked()
				}
				response = JSONError(http.StatusInternalServerError)
			case err != nil:
				log.Printf("ERROR handler failed with error: %v", err)
				if m != nil {
					_ = m.Faulted()
				}
			}

			if m == nil {
				log.Printf("ERROR cannot emit metrics since it wasn't created properly")
				return
			}

			_ = m.SetProperty("statusCode", response.StatusCode)
			if len(response.Cookies) != 0 {
				_ = m.SetProperty("responseCookies", response.Cookies)
			}
			if len(response.Headers) != 0 {
				_ = m.SetProperty("responseHeaders", response.Headers)
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
