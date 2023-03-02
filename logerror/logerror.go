package logerror

import (
	"errors"
	"github.com/aws/smithy-go"
	"log"
)

// LogAPIError checks that the given error is of type smithy.OperationError and/or smithy.APIError and logs the fields.
// Returns in this order: service, operation, code, message, and fault.
// service and operation come from smithy.OperationError; the rest from smithy.APIError.
func LogAPIError(err error) (service, operation, code, message string, fault smithy.ErrorFault) {
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
				log.Printf("ERROR %s.%s error: (%s) %s ", service, operation, code, message)
			case smithy.FaultServer:
				log.Printf("ERROR %s.%s fault: (%s) %s ", service, operation, code, message)
			default:
				log.Printf("ERROR %s.%s failure: (%s) %s ", service, operation, code, message)
			}
			return
		}

		switch fault {
		case smithy.FaultClient:
			log.Printf("ERROR unknown API error: (%s) %s ", code, message)
		case smithy.FaultServer:
			log.Printf("ERROR unknown API fault: (%s) %s ", code, message)
		default:
			log.Printf("ERROR unknown API failure: (%s) %s ", code, message)
		}
		return
	}

	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		service = oe.Service()
		operation = oe.Operation()

		log.Printf("ERROR %s.%s error: %#v", service, operation, oe.Error())
		return
	}

	log.Printf("ERROR unknown error: %#v", err)
	return
}
