# Go goodies for AWS Lambda
The main features of this module are the various wrappers around different AWS Lambda events, for example:
* [Lambda Function URL](https://pkg.go.dev/github.com/nguyengg/golambda/lambdafunctionurl)
* [API Gateway HTTP Integration](https://pkg.go.dev/github.com/nguyengg/golambda/apigatewayhttpapi)
* [DynamoDB Stream](https://pkg.go.dev/github.com/nguyengg/golambda/dynamodbevent)

Other utilities:
* [DynamoDB](https://pkg.go.dev/github.com/nguyengg/golambda/ddb) needs such as modeling version attribute, timestamps, utilities around making GetItem,
PutItem, UpdateItem, and DeleteItem requests.
* [Metrics](https://pkg.go.dev/github.com/nguyengg/golambda/metrics) measures arbitrary counters, timings, properties, and produce a JSON message describing
about those metrics.
* [Log and/or parse](https://pkg.go.dev/github.com/nguyengg/golambda/logerror) Smithy errors.

The module is very opinionated about how things are done because they work for me, but I'm always looking for feedback
and suggestions.

# Getting Started
The root module exposes a generic wrapper that attaches a metrics instance to the context:

```go
package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/golambda"
	"github.com/nguyengg/golambda/logsupport"
)

func main() {
	golambda.StartHandlerFunc(func(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		// will set LambdaContext.AwsRequestID to log prefix and reset upon completion.
		defer logsupport.SetUpGlobalLogger(ctx)()

		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Body:       "hello, world!",
		}, nil
	})
}
```
