package codepipelinelambdaaction

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/nguyengg/golambda/metrics"
)

// The full handler must indicate whether the job was a success or failure.
// It's an error for both success and failure to be returned.
type FullHandler func(ctx context.Context, request events.CodePipelineEvent) (success *codepipeline.PutJobSuccessResultInput, failure *codepipeline.PutJobFailureResultInput, err error)

// A simplified variant of FullHandler.
type SimpleHandler func(ctx context.Context, request events.CodePipelineEvent) (outputVariables map[string]string, failureDetails *types.FailureDetails, err error)

const (
	CounterJobSuccess = "jobSuccess"
	CounterJobFailure = "jobFailure"
)

// Wraps a FullHandler.
func WrapFullHandler(svc *codepipeline.Client, handler FullHandler) Handler {
	return func(ctx context.Context, request events.CodePipelineEvent) error {
		m := metrics.Ctx(ctx)
		m.AddCount(CounterJobSuccess, 0)
		m.AddCount(CounterJobFailure, 0)

		success, failure, err := handler(ctx, request)
		if err != nil {
			return err
		}
		if (success != nil) == (failure != nil) {
			return fmt.Errorf("handler returns both success and failure")
		}
		if success != nil {
			m.AddCount(CounterJobSuccess, 1)
			_, err = svc.PutJobSuccessResult(ctx, success)
			return err
		}
		m.AddCount(CounterJobFailure, 1)
		_, err = svc.PutJobFailureResult(ctx, failure)
		return err
	}
}

// Wraps a SimpleHandler.
func WrapSimpleHandler(svc *codepipeline.Client, handler SimpleHandler) Handler {
	return WrapFullHandler(svc, func(ctx context.Context, request events.CodePipelineEvent) (*codepipeline.PutJobSuccessResultInput, *codepipeline.PutJobFailureResultInput, error) {
		m := metrics.Ctx(ctx)
		m.AddCount(CounterJobSuccess, 0)
		m.AddCount(CounterJobFailure, 0)

		outputVariables, failureDetails, err := handler(ctx, request)
		if err != nil {
			return nil, nil, err
		}

		if failureDetails != nil {
			return nil, &codepipeline.PutJobFailureResultInput{
				FailureDetails: failureDetails,
				JobId:          aws.String(request.CodePipelineJob.ID),
			}, nil
		}

		return &codepipeline.PutJobSuccessResultInput{
			JobId:           aws.String(request.CodePipelineJob.ID),
			OutputVariables: outputVariables,
		}, nil, nil
	})
}
