package etag

import (
	"fmt"
	"regexp"
	"strings"
)

// ETag header value.
type ETag struct {
	Value string
	Weak  bool
}

// NewStrongETag returns an ETag with ETag.Weak set to false.
func NewStrongETag(value string) ETag {
	return ETag{
		Value: value,
		Weak:  false,
	}
}

// NewWeakETag returns an ETag with ETag.Weak set to true.
func NewWeakETag(value string) ETag {
	return ETag{
		Value: value,
		Weak:  true,
	}
}

// String implements the fmt.Stringer interface. If ETag.Weak is true, the prefix "W/" will be added.
func (e ETag) String() string {
	if e.Weak {
		return `W/"` + e.Value + `"`
	}
	return `"` + e.Value + `"`
}

// Directives contains the parsed values of either "If-Match" or "If-None-Match" request header values.
type Directives struct {
	ETags []ETag
	Any   bool
}

var weakETag = regexp.MustCompile(`^W/"(?P<value>.+)"$`)
var strongETag = regexp.MustCompile(`^"(?P<value>.+)"$`)

// ParseDirectives parses the "If-Match" or "If-None-Match" header value and returns the directives.
//
// If value is empty, return nil, nil,
func ParseDirectives(value string) (*Directives, error) {
	if value == "" {
		return nil, nil
	}

	if value == "*" {
		return &Directives{Any: true}, nil
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

	return &Directives{ETags: etags}, nil
}
