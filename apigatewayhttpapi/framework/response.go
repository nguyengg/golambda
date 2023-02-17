package framework

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Returns the current status code.
func (c Context) StatusCode() int {
	return c.response.StatusCode
}

// Sets the current status code.
func (c Context) SetStatusCode(statusCode int) Context {
	c.response.StatusCode = statusCode
	return c
}

// Sets the response body to the marshalled content of the argument v.
// If that succeeds, the response's status code is set to http.StatusOK. The response's header "Content-Type" is also
// automatically set to "application/json".
func (c *Context) RespondOKWithJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.response.Body = string(data)
	c.response.StatusCode = http.StatusOK
	c.responseHeader.Set("Content-Type", "application/json")
	return nil
}

// Sets the response body to plain-text.
// The response's header "Content-Type" is also automatically set to "text/plain".
func (c *Context) RespondOKWithText(body string) error {
	c.response.Body = body
	c.response.StatusCode = http.StatusOK
	c.responseHeader.Set("Content-Type", "text/plain")
	return nil
}

// Sets the response body to the base64 encoding of the given data.
func (c *Context) RespondOKWithBase64Data(data []byte) error {
	c.response.Body = base64.StdEncoding.EncodeToString(data)
	c.response.StatusCode = http.StatusOK
	c.response.IsBase64Encoded = true
	return nil
}

// Responds with a JSON message describing the status code.
// Use this if the status code is self sufficient and you don't need any additional message returned to caller.
func (c *Context) Respond(statusCode int) error {
	t := http.StatusText(statusCode)
	if t == "" {
		t = strconv.FormatInt(int64(statusCode), 10)
	}

	e := struct {
		Status int    `json:"status"`
		Type   string `json:"type,omitempty"`
	}{
		Status: statusCode,
		Type:   t,
	}

	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("ERROR marshal error response body")
		return err
	}

	c.response.StatusCode = statusCode
	c.response.Body = string(data)
	return nil
}

// Responds with a plain JSON message.
// Use this if you need to return additional messaging to caller.
func (c *Context) RespondMessage(statusCode int, message string) error {
	return c.RespondFormatted(statusCode, "%s", message)
}

// Responds with a formatted JSON message.
// Use this if you need to return additional messaging to caller.
func (c *Context) RespondFormatted(statusCode int, layout string, v ...interface{}) error {
	t := http.StatusText(statusCode)
	if t == "" {
		t = strconv.FormatInt(int64(statusCode), 10)
	}

	var m string
	switch len(v) {
	case 0:
		m = layout
	default:
		m = fmt.Sprintf(layout, v...)
	}

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
		return err
	}

	c.response.StatusCode = statusCode
	c.response.Body = string(data)
	return nil
}

// Variant of Respond for http.StatusBadRequest.
func (c *Context) RespondInternalServerError() error {
	return c.Respond(http.StatusInternalServerError)
}

// Variant of RespondFormatted for http.StatusBadRequest.
func (c *Context) RespondBadRequest(layout string, v ...interface{}) error {
	return c.RespondFormatted(http.StatusBadRequest, layout, v...)
}

// Variant of Respond for http.StatusNotFound.
func (c *Context) RespondNotFound() error {
	return c.Respond(http.StatusNotFound)
}

// Variant of Respond for http.StatusMethodNotAllowed with the Allow header value.
func (c *Context) RespondMethodNotAllowed(allow string) error {
	c.SetResponseHeader("Allow", allow)
	return c.Respond(http.StatusMethodNotAllowed)
}

// Sets the response header.
func (c *Context) SetResponseHeader(key, value string) *Context {
	c.responseHeader.Set(key, value)
	return c
}

// Adds the response header.
func (c *Context) AddResponseHeader(key, value string) *Context {
	c.responseHeader.Add(key, value)
	return c
}

// Sets the response Cache-Control header.
func (c *Context) SetCacheControlMaxAge(duration time.Duration) *Context {
	c.responseHeader.Set("Cache-Control", "max-age="+strconv.FormatInt(int64(duration/time.Second), 10))
	return c
}
