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

// StatusCode returns the current response's status code .
func (c *Context) StatusCode() int {
	return c.response.StatusCode
}

// SetStatusCode changes the current response's status code.
func (c *Context) SetStatusCode(statusCode int) *Context {
	c.response.StatusCode = statusCode
	return c
}

// RespondOKWithJSON sets the response body to the JSON content of the argument v.
// If that succeeds, the response's status code is set to http.StatusOK, and the response's header "Content-Type" to
// "application/json".
// Use this method if you want to return a generic JSON result with 200 status code.
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

// RespondOKWithText sets the response body to plain-text.
// The response's status code is set to http.StatusOK, and the response's header "Content-Type" to "text/plain".
// Use this method if you want to return a generic plain-text result with 200 status code.
func (c *Context) RespondOKWithText(body string) error {
	c.response.Body = body
	c.response.StatusCode = http.StatusOK
	c.responseHeader.Set("Content-Type", "text/plain")
	return nil
}

// RespondOKWithBase64Data sets the response body to the base64 encoding of the given data.
// The response's status code is set to http.StatusOK, the response's header "Content-Type" is unchanged.
func (c *Context) RespondOKWithBase64Data(data []byte) error {
	c.response.Body = base64.StdEncoding.EncodeToString(data)
	c.response.StatusCode = http.StatusOK
	c.response.IsBase64Encoded = true
	return nil
}

// Respond sets the response's status code and a generated JSON response that contains the numeric "status" and a string
// "type" describing that status.
// Use this method if the status code is self-sufficient, and you don't need any additional message returned to caller.
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

// RespondMessage is a variant of Respond that allows an additional custom JSON response "message" attribute in addition
// to "status" and "type".
func (c *Context) RespondMessage(statusCode int, message string) error {
	return c.RespondFormatted(statusCode, "%s", message)
}

// RespondFormatted is a variant of RespondMessage that allows formatting of the custom JSON response "message".
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

// RespondInternalServerError is a variant of Respond for http.StatusBadRequest.
func (c *Context) RespondInternalServerError() error {
	return c.Respond(http.StatusInternalServerError)
}

// RespondBadRequest is a variant of RespondFormatted for http.StatusBadRequest.
func (c *Context) RespondBadRequest(layout string, v ...interface{}) error {
	return c.RespondFormatted(http.StatusBadRequest, layout, v...)
}

// RespondNotFound is a variant of Respond for http.StatusNotFound.
func (c *Context) RespondNotFound() error {
	return c.Respond(http.StatusNotFound)
}

// RespondMethodNotAllowed is a variant of Respond for http.StatusMethodNotAllowed with the Allow header value.
func (c *Context) RespondMethodNotAllowed(allow string) error {
	c.SetResponseHeader("Allow", allow)
	return c.Respond(http.StatusMethodNotAllowed)
}

// SetResponseHeader is used to modify a response header.
func (c *Context) SetResponseHeader(key, value string) *Context {
	c.responseHeader.Set(key, value)
	return c
}

// AddResponseHeader is used to add to a response header.
func (c *Context) AddResponseHeader(key, value string) *Context {
	c.responseHeader.Add(key, value)
	return c
}

// SetCacheControlMaxAge Sets the response Cache-Control header.
func (c *Context) SetCacheControlMaxAge(duration time.Duration) *Context {
	c.responseHeader.Set("Cache-Control", "max-age="+strconv.FormatInt(int64(duration/time.Second), 10))
	return c
}
