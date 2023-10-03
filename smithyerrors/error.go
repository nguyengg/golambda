package smithyerrors

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/smithy-go"
)

// Parse uses [errors.As] to check if the given error is [http.ResponseError], [smithy.APIError] and/or
// [smithy.OperationError].
//
// The returned statusCode comes from [http.ResponseError]; service and operation from smithy.OperationError; the rest
// smithy.APIError.
//
// Generally you should prefer using [errors.As] on the particular error you're trying to catch (see
// https://aws.github.io/aws-sdk-go-v2/docs/handling-errors/#service-client-errors), but if it's insufficient or doesn't
// work (e.g. [.s3.IsNoSuchKey]) then use this.
func Parse(err error) (statusCode int, service, operation, code, message string, fault smithy.ErrorFault) {
	var re *http.ResponseError
	if errors.As(err, &re) {
		statusCode = re.HTTPStatusCode()
	}

	var ae smithy.APIError
	if errors.As(err, &ae) {
		code = ae.ErrorCode()
		message = ae.ErrorMessage()
		fault = ae.ErrorFault()
	}

	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		service = oe.Service()
		operation = oe.Operation()
	}

	return
}
