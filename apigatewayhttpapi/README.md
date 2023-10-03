# API Gateway HTTP Integration
Modules around API Gateway have fallen behind in terms of bug fixes and new features in comparison to
[Lambda Function URL](../lambdafunctionurl/README.md) since I've mostly migrated completely away from API Gateway.

The only redeeming feature of API Gateway at the moment is the authorisers and their ability to cache based on a header
or query parameter identity source. Since I've moved to cookie-based authenication and authorisation, however, the point
is moot.

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
