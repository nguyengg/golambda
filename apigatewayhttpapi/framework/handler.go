package framework

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	v2 "github.com/nguyengg/golambda/apigatewayhttpapi"
	"github.com/nguyengg/golambda/metrics"
	"net/http"
	"net/url"
)

// Context contains the original context from Lambda and methods to read attributes of the
// original events.APIGatewayV2HTTPRequest and write to the events.APIGatewayV2HTTPResponse.
type Context struct {
	ctx                context.Context
	request            *events.APIGatewayV2HTTPRequest
	requestHeader      http.Header
	requestQueryValues url.Values
	response           *events.APIGatewayV2HTTPResponse
	responseHeader     http.Header
}

// Start starts the Lambda runtime loop.
func Start(handler func(*Context) error) {
	v2.Start(func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		c := &Context{
			ctx:                ctx,
			request:            &req,
			requestHeader:      parseHeaders(req),
			requestQueryValues: parseQueryString(req),
			response:           &events.APIGatewayV2HTTPResponse{},
			responseHeader:     http.Header{},
		}
		err := handler(c)

		if len(c.response.Headers) == 0 {
			c.response.Headers = map[string]string{}
		}
		if len(c.response.MultiValueHeaders) == 0 {
			c.response.MultiValueHeaders = map[string][]string{}
		}

		for k, vs := range c.responseHeader {
			switch len(vs) {
			case 0:
				// how does this even happen...
			case 1:
				c.response.Headers[k] = vs[0]
			default:
				c.response.MultiValueHeaders[k] = vs
			}
		}

		return *c.response, err
	})
}

// Context returns the original context.Context of the request.
func (c *Context) Context() context.Context {
	return c.ctx
}

// WithValue replaces the underlying Context with the result from calling context.WithValue.
func (c *Context) WithValue(key, value any) context.Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	return c.ctx
}

// Value returns the value associated with the underlying Context that has been added with WithValue.
func (c *Context) Value(key any) any {
	return c.ctx.Value(key)
}

// Metrics returns the current metrics.Metrics instance from context.
func (c *Context) Metrics() metrics.Metrics {
	return metrics.Ctx(c.ctx)
}

// Request returns the original events.APIGatewayV2HTTPRequest request instance.
func (c *Context) Request() *events.APIGatewayV2HTTPRequest {
	return c.request
}

// Response returns the events.APIGatewayV2HTTPResponse instance that will be used to return contents back to caller.
func (c *Context) Response() *events.APIGatewayV2HTTPResponse {
	return c.response
}
