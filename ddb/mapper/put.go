package mapper

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/expr"
	"reflect"
)

// Put makes a DynamoDB PutItem request.
//
// optFns are applied before any logic regarding PutOpts.OptimisticLockingEnabled and
// PutOpts.AutoGenerateTimestampsEnabled take place to allow the disabling of those features. Use PutOpts.Finalizer if
// you need to apply additional modifiers right before executing the DynamoDB PutItem request.
func (m Mapper[T]) Put(ctx context.Context, item T, optFns ...func(*PutOpts[T])) (*dynamodb.PutItemOutput, error) {
	value := reflect.ValueOf(item)

	mav, err := m.MarshalMap(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item error: %w", err)
	}

	opts := &PutOpts[T]{
		Item: item,
		Input: &dynamodb.PutItemInput{
			Item:      mav,
			TableName: &m.model.tableName,
		},
		OptimisticLockingEnabled:      m.model.putVersion != nil,
		AutoGenerateTimestampsEnabled: m.model.putTimestamps != nil,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.OptimisticLockingEnabled {
		if m.model.putVersion == nil {
			return nil, fmt.Errorf("OptimisticLockingEnabled must be false because item does not implement HasVersion")
		}

		c, err := m.model.putVersion(item, value, opts.Input.Item)
		if err != nil {
			return nil, fmt.Errorf("create version condition expression error: %w", err)
		}
		if c.IsSet() {
			opts.Condition = expr.And(opts.Condition, c)
		}
	}

	if opts.AutoGenerateTimestampsEnabled {
		if m.model.putTimestamps == nil {
			return nil, fmt.Errorf("AutoGenerateTimestampsEnabled must be false because item does not implement HasTimestamps")
		}

		if err = m.model.putTimestamps(item, value, opts.Input.Item); err != nil {
			return nil, fmt.Errorf("create timestamp attributes error: %w", err)
		}
	}

	if opts.Condition != nil {
		e, err := expression.NewBuilder().WithCondition(*opts.Condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build condition expression error: %w", err)
		}
		opts.Input.ConditionExpression = e.Condition()
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ExpressionAttributeValues = e.Values()
	}

	if opts.Finalizer != nil {
		opts.Finalizer(opts.Input)
	}

	return m.client.PutItem(ctx, opts.Input)
}

// PutOpts provides customisation to the [Mapper.Put] operation.
//
// PutOpts.Input is guaranteed to exist when passed into the first modifier.
// PutOpts.Item is the original reference item. Changes to PutOpts.Item don't automatically update PutOpts.Input.
// Changes to PutOpts.Condition will affect the final PutOpts.Input.
//
// OptimisticLockingEnabled and AutoGenerateTimestampsEnabled will be true initially if Mapper detects that the item
// supports HasVersion and HasTimestamps respectively.
//
// Finalizer is the modifier applied to the final PutOpts.Input right before executing the DynamoDB PutItem request.
type PutOpts[T interface{}] struct {
	Item                          T
	Input                         *dynamodb.PutItemInput
	Condition                     *expression.ConditionBuilder
	OptimisticLockingEnabled      bool
	AutoGenerateTimestampsEnabled bool
	Finalizer                     func(*dynamodb.PutItemInput)
}

// WithTableName changes the table.
func (o *PutOpts[T]) WithTableName(tableName string) *PutOpts[T] {
	o.Input.TableName = &tableName
	return o
}

// DisableOptimisticLocking explicitly disables [HasVersion] logic.
func (o *PutOpts[T]) DisableOptimisticLocking() *PutOpts[T] {
	o.OptimisticLockingEnabled = false
	return o
}

// DisableAutoGenerateTimestamps explicitly disables [HasTimestamps] logic.
func (o *PutOpts[T]) DisableAutoGenerateTimestamps() *PutOpts[T] {
	o.AutoGenerateTimestampsEnabled = false
	return o
}

// ReturnAllOldValues sets the [dynamodb.PutItemInput.ReturnValues] to ALL_OLD.
//
// [dynamodb.PutItemInput.ReturnValues] only support NONE or ALL_OLD.
func (o *PutOpts[T]) ReturnAllOldValues() *PutOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
	return o
}

// ReturnAllOldValuesOnConditionCheckFailure sets the [dynamodb.PutItemInput.ReturnValuesOnConditionCheckFailure] to ALL_OLD.
//
// [dynamodb.PutItemInput.ReturnValuesOnConditionCheckFailure] supports NONE or ALL_OLD.
func (o *PutOpts[T]) ReturnAllOldValuesOnConditionCheckFailure() *PutOpts[T] {
	o.Input.ReturnValuesOnConditionCheckFailure = dynamodbtypes.ReturnValuesOnConditionCheckFailureAllOld
	return o
}
