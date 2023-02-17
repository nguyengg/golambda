package apigatewayhttpapi

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"reflect"
)

type in int

const (
	receiveNothing in = 1 << iota
	receiveContext
	receiveHttpV2Request
)

type out int

const (
	returnNothing out = 1 << iota
	returnError
	returnHttpV2Response
	returnOtherResponse
)

// Adapted from lambda.NewHandler.
func validate(v interface{}) (in, out, error) {
	if v == nil {
		return 0, 0, fmt.Errorf("nil handler")
	}

	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Func {
		return 0, 0, fmt.Errorf("handler must be a function instead of type %s", t.Kind())
	}

	take, err := validateIn(t)
	if err != nil {
		return 0, 0, err
	}

	give, err := validateOut(t)
	if err != nil {
		return 0, 0, err
	}

	// https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html#golang-handler-signatures forbids a few others.
	switch {
	case take != receiveNothing && give == returnNothing:
		return 0, 0, fmt.Errorf("handler must return at least one value if it also takes at least one argument")
	case take == receiveHttpV2Request && give != returnError:
		return 0, 0, fmt.Errorf("if the sole argument is events.APIGatewayV2HTTPRequest then handler must return exactly one value that implements error type")
	}
	return take, give, nil
}

func validateIn(handlerType reflect.Type) (take in, err error) {
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	requestType := reflect.TypeOf((*events.APIGatewayV2HTTPRequest)(nil)).Elem()

	switch handlerType.NumIn() {
	case 0:
		take = receiveNothing
	case 1:
		arg := handlerType.In(0)
		if arg.Implements(contextType) {
			take = receiveContext
		} else if arg == requestType {
			take = receiveHttpV2Request
		} else {
			return 0, fmt.Errorf("the sole argument must be either context.Context or events.APIGatewayV2HTTPRequest instead of %s", arg.Kind())
		}
	case 2:
		arg := handlerType.In(0)
		if arg.Implements(contextType) {
			take = receiveContext
		} else {
			return 0, fmt.Errorf("the 1st argument must be context.Context instead of %s", arg.Kind())
		}

		arg = handlerType.In(1)
		if arg == requestType {
			take |= receiveHttpV2Request
		} else {
			return 0, fmt.Errorf("the 2nd argument must be events.APIGatewayV2HTTPRequest instead of %s", arg.Kind())
		}
	default:
		return 0, fmt.Errorf("can only receive up to 2 arguments instead of %d", handlerType.NumIn())
	}

	return
}

func validateOut(handlerType reflect.Type) (give out, err error) {
	httpV2ResponseType := reflect.TypeOf((*events.APIGatewayV2HTTPResponse)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	switch handlerType.NumOut() {
	case 0:
		give = returnNothing
	case 1:
		val := handlerType.Out(0)
		if val.Implements(errorType) {
			give = returnError
		} else {
			return 0, fmt.Errorf("the sole returned value must be error instead of %s", val.Kind())
		}
	case 2:
		val := handlerType.Out(0)
		if val == httpV2ResponseType {
			give = returnHttpV2Response
		} else {
			give = returnOtherResponse
		}

		val = handlerType.Out(1)
		if val.Implements(errorType) {
			give |= returnError
		} else {
			return 0, fmt.Errorf("the 2nd returned value must be error instead of %s", val.Kind())
		}
	default:
		return 0, fmt.Errorf("can only return up to 2 values instead of %d", handlerType.NumOut())

	}

	return
}
