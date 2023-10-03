package apigatewayhttpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"time"
)

// ProxyS3 will call S3 with the appropriate GET or HEAD method and return the response as either plain text or
// base64-encoded data.
//
// The argument method will determine whether ProxyS3GET or ProxyS3HEAD will be used. If the method is not recognized,
// http.StatusMethodNotAllowed will be returned.
func ProxyS3(ctx context.Context, client *s3.Client, method, bucket, key string, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	return ProxyS3WithRequestHeaders(ctx, client, method, bucket, key, http.Header{}, opts...)
}

// ProxyS3WithRequestHeaders is a variant of ProxyS3 that is given an extra http.Header whose values will be passed into
// the S3's respective requests if the action supports it.
func ProxyS3WithRequestHeaders(ctx context.Context, client *s3.Client, method, bucket, key string, header http.Header, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	switch method {
	case http.MethodGet:
		return ProxyS3GETWithRequestHeaders(ctx, client, bucket, key, header, opts...)
	case http.MethodHead:
		return ProxyS3HEADWithRequestHeaders(ctx, client, bucket, key, header, opts...)
	default:
		res := events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Headers:    map[string]string{"Allow": "GET, HEAD"},
		}

		for _, opt := range opts {
			opt(&res)
		}

		return res, nil
	}
}

// ProxyS3GET is ProxyS3 for http.MethodGet and will call S3's GetObject.
func ProxyS3GET(ctx context.Context, client *s3.Client, bucket, key string, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	return ProxyS3GETWithRequestHeaders(ctx, client, bucket, key, http.Header{}, opts...)
}

// ProxyS3GETWithRequestHeaders is a variant of ProxyS3GET with request headers.
//
// Only these headers are proxied: If-Match, If-Modified-Since, If-None-Match, If-Unmodified-Since, and Range.
func ProxyS3GETWithRequestHeaders(ctx context.Context, client *s3.Client, bucket, key string, header http.Header, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	res, err := doGET(ctx, client, &s3.GetObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key),
		IfMatch:           getIfMatch(header),
		IfModifiedSince:   getIfModifiedSince(header),
		IfNoneMatch:       getIfNoneMatch(header),
		IfUnmodifiedSince: getIfUnmodifiedSince(header),
		Range:             getRange(header),
	})
	if err != nil {
		return res, err
	}

	for _, opt := range opts {
		opt(&res)
	}

	return res, err
}

// ProxyS3HEAD is ProxyS3 for http.MethodHead and will call S3's HeadObject.
func ProxyS3HEAD(ctx context.Context, client *s3.Client, bucket, key string, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	return ProxyS3HEADWithRequestHeaders(ctx, client, bucket, key, http.Header{}, opts...)
}

// ProxyS3HEADWithRequestHeaders is a variant of ProxyS3HEAD with request headers.
//
// Only these headers are proxied: If-Match, If-Modified-Since, If-None-Match, If-Unmodified-Since, and Range.
func ProxyS3HEADWithRequestHeaders(ctx context.Context, client *s3.Client, bucket, key string, header http.Header, opts ...Opt) (events.APIGatewayV2HTTPResponse, error) {
	res, err := doHEAD(ctx, client, &s3.HeadObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key),
		IfMatch:           getIfMatch(header),
		IfModifiedSince:   getIfModifiedSince(header),
		IfNoneMatch:       getIfNoneMatch(header),
		IfUnmodifiedSince: getIfUnmodifiedSince(header),
		Range:             getRange(header),
	})
	if err != nil {
		return res, err
	}

	for _, opt := range opts {
		opt(&res)
	}

	return res, err
}

func doGET(ctx context.Context, client *s3.Client, input *s3.GetObjectInput) (events.APIGatewayV2HTTPResponse, error) {
	output, err := client.GetObject(ctx, input)
	if err != nil {
		return convertS3Error(err), nil
	}

	data, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}, fmt.Errorf("read S3 content: %v", err)
	}

	if output.ContentType != nil {
		if t, _, err := mime.ParseMediaType(*output.ContentType); err == nil && strings.HasPrefix(t, "text") {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       string(data),
				Headers:    headersForGetObjectOutput(output),
			}, nil
		}
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      http.StatusOK,
		Body:            base64.StdEncoding.EncodeToString(data),
		Headers:         headersForGetObjectOutput(output),
		IsBase64Encoded: true,
	}, nil
}

func doHEAD(ctx context.Context, client *s3.Client, input *s3.HeadObjectInput) (events.APIGatewayV2HTTPResponse, error) {
	output, err := client.HeadObject(ctx, input)
	if err != nil {
		return convertS3Error(err), nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers:    headersForHeadObjectOutput(output),
	}, nil
}

func convertS3Error(err error) events.APIGatewayV2HTTPResponse {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNotFound}
	}

	var e smithy.APIError
	if errors.As(err, &e) {
		switch e.ErrorCode() {
		case "NotFound":
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNotFound}
		case "NotModified":
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusNotModified}
		case "PreconditionFailed":
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusPreconditionFailed}
		}
	}

	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusInternalServerError}
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
