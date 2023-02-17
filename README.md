# Golang Utilities for AWS Lambda
This module contains utility classes to implement AWS Lambda handlers in Golang.

# Getting Started
```go
// main.go
package main

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"nguyen.gg/golambda/apigatewayhttpapi/framework"
	"nguyen.gg/golambda/metrics"
)

type Handler struct {
	dynamodbSvc *dynamodb.Client
	s3Svc       *s3.Client
}

// Handles requests to API Gateway HTTP Integration.
func (h *Handler) handle(c *framework.Context) error {
	return c.RespondOKWithText("hello, world!")
}

func main() {
	cfg, err := config.LoadDefaultConfig(c)
	if err != nil {
		panic(err)
	}
	cfg.APIOptions = append(cfg.APIOptions, metrics.ClientSideMetricsAPIOption())

	h := &Handler{
		dynamodbSvc: dynamodb.NewFromConfig(cfg),
		s3Svc:       s3.NewFromConfig(cfg),
	}
	framework.Start(h.handle)
}
```
