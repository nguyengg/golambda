package lambdafunctionurl

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nguyengg/golambda/lambdafunctionurl/cachecontrol"
	"github.com/nguyengg/golambda/lambdafunctionurl/etag"
	"github.com/nguyengg/golambda/metrics"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Context is the context passed into the wrapped handler of Lambda Function URL requests.
type Context interface {
	// Context returns the original context.Context of the request.
	Context() context.Context
	// WithValue replaces the underlying baseContext with the result from calling context.WithValue.
	WithValue(key, value any) context.Context
	// Value returns the value associated with the underlying baseContext that has been added with WithValue.
	Value(key any) any
	// Metrics returns the current metrics.Metrics instance from context.
	Metrics() metrics.Metrics

	// Request returns the original events.LambdaFunctionURLRequest instance.
	Request() *events.LambdaFunctionURLRequest
	// RequestHeaders returns the http.Header headers parsed from the original events.LambdaFunctionURLRequest instance.
	RequestHeaders() http.Header
	// RequestQueryValues returns the url.Values query string  parsed from the original events.lambdaFunctionURLRequest
	// instance.
	RequestQueryValues() url.Values
	// RequestMethod provides a convenient method to retrieve the HTTP method of the request from its
	// events.LambdaFunctionURLRequestContext.
	RequestMethod() string
	// RequestPath provides a convenient method to retrieve the HTTP path of the request from its
	// events.LambdaFunctionURLRequestContext.
	RequestPath() string
	// RequestTimestamp returns the TimeEpoch of events.LambdaFunctionURLRequestContext wrapped as time.Time.
	//
	// Check [time.Time.IsZero] in case the TimeEpoch is missing or 0.
	RequestTimestamp() time.Time
	// RequestHeader returns the request header for the specified key.
	RequestHeader(key string) string
	// HasQueryParam returns true if the key is set. See url.Values.
	HasQueryParam(key string) bool
	// QueryParam returns the query parameter value for the specified key.
	//
	// If there are multiple values for the same key, QueryParam will return the first. Use QueryParamValues to retrieve
	// all of them.
	QueryParam(key string) string
	// QueryParamValues returns the query parameter values for the specified key.
	//
	// Use this method if there are multiple values for the same key.
	QueryParamValues(key string) []string
	// QueryParamParseInt parses a query parameter value as numeric using strconv.ParseInt, passing the base and bitSize
	// arguments. Returns the parsed numeric value, true, nil if successful.
	//
	// Otherwise, return the error from strconv.ParseInt.
	QueryParamParseInt(key string, base, bitSize int) (int64, bool, error)
	// RequestCookie returns cookie value from the request.
	RequestCookie(key string) string
	// UnmarshalRequestBody parses the request body as JSON.
	UnmarshalRequestBody(v interface{}) error
	// UnmarshalRequestBodyWithOpts parses the request body as JSON with additional options for the decoding process.
	//
	// DisallowUnknownFields is often used with this method.
	UnmarshalRequestBodyWithOpts(v interface{}, opts ...func(decoder *json.Decoder)) error

	// StatusCode returns the current response's status code.
	StatusCode() int
	// SetStatusCode changes the current response's status code.
	SetStatusCode(statusCode int)
	// SetResponseHeader can be used to modify any response header.
	SetResponseHeader(key, value string)
	// SetCookie adds the cookie to the response.
	//
	// It is safe to call the method multiple times on the same cookie's name.
	//
	// Be sure to read the documentation of http.Cookie, especially on how you need to fill out more than just name and
	// value for the [http.Cookie.String] to return a Set-Cookie response.
	// Returns the response of [http.Cookie.Valid]; if the cookie is not valid, it will not be added.
	SetCookie(cookie http.Cookie) error
	// SetCacheControl is a convenient method to modify the Cache-Control response header.
	SetCacheControl(directives ...cachecontrol.ResponseDirective)
	// SetResponseCachingHeaders can be used to modify the "ETag" and "Last-Modified" response headers.
	//
	// If the value doesn't implement either HasETag and/or HasLastModified, false will be return.
	SetResponseCachingHeaders(v interface{}) (set bool)
	// RespondOKWithJSON sets the response body to the JSON-encoded content of the argument v.
	//
	// Upon successfully setting the new response body, the status code is set to http.StatusOK, "Content-Type" header
	// to "application/json; charset=utf-8", and "Content-Length" header to the number of bytes of the JSON content. If
	// the value implements HasETag and/or HasLastModified, their value are added to the response headers as well.
	//
	// Use this method if you want to return a generic JSON result with 200 status code.
	// Use RespondWithJSON if you need to customise the response further (set status code, headers, etc.).
	RespondOKWithJSON(v interface{}) (err error)
	// RespondWithJSON is a variant of RespondOKWithJSON without further side effects.
	//
	// Use this method if you need to customise the response further (set status code, headers, etc.).
	// Use RespondOKWithJSON if you want some sensible settings to accompany the body.
	RespondWithJSON(v interface{}) (err error)
	// RespondOKWithText sets the response body to the specified value.
	//
	// Upon successfully setting the new response body, the status code is set to http.StatusOK, "Content-Type" header
	// to "text/plain; charset=utf-8", and "Content-Length" header to the length of the body which is the number of
	// bytes, not the number of runes.
	//
	// Use this method if you want to return a generic plain-text result with 200 status code.
	// Use RespondWithText if you need to customise the response further (set status code, headers, etc.).
	RespondOKWithText(body string) (err error)
	// RespondWithText is a variant of RespondOKWithText without further side effects.
	//
	// Use this method if you need to customise the response further (set status code, headers, etc.).
	// Use RespondOKWithText if you want some sensible settings to accompany the body.
	RespondWithText(body string) (err error)
	// RespondOKWithBase64Data sets the response body to the base64 encoding of the given data.
	//
	// Upon successfully setting the new response body, the status code is set to http.StatusOK. You must still manually
	// set "Content-Type" header.
	RespondOKWithBase64Data(data []byte) (err error)
	// RespondWithBase64Data is a variant of RespondOKWithBase64Data without effecting status code changes.
	RespondWithBase64Data(data []byte) (err error)
	// RespondOKWithBody sets the response body to the given [io.Reader].
	//
	// Useful if the handler is in STREAMING instead of BUFFERED mode. In BUFFERED mode, the reader will be read in full
	// and passed to RespondWithBase64Data.
	RespondOKWithBody(body io.Reader) (err error)
	// RespondWithBody is a variant of RespondOKWithBody without effecting status code changes.
	RespondWithBody(body io.Reader) (err error)
	// SetResponseFormatterContentType changes the content type of the response generated by RespondFormatted.
	SetResponseFormatterContentType(t ResponseFormatterContentType)
	// RespondFormatted generates a response with the specified status code and formatted message.
	//
	// Upon successfully setting the new response body, the status code is also changed accordingly, and header
	// "Content-Type" is set to match the type of response which is JSONResponse by default. The format can be changed with
	// [baseContext.SetResponseFormatterContentType].
	//
	// The JSON response's body looks like this:
	//
	//	{ "status": statusCode, "message": sprintf(layout, v...) }
	//
	// The plain text response's body is the message.
	//
	// If you don't need a custom message, use RespondFormattedStatus which will use http.StatusText as the message.
	RespondFormatted(statusCode int, layout string, v ...interface{}) error
	// RespondFormattedStatus is a variant of RespondFormatted that derives the message from the status code.
	//
	// Equivalent to:
	//
	//	c.RespondFormatted(statusCode, "%s", http.StatusText(statusCode))
	//
	// Use this if the status code is sufficient, and you don't need a customised message.
	RespondFormattedStatus(statusCode int) (err error)
	// RespondInternalServerError calls RespondFormattedStatus with http.StatusInternalServerError.
	RespondInternalServerError() error
	// RespondBadRequest calls RespondFormatted passing http.StatusBadRequest and the message..
	RespondBadRequest(layout string, v ...interface{}) error
	// RespondNotFound calls RespondFormattedStatus with http.StatusNotFound.
	RespondNotFound() error
	// RespondMethodNotAllowed calls RespondFormattedStatus with http.StatusMethodNotAllowed and upon success also sets the
	// "Allow" response header.
	RespondMethodNotAllowed(allow string) (err error)

	// ParseIfMatch parses the "If-Match" request header and returns the directives.
	//
	// If the request doesn't contain "If-Match" header, returns nil, nil.
	ParseIfMatch() (*etag.Directives, error)
	// ParseIfNoneMatch parses the "If-None-Match" request header and returns the directives.
	//
	// If the request doesn't contain "If-None-Match" header, returns nil, nil.
	ParseIfNoneMatch() (*etag.Directives, error)
	// ParseIfModifiedSince parses the "If-Modified-Since" request header and returns the time.Time.
	//
	// If the request doesn't contain "If-Modified-Since" header, returns zero-value time.Time, nil.
	ParseIfModifiedSince() (time.Time, error)
	// ParseIfUnmodifiedSince parses the "If-Unmodified-Since" request header and returns the directives.
	//
	// If the request doesn't contain "If-Unmodified-Since" header, returns zero-value time.Time, nil.
	ParseIfUnmodifiedSince() (time.Time, error)

	// ProxyS3 will call S3 with the appropriate GET or HEAD method and return the response as either plain text or
	// base64-encoded data.
	//
	// Only GET and HEAD are supported. If the method is not recognized, RespondMethodNotAllowed is used.
	ProxyS3(client *s3.Client, bucket, key string) error
	// ProxyS3WithRequestHeaders is a variant of ProxyS3 that is given an extra http.Header whose values will be passed
	// into the S3's respective requests if the action supports it.
	ProxyS3WithRequestHeaders(client *s3.Client, bucket, key string, header http.Header) error
}

// DisallowUnknownFields is to be used with UnmarshalRequestBodyWithOpts to disallow unknown fields in decoded JSON.
func DisallowUnknownFields(dec *json.Decoder) {
	dec.DisallowUnknownFields()
}

// ResponseFormatterContentType describes which format [Context.RespondFormatted] use which is JSONResponse by default.
type ResponseFormatterContentType int

const (
	JSONResponse ResponseFormatterContentType = iota
	TextResponse
)

// HasETag allows [Context.RespondOKWithJSON] to add "ETag" header to the response.
type HasETag interface {
	GetETag() etag.ETag
}

// HasLastModified allows [Context.RespondOKWithJSON] to add "Last-Modified" header to the response.
type HasLastModified interface {
	GetLastModified() time.Time
}
