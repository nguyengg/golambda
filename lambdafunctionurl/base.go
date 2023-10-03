package lambdafunctionurl

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/nguyengg/golambda/lambdafunctionurl/etag"
	"github.com/nguyengg/golambda/metrics"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// baseContext abstracts the request which must be of type events.LambdaFunctionURLRequest, and the response which ideally
// should wrap either events.LambdaFunctionURLResponse or events.LambdaFunctionURLStreamingResponse.
type baseContext[T any] struct {
	ctx                          context.Context
	request                      *events.LambdaFunctionURLRequest
	requestHeaders               http.Header
	requestQueryValues           url.Values
	requestCookies               map[string]string
	response                     Response[T]
	responseFormatterContentType ResponseFormatterContentType
}

func newContext[T any](ctx context.Context, request *events.LambdaFunctionURLRequest, response Response[T]) *baseContext[T] {
	return &baseContext[T]{
		ctx:                          ctx,
		request:                      request,
		requestHeaders:               parseHeaders(request),
		requestQueryValues:           parseQueryString(request),
		requestCookies:               parseCookies(request),
		response:                     response,
		responseFormatterContentType: JSONResponse,
	}
}

func (c *baseContext[T]) Context() context.Context {
	return c.ctx
}

func (c *baseContext[T]) WithValue(key, value any) context.Context {
	c.ctx = context.WithValue(c.ctx, key, value)
	return c.ctx
}

func (c *baseContext[T]) Value(key any) any {
	return c.ctx.Value(key)
}

func (c *baseContext[T]) Metrics() metrics.Metrics {
	return metrics.Ctx(c.ctx)
}

func (c *baseContext[T]) Request() *events.LambdaFunctionURLRequest {
	return c.request
}

func (c *baseContext[T]) RequestHeaders() http.Header {
	return c.requestHeaders
}

func (c *baseContext[T]) RequestQueryValues() url.Values {
	return c.requestQueryValues
}

func (c *baseContext[T]) SetResponseFormatterContentType(t ResponseFormatterContentType) {
	c.responseFormatterContentType = t
}

func parseHeaders(request *events.LambdaFunctionURLRequest) http.Header {
	header := http.Header{}
	for k, v := range request.Headers {
		header.Add(k, v)
	}

	return header
}

func parseQueryString(request *events.LambdaFunctionURLRequest) url.Values {
	values := url.Values{}
	for k, vs := range request.QueryStringParameters {
		for _, v := range strings.Split(vs, ",") {
			values.Add(k, v)
		}
	}
	return values
}

func parseCookies(request *events.LambdaFunctionURLRequest) map[string]string {
	cookies := make(map[string]string, len(request.Cookies))
	for _, cookie := range request.Cookies {
		values := strings.SplitN(cookie, "=", 2)
		if len(values) == 2 {
			cookies[values[0]] = values[1]
		}
	}

	return cookies
}

func (c *baseContext[T]) ParseIfMatch() (*etag.Directives, error) {
	return etag.ParseDirectives(c.RequestHeader("If-Match"))
}

func (c *baseContext[T]) ParseIfNoneMatch() (*etag.Directives, error) {
	return etag.ParseDirectives(c.RequestHeader("If-None-Match"))
}

func (c *baseContext[T]) ParseIfModifiedSince() (t time.Time, err error) {
	if v := c.RequestHeader("If-Modified-Since"); v != "" {
		t, err = time.Parse(http.TimeFormat, v)
		if err != nil {
			return t, fmt.Errorf("parse If-Modified-Since: %w", err)
		}
	}

	return
}

func (c *baseContext[T]) ParseIfUnmodifiedSince() (t time.Time, err error) {
	if v := c.RequestHeader("If-Unmodified-Since"); v != "" {
		t, err = time.Parse(http.TimeFormat, v)
		if err != nil {
			return t, fmt.Errorf("parse If-Unmodified-Since: %w", err)
		}
	}

	return
}
