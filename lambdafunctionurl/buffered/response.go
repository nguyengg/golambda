package buffered

import (
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

// Response implements base.Response for events.LambdaFunctionURLResponse.
type Response struct {
	response *events.LambdaFunctionURLResponse
}

func Wrap(init *events.LambdaFunctionURLResponse) *Response {
	return &Response{response: init}
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

func (r *Response) RespondJSON(v interface{}) (int, error) {
	data, err := json.Marshal(v)
	if err == nil {
		r.response.Body = string(data)
	}

	return len(data), err
}

func (r *Response) RespondText(body string) error {
	r.response.Body = body
	return nil
}

func (r *Response) RespondBase64Data(data []byte) error {
	r.response.Body = base64.StdEncoding.EncodeToString(data)
	r.response.IsBase64Encoded = true
	return nil
}

func (r *Response) RespondBody(body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	return r.RespondBase64Data(data)
}

func (r *Response) Response() *events.LambdaFunctionURLResponse {
	return r.response
}
