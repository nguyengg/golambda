package s3

import (
	"fmt"
	"regexp"
	"strings"
)

// S3URIWithOwner contains required Bucket, Key, and ExpectedBucketOwner fields.
type S3URIWithOwner struct {
	Bucket              string
	Key                 string
	ExpectedBucketOwner string
}

var s3UriWithOwnerPattern = regexp.MustCompile(`s3://([a-z0-9][a-z0-9.-]+?)\[(\d+)](/(.*))?`)

// ParseS3URIWithOwner parses a URL in expected format s3://bucket[owner]/key.
//
// Only the bucket name and expected bucket owner is required. The key can be empty, or can be a prefix that possibly
// ends in "/".
func ParseS3URIWithOwner(rawURL string) (value S3URIWithOwner, err error) {
	if !strings.HasPrefix(rawURL, "s3://") {
		err = fmt.Errorf("URL does not start with s3://")
		return
	}

	m := s3UriWithOwnerPattern.FindStringSubmatch(rawURL)
	if len(m) != 5 {
		err = fmt.Errorf("URL is not in format s3://bucket[owner]/key")
		return
	}

	value.Bucket = m[1]
	value.ExpectedBucketOwner = m[2]
	value.Key = m[4]

	return
}
