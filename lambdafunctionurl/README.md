# Lambda Function URL

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/golambda"
	"github.com/nguyengg/golambda/apigatewayhttpapi"
	"github.com/nguyengg/golambda/apigatewayhttpapi/auth"
	"github.com/nguyengg/golambda/apigatewayhttpapi/framework"
)

func main() {
	// without a context wrapper.
	apigatewayhttpapi.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		// will set LambdaContext.AwsRequestID to log prefix and reset upon completion.
		defer logsupport.SetUpGlobalLogger(ctx)()

		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       "hello, world!",
		}, nil
	})

	// with a context wrapper.
	framework.Start(func(c *framework.Context) error {
		return c.RespondOKWithText("hello, world!")
	})

	// authorizer example.
	auth.StartV2(func(ctx context.Context, request events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: true,
		}, nil
	})
}
```
