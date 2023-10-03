package codepipelinelambdaaction

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nguyengg/golambda/configsupport"
	"github.com/nguyengg/golambda/logsupport"
	"github.com/nguyengg/golambda/metrics"
	"github.com/nguyengg/golambda/start"
	"log"
)

// Handler for CodePipeline Lambda action.
type Handler func(ctx context.Context, request events.CodePipelineEvent) (err error)

// Start starts the Lambda runtime loop with the specified Handler.
func Start(handler Handler, options ...start.Option) {
	opts := start.New(options)

	lambda.Start(func(ctx context.Context, request events.CodePipelineEvent) (err error) {
		ctx, m := metrics.NewSimpleMetricsContext(
			opts.LoggerProvider(ctx).WithContext(ctx),
			"",
			0)

		if !opts.DisableSetUpGlobalLogger {
			defer logsupport.SetUpGlobalLogger(ctx)()
		}

		if !opts.DisableRequestDebugLogging && configsupport.IsDebug() {
			data, err := json.Marshal(request)
			if err != nil {
				log.Printf("ERROR marshal request: %v\n", err)
			} else {
				log.Printf("INFO request: %s\n", data)
			}
		}

		panicked := true

		if !opts.DisableMetricsLogging {
			m.
				SetProperty("codePipelineJobId", request.CodePipelineJob.ID).
				SetProperty("accountId", request.CodePipelineJob.AccountID).
				SetProperty("functionName", request.CodePipelineJob.Data.ActionConfiguration.Configuration.FunctionName).
				SetProperty("userParameters", request.CodePipelineJob.Data.ActionConfiguration.Configuration.UserParameters)

			defer func() {
				if panicked {
					m.Panicked()
				}
				if err != nil {
					m.Faulted()
				}

				m.Log()
			}()

		}

		err = handler(ctx, request)
		panicked = false
		return
	})
}
