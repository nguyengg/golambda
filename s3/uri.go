package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"regexp"
	"strings"
)

// URIWithOwner contains required Bucket, Key, and ExpectedBucketOwner fields.
type URIWithOwner struct {
	Bucket              string
	Key                 string
	ExpectedBucketOwner string
}

var uriWithOwnerPattern = regexp.MustCompile(`s3://([a-z0-9][a-z0-9.-]+?)\[(\d+)](/(.*))?`)

// Parse parses a URL in expected format s3://bucket[owner]/key.
//
// Only the bucket name and expected bucket owner is required. The key can be empty, or can be a prefix that possibly
// ends in "/".
func Parse(rawURL string) (value URIWithOwner, err error) {
	if !strings.HasPrefix(rawURL, "s3://") {
		err = fmt.Errorf("URL does not start with s3://")
		return
	}

	m := uriWithOwnerPattern.FindStringSubmatch(rawURL)
	if len(m) != 5 {
		err = fmt.Errorf("URL is not in format s3://bucket[owner]/key")
		return
	}

	value.Bucket = m[1]
	value.ExpectedBucketOwner = m[2]
	value.Key = m[4]

	return
}

// Append creates a new URIWithOwner by appending the given key to the existing key.
//
// The new key is a simple concatenation of existing key + new key without any '/' separator.
func (u URIWithOwner) Append(key string) URIWithOwner {
	return URIWithOwner{
		Bucket:              u.Bucket,
		Key:                 u.Key + key,
		ExpectedBucketOwner: u.ExpectedBucketOwner,
	}
}

// Get decorates the given s3.GetObjectInput the fields from the URIWithOwner.
//
// If given a nil input, a new one will be created.
func (u URIWithOwner) Get(input *s3.GetObjectInput) *s3.GetObjectInput {
	if input == nil {
		input = &s3.GetObjectInput{}
	}

	input.Bucket = aws.String(u.Bucket)
	input.Key = aws.String(u.Key)
	input.ExpectedBucketOwner = aws.String(u.ExpectedBucketOwner)
	return input
}

// Head decorates the given s3.HeadObjectInput the fields from the URIWithOwner.
//
// If given a nil input, a new one will be created.
func (u URIWithOwner) Head(input *s3.HeadObjectInput) *s3.HeadObjectInput {
	if input == nil {
		input = &s3.HeadObjectInput{}
	}

	input.Bucket = aws.String(u.Bucket)
	input.Key = aws.String(u.Key)
	input.ExpectedBucketOwner = aws.String(u.ExpectedBucketOwner)
	return input
}

// Put decorates the given s3.PutObjectInput the fields from the URIWithOwner.
//
// If given a nil input, a new one will be created.
func (u URIWithOwner) Put(input *s3.PutObjectInput) *s3.PutObjectInput {
	if input == nil {
		input = &s3.PutObjectInput{}
	}

	input.Bucket = aws.String(u.Bucket)
	input.Key = aws.String(u.Key)
	input.ExpectedBucketOwner = aws.String(u.ExpectedBucketOwner)
	return input
}

// String returns s3://bucket/key, or s3://bucket if key is empty string.
func (u URIWithOwner) String() string {
	if u.Key == "" {
		return "s3://" + u.Bucket
	}
	return "s3://" + u.Bucket + "/" + u.Key
}

// GoString returns s3://bucket[owner]/key, or s3://bucket[owner] if key is empty string.
func (u URIWithOwner) GoString() string {
	if u.Key == "" {
		return fmt.Sprintf("s3://%s[%s]", u.Bucket, u.ExpectedBucketOwner)
	}
	return fmt.Sprintf("s3://%s[%s]/%s", u.Bucket, u.ExpectedBucketOwner, u.Key)
}
