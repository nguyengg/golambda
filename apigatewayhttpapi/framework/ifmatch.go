package framework

import (
	"fmt"
	"regexp"
	"strings"
)

// IfMatch header value.
type IfMatch struct {
	ETags []ETag
	Any   bool
}

var weakETag = regexp.MustCompile(`^W/"(?P<value>.+)"$`)
var strongETag = regexp.MustCompile(`^"(?P<value>.+)"$`)

// ParseIfMatchHeader parses and returns the If-Match request header.
func (c *Context) ParseIfMatchHeader() (*IfMatch, error) {
	value := c.RequestHeader("If-Match")
	if value == "" {
		return nil, nil
	}

	if value == "*" {
		return &IfMatch{Any: true}, nil
	}

	values := strings.Split(value, ", ")
	if len(values) == 0 {
		return nil, fmt.Errorf("no ETag values")
	}

	etags := make([]ETag, 0)
	for _, v := range values {
		m := weakETag.FindAllStringSubmatch(v, -1)
		if len(m) == 1 {
			etags = append(etags, NewWeakETag(m[0][1]))
			continue
		}

		m = strongETag.FindAllStringSubmatch(v, -1)
		if len(m) == 1 {
			etags = append(etags, NewStrongETag(m[0][1]))
			continue
		}

		return nil, fmt.Errorf("invalid request ETag header")
	}

	return &IfMatch{ETags: etags}, nil
}
