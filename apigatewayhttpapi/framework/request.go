package framework

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Method returns the HTTP method of the request.
func (c *Context) Method() string {
	return c.request.RequestContext.HTTP.Method
}

// Path returns the HTTP path of the request.
func (c *Context) Path() string {
	return c.request.RequestContext.HTTP.Path
}

// UnmarshalRequestBody returns the request body as JSON.
func (c *Context) UnmarshalRequestBody(v interface{}) error {
	if !c.request.IsBase64Encoded {
		return json.Unmarshal([]byte(c.request.Body), v)
	}

	data, err := base64.StdEncoding.DecodeString(c.request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// RequestTimestamp returns the TimeEpoch of events.APIGatewayV2HTTPRequestContext wrapped as time.Time.
// Check time.Time.IsZero in case the TimeEpoch is missing or 0.
func (c *Context) RequestTimestamp() time.Time {
	return time.UnixMilli(c.request.RequestContext.TimeEpoch)
}

// RequestHeader returns the request header for the specified key.
func (c *Context) RequestHeader(key string) string {
	return c.requestHeader.Get(key)
}

// QueryParam returns the query parameter value for the specified key. If there are multiple values for the same header,
// QueryParam will return the first. See QueryParamValues.
func (c *Context) QueryParam(key string) string {
	return c.requestQueryValues.Get(key)
}

// QueryParamValues returns the query parameter values for the specified key. Use this method if there are multiple
// values for the same header.
func (c *Context) QueryParamValues(key string) []string {
	return c.requestQueryValues[key]
}

// PathParam returns the path parameter value for the specified key.
func (c *Context) PathParam(key string) string {
	return c.request.PathParameters[key]
}

// StageVariable returns the stage variable for the specified key.
func (c *Context) StageVariable(key string) string {
	return c.request.StageVariables[key]
}

// Create parseHeaders from the events.APIGatewayV2HTTPRequest's parseHeaders.
func parseHeaders(request events.APIGatewayV2HTTPRequest) http.Header {
	header := http.Header{}
	for k, vs := range request.Headers {
		for _, v := range strings.Split(vs, ",") {
			header.Add(k, v)
		}
	}

	return header
}

// Create query string values from events.APIGatewayV2HTTPRequest's QueryStringParameters.
func parseQueryString(request events.APIGatewayV2HTTPRequest) url.Values {
	values := url.Values{}
	for k, vs := range request.QueryStringParameters {
		for _, v := range strings.Split(vs, ",") {
			values.Add(k, v)
		}
	}
	return values
}
