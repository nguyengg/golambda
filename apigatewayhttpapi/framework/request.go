package framework

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"net/url"
	"strings"
)

// Returns the HTTP method of the request.
func (c Context) Method() string {
	return c.request.RequestContext.HTTP.Method
}

// Returns the HTTP path of the request.
func (c Context) Path() string {
	return c.request.RequestContext.HTTP.Path
}

// Unmarshalls the request body as JSON.
func (c Context) UnmarshalRequestBody(v interface{}) error {
	if !c.request.IsBase64Encoded {
		return json.Unmarshal([]byte(c.request.Body), v)
	}

	data, err := base64.StdEncoding.DecodeString(c.request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Returns the request header.
func (c Context) RequestHeader(key string) string {
	return c.requestHeader.Get(key)
}

// Returns the query parameter value.
func (c Context) QueryParam(key string) string {
	return c.requestQueryValues.Get(key)
}

// Returns the query parameter values.
func (c Context) QueryParamValues(key string) []string {
	return c.requestQueryValues[key]
}

// Returns the path parameter value.
func (c Context) PathParam(key string) string {
	return c.request.PathParameters[key]
}

// Returns the stage variable.
func (c Context) StageVariable(key string) string {
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
