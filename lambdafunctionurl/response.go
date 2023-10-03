package lambdafunctionurl

import (
	"fmt"
	"github.com/nguyengg/golambda/lambdafunctionurl/cachecontrol"
	"io"
	"net/http"
	"strconv"
)

// Response is a wrapper around a specific Lambda response type (events.LambdaFunctionURLResponse or
// events.LambdaFunctionURLStreamingResponse) with additional functionality.
type Response[T any] interface {
	// StatusCode returns the current response's status code.
	StatusCode() int
	// SetStatusCode changes the current response's status code.
	SetStatusCode(statusCode int)
	// SetHeader sets the key-value entry for a specific header.
	SetHeader(key, value string)
	// SetCookie adds the cookie to the response.
	//
	// It is safe to call the method multiple times on the same cookie's name.
	//
	// Be sure to read the documentation of http.Cookie, especially on how you need to fill out more than just name and
	// value for the [http.Cookie.String] to return a Set-Cookie response.
	// Returns the response of [http.Cookie.Valid]; if the cookie is not valid, it will not be added.
	SetCookie(http.Cookie) error

	// RespondJSON sets the response body to the JSON-encoded content of the argument v.
	RespondJSON(v interface{}) (int, error)
	// RespondText sets the response body to the specified value.
	RespondText(body string) error
	// RespondBase64Data creates a response containing base64-encoded data.
	RespondBase64Data(data []byte) error
	// RespondBody creates a response that accepts an io.Reader.
	RespondBody(body io.Reader) error
}

func (c *baseContext[T]) StatusCode() int {
	return c.response.StatusCode()
}

func (c *baseContext[T]) SetStatusCode(statusCode int) {
	c.response.SetStatusCode(statusCode)
}

func (c *baseContext[T]) SetResponseHeader(key, value string) {
	c.response.SetHeader(key, value)
}

func (c *baseContext[T]) SetCookie(cookie http.Cookie) error {
	return c.response.SetCookie(cookie)
}

func (c *baseContext[T]) SetCacheControl(directives ...cachecontrol.ResponseDirective) {
	c.SetResponseHeader("Cache-Control", cachecontrol.Join(directives...))
}

func (c *baseContext[T]) SetResponseCachingHeaders(v interface{}) (set bool) {
	switch i := v.(type) {
	case HasETag:
		c.SetResponseHeader("ETag", i.GetETag().String())
		set = true
	}

	switch i := v.(type) {
	case HasLastModified:
		c.SetResponseHeader("Last-Modified", i.GetLastModified().Format(http.TimeFormat))
		set = false
	}

	return
}

func (c *baseContext[T]) RespondOKWithJSON(v interface{}) error {
	n, err := c.response.RespondJSON(v)
	if err == nil {
		c.SetStatusCode(http.StatusOK)
		c.SetResponseHeader("Content-Type", "application/json; charset=utf-8")
		c.SetResponseHeader("Content-Length", strconv.FormatInt(int64(n), 10))

		switch i := v.(type) {
		case HasETag:
			c.SetResponseHeader("ETag", i.GetETag().String())
		}

		switch i := v.(type) {
		case HasLastModified:
			c.SetResponseHeader("Last-Modified", i.GetLastModified().Format(http.TimeFormat))
		}
	}

	return err
}

func (c *baseContext[T]) RespondWithJSON(v interface{}) (err error) {
	_, err = c.response.RespondJSON(v)
	return
}

func (c *baseContext[T]) RespondOKWithText(body string) (err error) {
	if err = c.response.RespondText(body); err == nil {
		c.SetStatusCode(http.StatusOK)
		c.SetResponseHeader("Content-Type", "text/plain; charset=utf-8")
		c.SetResponseHeader("Content-Length", strconv.FormatInt(int64(len(body)), 10))
	}

	return
}

func (c *baseContext[T]) RespondWithText(body string) (err error) {
	return c.response.RespondText(body)
}

func (c *baseContext[T]) RespondOKWithBase64Data(data []byte) (err error) {
	if err = c.response.RespondBase64Data(data); err == nil {
		c.SetStatusCode(http.StatusOK)
	}

	return
}

func (c *baseContext[T]) RespondWithBase64Data(data []byte) (err error) {
	return c.response.RespondBase64Data(data)
}

func (c *baseContext[T]) RespondOKWithBody(body io.Reader) (err error) {
	if err = c.response.RespondBody(body); err == nil {
		c.SetStatusCode(http.StatusOK)
	}

	return
}

func (c *baseContext[T]) RespondWithBody(body io.Reader) (err error) {
	return c.response.RespondBody(body)
}

func (c *baseContext[T]) RespondFormatted(statusCode int, layout string, v ...interface{}) (err error) {
	var m string
	switch len(v) {
	case 0:
		m = layout
	default:
		m = fmt.Sprintf(layout, v...)
	}

	if c.responseFormatterContentType == TextResponse {
		if err = c.response.RespondText(m); err == nil {
			c.SetStatusCode(statusCode)
			c.SetResponseHeader("Content-Type", "text/plain; charset=utf-8")
		}
		return
	}

	if _, err = c.response.RespondJSON(struct {
		Status  int    `json:"status"`
		Message string `json:"message,omitempty"`
	}{
		Status:  statusCode,
		Message: m,
	}); err == nil {
		c.SetStatusCode(statusCode)
		c.SetResponseHeader("Content-Type", "application/json; charset=utf-8")
	}
	return
}

func (c *baseContext[T]) RespondFormattedStatus(statusCode int) (err error) {
	return c.RespondFormatted(statusCode, "%s", http.StatusText(statusCode))
}

func (c *baseContext[T]) RespondInternalServerError() error {
	return c.RespondFormattedStatus(http.StatusInternalServerError)
}

func (c *baseContext[T]) RespondBadRequest(layout string, v ...interface{}) error {
	return c.RespondFormatted(http.StatusBadRequest, layout, v...)
}

func (c *baseContext[T]) RespondNotFound() error {
	return c.RespondFormattedStatus(http.StatusNotFound)
}

func (c *baseContext[T]) RespondMethodNotAllowed(allow string) (err error) {
	if err = c.RespondFormattedStatus(http.StatusMethodNotAllowed); err == nil {
		c.SetResponseHeader("Allow", allow)
	}

	return
}
