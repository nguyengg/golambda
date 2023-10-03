package logsupport

import (
	"errors"
	"github.com/aws/smithy-go"
	"log"
)

// LogSmithyError checks that the given error is of type smithy.OperationError and/or smithy.APIError and logs the fields.
// Returns in this order: service, operation, code, message, and fault. See ParseSmithyError.
func LogSmithyError(err error) (service, operation, code, message string, fault smithy.ErrorFault) {
	return LogSmithyErrorWithLogger(err, log.Default())
}

// LogSmithyErrorWithLogger is a variant of LogSmithyError that allows specifying a log.Logger to use.
func LogSmithyErrorWithLogger(err error, logger *log.Logger) (service, operation, code, message string, fault smithy.ErrorFault) {
	var ae smithy.APIError
	if errors.As(err, &ae) {
		code = ae.ErrorCode()
		message = ae.ErrorMessage()
		fault = ae.ErrorFault()

		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			service = oe.Service()
			operation = oe.Operation()

			switch fault {
			case smithy.FaultClient:
				logger.Printf("ERROR %s.%s error: (%s) %s: %#v\n", service, operation, code, message, err)
			case smithy.FaultServer:
				logger.Printf("ERROR %s.%s fault: (%s) %s: %#v\n", service, operation, code, message, err)
			default:
				logger.Printf("ERROR %s.%s failure: (%s) %s: %#v\n", service, operation, code, message, err)
			}
			return
		}

		switch fault {
		case smithy.FaultClient:
			logger.Printf("ERROR unknown API error: (%s) %s: %#v\n", code, message, err)
		case smithy.FaultServer:
			logger.Printf("ERROR unknown API fault: (%s) %s: %#v\n", code, message, err)
		default:
			logger.Printf("ERROR unknown API failure: (%s) %s: %#v\n", code, message, err)
		}
		return
	}

	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		service = oe.Service()
		operation = oe.Operation()

		logger.Printf("ERROR %s.%s error: %#v\n", service, operation, err)
		return
	}

	logger.Printf("ERROR unknown error: %#v\n", err)
	return
}

// ParseSmithyError uses errors.As to check if the given error is a smithy.APIError and/or smithy.OperationError.
// service and operation return values come from smithy.OperationError, while the rest come from smithy.APIError.
func ParseSmithyError(err error) (service, operation, code, message string, fault smithy.ErrorFault) {
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
