# Go goodies for AWS Lambda

[![Go Reference](https://pkg.go.dev/badge/github.com/nguyengg/golambda.svg)](https://pkg.go.dev/github.com/nguyengg/golambda)

The main features of this module are the various wrappers around different AWS Lambda events, for example:
* [Lambda Function URL](https://pkg.go.dev/github.com/nguyengg/golambda/lambdafunctionurl), supporting both BUFFERED and RESPONSE_STREAM modes.
* [API Gateway HTTP Integration](https://pkg.go.dev/github.com/nguyengg/golambda/apigatewayhttpapi) with 
custom [authoriser](https://pkg.go.dev/github.com/nguyengg/golambda/apigatewayhttpapi/auth) wrapper.
* [DynamoDB Stream](https://pkg.go.dev/github.com/nguyengg/golambda/dynamodbevent) and other events.

Other utilities:
* [Must](https://pkg.go.dev/github.com/nguyengg/golambda/must#section-readme) provides `Must`, `Must0`, `Must2`, etc. to
reduce typing out `if a, err := someFunction(); err != nil`.
* [DynamoDB](https://pkg.go.dev/github.com/nguyengg/golambda/ddb) needs such as modeling version attribute, timestamps, utilities around making GetItem,
PutItem, UpdateItem, and DeleteItem requests.
* [Metrics](https://pkg.go.dev/github.com/nguyengg/golambda/metrics) measures arbitrary counters, timings, properties, and produce a JSON message describing
about those metrics.
* [Parse](https://pkg.go.dev/github.com/nguyengg/golambda/smithyerrors) or [log](https://pkg.go.dev/github.com/nguyengg/golambda/logerror) Smithy errors.

The module is very opinionated about how things are done because they work for me, but I'm always looking for feedback
and suggestions.

# Getting Started
The root module exposes a generic wrapper that attaches a metrics instance to the context:
```shell
# Download build.py to make it easier to build and update Lambda functions.
curl --proto '=https' -fo build.py https://raw.githubusercontent.com/nguyengg/golambda/main/build.py
```

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
