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

// Delete makes a DynamoDB DeleteItem request.
//
// optFns are applied before any logic regarding DeleteOpts.OptimisticLockingEnabled and
// DeleteOpts.AutoGenerateTimestampsEnabled take place to allow the disabling of those features. Use DeleteOpts.Finalizer if
// you need to apply additional modifiers right before executing the DynamoDB DeleteItem request.
func (m Mapper[T]) Delete(ctx context.Context, item T, optFns ...func(*DeleteOpts[T])) (*dynamodb.DeleteItemOutput, error) {
	value := reflect.ValueOf(item)
	key, err := m.getKey(item, value)
	if err != nil {
		return nil, fmt.Errorf("create DeleteItem's Key error: %w", err)
	}

	opts := &DeleteOpts[T]{
		Item: item,
		Input: &dynamodb.DeleteItemInput{
			Key:       key,
			TableName: &m.tableName,
		},
		OptimisticLockingEnabled: m.deleteVersion != nil,
	}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.OptimisticLockingEnabled {
		if m.deleteVersion == nil {
			return nil, fmt.Errorf("OptimisticLockingEnabled must be false because item does not implement HasVersion")
		}

		c, err := m.deleteVersion(item, value)
		if err != nil {
			return nil, fmt.Errorf("create version condition expression error: %w", err)
		}
		if c.IsSet() {
			opts.Condition = expr.And(opts.Condition, c)
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

	return m.client.DeleteItem(ctx, opts.Input)
}

// DeleteOpts provides customisation to the [Mapper.Delete] operation.
//
// DeleteOpts.Input is guaranteed to exist when passed into the first modifier.
// DeleteOpts.Item is the original reference item. Changes to DeleteOpts.Item don't automatically update DeleteOpts.Input.
// Changes to DeleteOpts.Condition will affect the final DeleteOpts.Input.
//
// OptimisticLockingEnabled will be true initially if Mapper detects that the item supports HasVersion.
//
// Finalizer is the modifier applied to the final DeleteOpts.Input right before executing the DynamoDB DeleteItem request.
type DeleteOpts[T interface{}] struct {
	Item                     T
	Input                    *dynamodb.DeleteItemInput
	Condition                *expression.ConditionBuilder
	OptimisticLockingEnabled bool
	Finalizer                func(*dynamodb.DeleteItemInput)
}

// WithTableName changes the table.
func (o *DeleteOpts[T]) WithTableName(tableName string) *DeleteOpts[T] {
	o.Input.TableName = &tableName
	return o
}

// DisableOptimisticLocking explicitly disables [HasVersion] logic.
func (o *DeleteOpts[T]) DisableOptimisticLocking() *DeleteOpts[T] {
	o.OptimisticLockingEnabled = false
	return o
}

// ReturnAllOldValues sets the [dynamodb.DeleteItemInput.ReturnValues] to ALL_OLD.
//
// [dynamodb.DeleteItemInput.ReturnValues] only support NONE or ALL_OLD.
func (o *DeleteOpts[T]) ReturnAllOldValues() *DeleteOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
	return o
}

// ReturnAllOldValuesOnConditionCheckFailure sets the [dynamodb.DeleteItemInput.ReturnValuesOnConditionCheckFailure] to ALL_OLD.
//
// [dynamodb.DeleteItemInput.ReturnValuesOnConditionCheckFailure] supports NONE or ALL_OLD.
func (o *DeleteOpts[T]) ReturnAllOldValuesOnConditionCheckFailure() *DeleteOpts[T] {
	o.Input.ReturnValuesOnConditionCheckFailure = dynamodbtypes.ReturnValuesOnConditionCheckFailureAllOld
	return o
}
