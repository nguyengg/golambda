package lambdafunctionurl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

func (c *baseContext[T]) RequestMethod() string {
	return c.request.RequestContext.HTTP.Method
}

func (c *baseContext[T]) RequestPath() string {
	return c.request.RequestContext.HTTP.Path
}

func (c *baseContext[T]) RequestTimestamp() time.Time {
	return time.UnixMilli(c.request.RequestContext.TimeEpoch)
}

func (c *baseContext[T]) RequestHeader(key string) string {
	return c.requestHeaders.Get(key)
}

func (c *baseContext[T]) HasQueryParam(key string) bool {
	return c.requestQueryValues.Has(key)
}

func (c *baseContext[T]) QueryParam(key string) string {
	return c.requestQueryValues.Get(key)
}

func (c *baseContext[T]) QueryParamValues(key string) []string {
	return c.requestQueryValues[key]
}

func (c *baseContext[T]) QueryParamParseInt(key string, base, bitSize int) (int64, bool, error) {
	if t := c.requestQueryValues.Get(key); t != "" {
		v, err := strconv.ParseInt(t, base, bitSize)
		return v, true, err
	}

	return 0, false, nil
}

func (c *baseContext[T]) RequestCookie(key string) string {
	return c.requestCookies[key]
}

func (c *baseContext[T]) UnmarshalRequestBody(v interface{}) error {
	if !c.request.IsBase64Encoded {
		return json.Unmarshal([]byte(c.request.Body), v)
	}

	data, err := base64.StdEncoding.DecodeString(c.request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (c *baseContext[T]) UnmarshalRequestBodyWithOpts(v interface{}, opts ...func(decoder *json.Decoder)) error {
	if !c.request.IsBase64Encoded {
		dec := json.NewDecoder(strings.NewReader(c.request.Body))
		for _, opt := range opts {
			opt(dec)
		}
		return dec.Decode(&v)
	}

	data, err := base64.StdEncoding.DecodeString(c.request.Body)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	for _, opt := range opts {
		opt(dec)
	}
	return dec.Decode(&v)
}
