package framework

import (
	"net/http"
	"time"
)

// WithResponseCachingHeaders allows implement types to generate ETag and Last-Modified response headers for caching
// purposes.
type WithResponseCachingHeaders interface {
	ETag() *ETag
	LastModified() *time.Time
}

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

// SetResponseCachingHeaders adds ETag and Last-Modified headers to the response.
func (c *Context) SetResponseCachingHeaders(v WithResponseCachingHeaders) {
	etag := v.ETag()
	if etag != nil {
		c.responseHeader.Set("ETag", etag.String())
	}

	t := v.LastModified()
	if t != nil {
		c.responseHeader.Set("Last-Modified", t.Format(http.TimeFormat))
	}
}
