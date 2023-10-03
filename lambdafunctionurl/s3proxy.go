package lambdafunctionurl

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"log"
	"net/http"
	"time"
)

func (c *baseContext[T]) ProxyS3(client *s3.Client, bucket, key string) error {
	return c.ProxyS3WithRequestHeaders(client, bucket, key, http.Header{})
}

func (c *baseContext[T]) ProxyS3WithRequestHeaders(client *s3.Client, bucket, key string, header http.Header) error {
	switch c.RequestMethod() {
	case http.MethodGet:
		return doGET(c, client, &s3.GetObjectInput{
			Bucket:            aws.String(bucket),
			Key:               aws.String(key),
			IfMatch:           getIfMatch(header),
			IfModifiedSince:   getIfModifiedSince(header),
			IfNoneMatch:       getIfNoneMatch(header),
			IfUnmodifiedSince: getIfUnmodifiedSince(header),
			Range:             getRange(header),
		})
	case http.MethodHead:
		return doHEAD(c, client, &s3.HeadObjectInput{
			Bucket:            aws.String(bucket),
			Key:               aws.String(key),
			IfMatch:           getIfMatch(header),
			IfModifiedSince:   getIfModifiedSince(header),
			IfNoneMatch:       getIfNoneMatch(header),
			IfUnmodifiedSince: getIfUnmodifiedSince(header),
			Range:             getRange(header),
		})
	default:
		return c.RespondMethodNotAllowed("GET, HEAD")
	}
}
func doGET[T any](c *baseContext[T], client *s3.Client, input *s3.GetObjectInput) error {
	output, err := client.GetObject(c.Context(), input)
	if err != nil {
		return c.RespondFormattedStatus(toStatusCode(err))
	}

	c.SetStatusCode(http.StatusOK)
	for k, v := range headersForGetObjectOutput(output) {
		c.SetResponseHeader(k, v)
	}

	return c.RespondOKWithBody(output.Body)
}

func doHEAD[T any](c *baseContext[T], client *s3.Client, input *s3.HeadObjectInput) error {
	output, err := client.HeadObject(c.Context(), input)
	if err != nil {
		return c.RespondFormattedStatus(toStatusCode(err))
	}

	c.SetStatusCode(http.StatusOK)
	for k, v := range headersForHeadObjectOutput(output) {
		c.SetResponseHeader(k, v)
	}

	return nil
}

func toStatusCode(err error) int {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return http.StatusNotFound
	}

	var e smithy.APIError
	if errors.As(err, &e) {
		switch e.ErrorCode() {
		case "NotFound":
			return http.StatusNotFound
		case "NotModified":
			return http.StatusNotModified
		case "PreconditionFailed":
			return http.StatusPreconditionFailed
		}
	}

	var re *awshttp.ResponseError
	if errors.As(err, &re) {
		return re.Response.StatusCode
	}

	log.Printf("unknown S3 error: %#v", e)

	return http.StatusInternalServerError
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

func headersForHeadObjectOutput(output *s3.HeadObjectOutput) map[string]string {
	return map[string]string{
		"Content-Type":  aws.ToString(output.ContentType),
		"ETag":          aws.ToString(output.ETag),
		"Last-Modified": output.LastModified.Format(http.TimeFormat),
	}
}

func headersForGetObjectOutput(output *s3.GetObjectOutput) map[string]string {
	return map[string]string{
		"Content-Type":  aws.ToString(output.ContentType),
		"ETag":          aws.ToString(output.ETag),
		"Last-Modified": output.LastModified.Format(http.TimeFormat),
	}
}
