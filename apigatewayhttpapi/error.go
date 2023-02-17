package apigatewayhttpapi

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"net/http"
	"strconv"
)

func Error(statusCode int, opts ...Opt) events.APIGatewayV2HTTPResponse {
	m := http.StatusText(statusCode)
	if m == "" {
		m = strconv.FormatInt(int64(statusCode), 10)
	}
	return ErrorWithMessage(statusCode, m, opts...)
}

func JSONError(statusCode int, opts ...Opt) events.APIGatewayV2HTTPResponse {
	m := http.StatusText(statusCode)
	if m == "" {
		m = strconv.FormatInt(int64(statusCode), 10)
	}

	return JSONErrorWithMessage(statusCode, m, opts...)
}

func ErrorWithMessage(statusCode int, message string, opts ...Opt) events.APIGatewayV2HTTPResponse {
	res := Errorf(statusCode, "%s", message)

	for _, opt := range opts {
		opt(&res)
	}

	return res
}

func JSONErrorWithMessage(statusCode int, message string, opts ...Opt) events.APIGatewayV2HTTPResponse {
	res := JSONErrorf(statusCode, "%s", message)

	for _, opt := range opts {
		opt(&res)
	}

	return res
}

func Errorf(statusCode int, layout string, v ...interface{}) events.APIGatewayV2HTTPResponse {
	m := fmt.Sprintf(layout, v...)

	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "text/plain"},
		Body:       m,
	}
}

func JSONErrorf(statusCode int, layout string, v ...interface{}) events.APIGatewayV2HTTPResponse {
	t := http.StatusText(statusCode)
	if t == "" {
		t = strconv.FormatInt(int64(statusCode), 10)
	}

	m := fmt.Sprintf(layout, v...)
	e := struct {
		Status  int    `json:"status"`
		Type    string `json:"type,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Status:  statusCode,
		Type:    t,
		Message: m,
	}

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("ERROR marshal error response body")
		return Errorf(statusCode, layout, v...)
	}
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(data),
	}
}
