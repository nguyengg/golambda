package s3

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
	"strconv"
	"time"
)

// AddToGetObject adds conditional request headers to the [s3.GetObjectInput] and return the same input.
//
// Specifically, these headers are parsed from the "header" argument:
//   - If-Match and If-None-Match
//   - If-Modified-Since and If-Unmodified-Since
//   - Range
func AddToGetObject(input *s3.GetObjectInput, header http.Header) *s3.GetObjectInput {
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// AddToHeadObject adds conditional request headers to the [s3.HeadObjectInput] and return the same input.
//
// Specifically, these headers are parsed from the "header" argument:
//   - If-Match and If-None-Match
//   - If-Modified-Since and If-Unmodified-Since
//   - Range
func AddToHeadObject(input *s3.HeadObjectInput, header http.Header) *s3.HeadObjectInput {
	input.IfMatch = getIfMatch(header)
	input.IfModifiedSince = getIfModifiedSince(header)
	input.IfNoneMatch = getIfNoneMatch(header)
	input.IfUnmodifiedSince = getIfUnmodifiedSince(header)
	input.Range = getRange(header)
	return input
}

// HeadersFromGetObjectOutput parses response headers from the [s3.GetObjectOutput] and passes it to the callback.
func HeadersFromGetObjectOutput(output *s3.GetObjectOutput, cb func(k, v string)) {
	if output.ContentDisposition != nil {
		cb("Content-Disposition", *output.ContentDisposition)
	}
	if output.ContentEncoding != nil {
		cb("Content-Encoding", *output.ContentEncoding)
	}
	if output.ContentLanguage != nil {
		cb("Content-Language", *output.ContentLanguage)
	}
	if output.ContentLength != 0 {
		cb("Content-Length", strconv.FormatInt(output.ContentLength, 10))
	}
	if output.ContentRange != nil {
		cb("Content-Range", *output.ContentRange)
	}
	if output.ContentType != nil {
		cb("Content-Type", *output.ContentType)
	}
	if output.ETag != nil {
		cb("ETag", *output.ETag)
	}
	if output.Expires != nil {
		cb("Expires", output.Expires.Format(http.TimeFormat))
	}
	if output.LastModified != nil {
		cb("Last-Modified", output.LastModified.Format(http.TimeFormat))
	}
}

// HeadersFromHeadObjectOutput parses response headers from the [s3.HeadObjectOutput] and passes it to the callback.
func HeadersFromHeadObjectOutput(output *s3.HeadObjectOutput, cb func(k, v string)) {
	if output.ContentDisposition != nil {
		cb("Content-Disposition", *output.ContentDisposition)
	}
	if output.ContentEncoding != nil {
		cb("Content-Encoding", *output.ContentEncoding)
	}
	if output.ContentLanguage != nil {
		cb("Content-Language", *output.ContentLanguage)
	}
	if output.ContentLength != 0 {
		cb("Content-Length", strconv.FormatInt(output.ContentLength, 10))
	}
	if output.ContentType != nil {
		cb("Content-Type", *output.ContentType)
	}
	if output.ETag != nil {
		cb("ETag", *output.ETag)
	}
	if output.Expires != nil {
		cb("Expires", output.Expires.Format(http.TimeFormat))
	}
	if output.LastModified != nil {
		cb("Last-Modified", output.LastModified.Format(http.TimeFormat))
	}
}

func getIfMatch(header http.Header) *string {
	value := header.Get("If-Match")
	if value == "" {
		return nil
	}
	return &value
}

func getIfModifiedSince(header http.Header) *time.Time {
	value := header.Get("If-Modified-Since")
	if value == "" {
		return nil
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return nil
	}

	return &t
}

func getIfNoneMatch(header http.Header) *string {
	value := header.Get("If-None-Match")
	if value == "" {
		return nil
	}
	return &value
}

func getIfUnmodifiedSince(header http.Header) *time.Time {
	value := header.Get("If-Unmodified-Since")
	if value == "" {
		return nil
	}

	t, err := http.ParseTime(value)
	if err != nil {
		return nil
	}

	return &t
}

func getRange(header http.Header) *string {
	value := header.Get("Range")
	if value == "" {
		return nil
	}
	return &value
}
