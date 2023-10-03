package s3

import (
	"errors"
	"github.com/aws/smithy-go"
)

// IsNoSuchKey returns true if the error's [smithy.APIError.ErrorCode] returns "NoSuchKey".
//
// The documentation at https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/#api-error-responses doesn't work for
// [github.com/aws/aws-sdk-go-v2/service/s3/types#NoSuchKey] (maybe you have better luck?) but this works for me.
func IsNoSuchKey(err error) bool {
	var ae smithy.APIError
	return errors.As(err, &ae) && ae.ErrorCode() == "NoSuchKey"
}
