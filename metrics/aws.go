package metrics

import (
	"context"
	. "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	"github.com/nguyengg/golambda/logerror"
	"log"
	"strings"
	"time"
	"unicode"
)

const DefaultId = "ClientSideLatencyMetrics"

func NewClientSideMetricsMiddleware() middleware.DeserializeMiddleware {
	return NewClientSideMetricsMiddlewareWithId(DefaultId)
}

func NewClientSideMetricsMiddlewareWithId(id string) middleware.DeserializeMiddleware {
	return middleware.DeserializeMiddlewareFunc(
		id,
		func(ctx context.Context, input middleware.DeserializeInput, handler middleware.DeserializeHandler) (middleware.DeserializeOutput, middleware.Metadata, error) {
			m, ok := FromContext(ctx)
			if !ok {
				log.Printf("WARN no metrics from context")
				return handler.HandleDeserialize(ctx, input)
			}

			start := time.Now().UTC()

			output, metadata, err := handler.HandleDeserialize(ctx, input)

			end := time.Now().UTC()

			key := computeKey(ctx)
			_ = m.AddTiming(key, end.Sub(start))
			_ = m.AddCount(key+".fault", 0)
			_ = m.AddCount(key+".error", 0)
			_ = m.AddCount(key+".failure", 0)

			if err != nil {
				_, _, _, _, fault := logerror.LogAPIError(err)

				switch fault {
				case smithy.FaultClient:
					_ = m.AddCount(key+".error", 1)
				case smithy.FaultServer:
					_ = m.AddCount(key+".fault", 1)
				default:
					_ = m.AddCount(key+".failure", 1)
				}
			}

			return output, metadata, err
		})
}

func ClientSideMetricsAPIOption() func(stack *middleware.Stack) error {
	return ClientSideMetricsAPIOptionWithId(DefaultId)
}

func ClientSideMetricsAPIOptionWithId(id string) func(stack *middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Deserialize.Add(NewClientSideMetricsMiddlewareWithId(id), middleware.After)
	}
}

var cachedServiceIDs = map[string]string{}

func computeKey(ctx context.Context) string {
	serviceId := GetServiceID(ctx)
	operationName := GetOperationName(ctx)

	var builder strings.Builder
	builder.Grow(len(serviceId) + len(operationName))

	if value := cachedServiceIDs[serviceId]; value == "" {
		for i, ch := range serviceId {
			if i == 0 {
				ch = unicode.ToLower(ch)
			}
			if !unicode.IsSpace(ch) {
				builder.WriteRune(ch)
			}
		}
		value = builder.String()
		cachedServiceIDs[serviceId] = value
		serviceId = value
	} else {
		builder.WriteString(value)
	}

	builder.WriteRune('.')
	for i, ch := range operationName {
		if i == 0 {
			ch = unicode.ToLower(ch)
		}
		builder.WriteRune(ch)
	}

	return builder.String()
}
