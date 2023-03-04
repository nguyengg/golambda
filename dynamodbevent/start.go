package dynamodbevent

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/nguyengg/golambda/metrics"
	"log"
	"time"
)

// Handler for DynamoDB events.
type Handler func(ctx context.Context, request events.DynamoDBEvent) error

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler) {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.LUTC | log.Lshortfile | log.Lmsgprefix)

	lambda.Start(func(ctx context.Context, request events.DynamoDBEvent) (err error) {
		startTime := time.Now().UTC()

		lc, ok := lambdacontext.FromContext(ctx)
		if !ok {
			return fmt.Errorf("no lambda context")
		}
		log.SetPrefix(lc.AwsRequestID + " ")

		var m *metrics.Metrics

		defer func() {
			switch r := recover(); {
			case r != nil:
				if e, ok := r.(error); ok {
					log.Printf("ERROR panicked with error: %v", e)
					err = e
					break
				}
				log.Printf("ERROR panicked due to: %v", r)
				err = fmt.Errorf("recover: %v", r)
				if m != nil {
					_ = m.Faulted()
					_ = m.Panicked()
				}
			case err != nil:
				log.Printf("ERROR failed with error: %v", err)
				if m != nil {
					_ = m.Faulted()
				}
			}

			if m == nil {
				log.Printf("ERROR cannot emit metrics since it wasn't created properly")
				return
			}
			fmt.Printf("%s\n", m.JSONString())
		}()

		m = metrics.NewWithStartTime(startTime)
		_ = m.SetProperty("lambdaRequestId", lc.AwsRequestID)
		_ = m.AddCount("recordCount", int64(len(request.Records)))

		err = handler(metrics.NewContext(ctx, m), request)
		return
	})
}
