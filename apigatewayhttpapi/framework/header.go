package framework

import (
	"net/http"
	"time"
)

type WithResponseHeaders interface {
	ETag() *ETag
	LastModified() *time.Time
}

// ETag header value.
type ETag struct {
	Value string
	Weak  bool
}

func NewStrongETag(value string) ETag {
	return ETag{
		Value: value,
		Weak:  false,
	}
}

func NewWeakETag(value string) ETag {
	return ETag{
		Value: value,
		Weak:  true,
	}
}

func (e ETag) String() string {
	if e.Weak {
		return `W/"` + e.Value + `"`
	}
	return `"` + e.Value + `"`
}

func (c *Context) SetResponseHeaders(v WithResponseHeaders) {
	etag := v.ETag()
	if etag != nil {
		c.responseHeader.Set("Etag", etag.String())
	}

	t := v.LastModified()
	if t != nil {
		c.responseHeader.Set("Last-Modified", t.Format(http.TimeFormat))
	}
}
