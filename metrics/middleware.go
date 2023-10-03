package metrics

import (
	"context"
	awsmw "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go"
	smithymw "github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

// ClientSideMetricsMiddleware adds client-side latency metrics about the requests made by this stack.
//
// Usage:
//
//	cfg, _ := config.LoadDefaultConfig(ctx)
//	cfg.APIOptions = append(cfg.APIOptions, metrics.ClientSideMetricsMiddleware())
//
// A metrics.Metrics instance must be available from context by the time the middleware receives a response.
func ClientSideMetricsMiddleware(options ...Option) func(stack *smithymw.Stack) error {
	c := &clientSideMetricsMiddleware{}

	for _, option := range options {
		option(c)
	}

	return func(stack *smithymw.Stack) error {
		return stack.Deserialize.Add(&clientSideMetricsMiddleware{}, smithymw.After)
	}
}

// Should implement middleware.DeserializeMiddleware.
type clientSideMetricsMiddleware struct {
	disableDebugLoggingInput bool
}

type Option func(*clientSideMetricsMiddleware)

// DisableDebugLoggingInput disables feature where the requests are logged at Debug level.
func DisableDebugLoggingInput(c *clientSideMetricsMiddleware) {
	c.disableDebugLoggingInput = true
}

func (c clientSideMetricsMiddleware) ID() string {
	return "ClientSideLatencyMetrics"
}

func (c clientSideMetricsMiddleware) HandleDeserialize(ctx context.Context, input smithymw.DeserializeInput, handler smithymw.DeserializeHandler) (smithymw.DeserializeOutput, smithymw.Metadata, error) {
	start := time.Now().UTC()

	output, metadata, err := handler.HandleDeserialize(ctx, input)

	end := time.Now().UTC()
	if t, ok := awsmw.GetResponseAt(metadata); ok {
		end = t
	}

	serviceId := awsmw.GetServiceID(ctx)
	operationName := awsmw.GetOperationName(ctx)

	logContext := zerolog.Ctx(ctx)
	logger := logContext.
		Log().
		Str("service", serviceId).
		Str("operation", operationName).
		Int64(ReservedKeyStartTime, start.UnixNano()/int64(time.Millisecond)).
		Str(ReservedKeyEndTime, end.Format(http.TimeFormat)).
		Str(ReservedKeyTime, FormatDuration(end.Sub(start)))

	counters := zerolog.Dict()

	switch resp := output.RawResponse.(type) {
	case *smithyhttp.Response:
		logger.Int("statusCode", resp.StatusCode)

		switch resp.StatusCode / 100 {
		case 2:
			counters.Int("2xx", 1).Int("4xx", 0).Int("5xx", 0)
		case 4:
			counters.Int("2xx", 0).Int("4xx", 1).Int("5xx", 0)
		case 5:
			counters.Int("2xx", 0).Int("4xx", 0).Int("5xx", 1)
		default:
			counters.Int("2xx", 0).Int("4xx", 0).Int("5xx", 0)
		}
	}

	// DynamoDB GetItem => DynamoDB.GetItem
	// log filter can use { $.['DynamoDB.GetItem.Fault'] = * }
	// see https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html#matching-terms-json-log-events
	key := serviceId + "." + operationName

	m := Ctx(ctx)
	m.AddTiming(key, end.Sub(start))

	if err != nil {
		_, _, _, _, fault := logsupport.LogSmithyError(err)

		switch fault {
		case smithy.FaultClient:
			m.AddCount(key+".ClientFault", 1, key+".ServerFault", key+".UnknownFault")
			counters.Int("clientFault", 1).Int("serverFault", 0).Int("unknownFault", 0)
			logger.AnErr("clientError", err)
		case smithy.FaultServer:
			m.AddCount(key+".ServerFault", 1, key+".ClientFault", key+".UnknownFault")
			counters.Int("clientFault", 0).Int("serverFault", 1).Int("unknownFault", 0)
			logger.AnErr("serverError", err)
		default:
			m.AddCount(key+".UnknownFault", 1, key+".ClientFault", key+".ServerFault")
			counters.Int("clientFault", 0).Int("serverFault", 0).Int("unknownFault", 1)
			logger.AnErr("unknownError", err)
		}
	} else {
		m.AddCount(key+".ClientFault", 0, key+".ServerFault", key+".UnknownFault")
		counters.Int("clientFault", 0).Int("serverFault", 0).Int("unknownFault", 0)
	}

	logger.Dict("counters", counters).Msg("")

	return output, metadata, err
}
