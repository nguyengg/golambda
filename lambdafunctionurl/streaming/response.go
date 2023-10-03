package streaming

import (
	"bytes"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

// Response implements base.Response for events.LambdaFunctionURLStreamingResponse.
type Response struct {
	response *events.LambdaFunctionURLStreamingResponse
}

func Wrap(response *events.LambdaFunctionURLStreamingResponse) *Response {
	return &Response{response: response}
}

func (r *Response) StatusCode() int {
	return r.response.StatusCode
}

func (r *Response) SetStatusCode(statusCode int) {
	r.response.StatusCode = statusCode
}

func (r *Response) SetHeader(key, value string) {
	r.response.Headers[textproto.CanonicalMIMEHeaderKey(key)] = value
}

func (r *Response) SetCookie(c http.Cookie) error {
	if err := c.Valid(); err != nil {
		return err
	}

	for i, e := range r.response.Cookies {
		if strings.HasPrefix(e, c.Name+"=") {
			r.response.Cookies[i] = c.String()
			return nil
		}
	}

	r.response.Cookies = append(r.response.Cookies, c.String())
	return nil
}
func (r *Response) RespondJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err == nil {
		r.response.Body = bytes.NewReader(data)
	}

	return err
}

func (r *Response) RespondText(body string) error {
	r.response.Body = strings.NewReader(body)
	return nil
}

func (r *Response) RespondBase64Data(data []byte) error {
	r.response.Body = bytes.NewReader(data)
	return nil
}

func (r *Response) RespondBody(body io.Reader) error {
	r.response.Body = body
	return nil
}
