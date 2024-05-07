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

// Update makes a DynamoDB UpdateItem request.
//
// At least one update expression must be given. This is because there's no easy way to scan through all attributes to
// see which are non-nil or non-zero, and then create the SET or REMOVE actions accordingly.
//
// optFns are applied before any logic regarding UpdateOpts.OptimisticLockingEnabled and
// UpdateOpts.AutoGenerateTimestampsEnabled take place to allow the disabling of those features. Use
// UpdateOpts.Finalizer if you need to apply additional modifiers right before executing the DynamoDB UpdateItem
// request.
func (m Mapper[T]) Update(ctx context.Context, item T, required func(*UpdateOpts[T]), optFns ...func(*UpdateOpts[T])) (*dynamodb.UpdateItemOutput, error) {
	value := reflect.ValueOf(item)
	key, err := m.getKey(item, value)
	if err != nil {
		return nil, fmt.Errorf("create UpdateItem's Key error: %w", err)
	}

	opts := &UpdateOpts[T]{
		Item: item,
		Input: &dynamodb.UpdateItemInput{
			Key:       key,
			TableName: &m.tableName,
		},
		OptimisticLockingEnabled:      m.updateVersion != nil,
		AutoGenerateTimestampsEnabled: m.updateTimestamps != nil,
	}
	required(opts)
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.OptimisticLockingEnabled {
		if m.updateVersion == nil {
			return nil, fmt.Errorf("OptimisticLockingEnabled must be false because item does not implement HasVersion")
		}

		update, cond, err := m.updateVersion(item, value, opts.Update)
		if err != nil {
			return nil, fmt.Errorf("create version condition expression error: %w", err)
		}
		opts.Update = update
		if cond.IsSet() {
			opts.Condition = expr.And(opts.Condition, cond)
		}
	}

	if opts.AutoGenerateTimestampsEnabled {
		if m.updateTimestamps == nil {
			return nil, fmt.Errorf("AutoGenerateTimestampsEnabled must be false because item does not implement HasTimestamps")
		}

		opts.Update, err = m.updateTimestamps(item, value, opts.Update)
		if err != nil {
			return nil, fmt.Errorf("create timestamp attributes error: %w", err)
		}
	}

	builder := expression.NewBuilder().WithUpdate(opts.Update)
	if opts.Condition != nil {
		builder = builder.WithCondition(*opts.Condition)
	}

	e, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build expressions error: %w", err)
	}
	opts.Input.ConditionExpression = e.Condition()
	opts.Input.ExpressionAttributeNames = e.Names()
	opts.Input.ExpressionAttributeValues = e.Values()
	opts.Input.UpdateExpression = e.Update()

	return m.client.UpdateItem(ctx, opts.Input)
}

// UpdateOpts provides customisation to the [Mapper.Update] operation.
//
// UpdateOpts.Input is guaranteed to exist when passed into the first modifier.
// UpdateOpts.Item is the original reference item. Changes to UpdateOpts.Item don't automatically update UpdateOpts.Input.
// Changes to UpdateOpts.Condition and UpdateOpts.Update will affect the final UpdateOpts.Input.
//
// OptimisticLockingEnabled and AutoGenerateTimestampsEnabled will be true initially if Mapper detects that the item
// supports HasVersion and HasTimestamps respectively.
//
// Finalizer is the modifier applied to the final UpdateOpts.Input right before executing the DynamoDB UpdateItem request.
type UpdateOpts[T interface{}] struct {
	Item                          T
	Input                         *dynamodb.UpdateItemInput
	Condition                     *expression.ConditionBuilder
	Update                        expression.UpdateBuilder
	OptimisticLockingEnabled      bool
	AutoGenerateTimestampsEnabled bool
	Finalizer                     func(*dynamodb.UpdateItemInput)
}

// WithTableName changes the table.
func (o *UpdateOpts[T]) WithTableName(tableName string) *UpdateOpts[T] {
	o.Input.TableName = &tableName
	return o
}

// DisableOptimisticLocking explicitly disables [HasVersion] logic.
func (o *UpdateOpts[T]) DisableOptimisticLocking() *UpdateOpts[T] {
	o.OptimisticLockingEnabled = false
	return o
}

// DisableAutoGenerateTimestamps explicitly disables [HasTimestamps] logic.
func (o *UpdateOpts[T]) DisableAutoGenerateTimestamps() *UpdateOpts[T] {
	o.AutoGenerateTimestampsEnabled = false
	return o
}

// ReturnAllOldValues sets the [dynamodb.UpdateItemInput.ReturnValues] to ALL_OLD.
func (o *UpdateOpts[T]) ReturnAllOldValues() *UpdateOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
	return o
}

// ReturnUpdatedOldValues sets the [dynamodb.UpdateItemInput.ReturnValues] to UPDATED_OLD.
func (o *UpdateOpts[T]) ReturnUpdatedOldValues() *UpdateOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueUpdatedOld
	return o
}

// ReturnAllNewValues sets the [dynamodb.UpdateItemInput.ReturnValues] to ALL_NEW.
func (o *UpdateOpts[T]) ReturnAllNewValues() *UpdateOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueAllNew
	return o
}

// ReturnUpdatedNewValues sets the [dynamodb.UpdateItemInput.ReturnValues] to UPDATED_NEW.
func (o *UpdateOpts[T]) ReturnUpdatedNewValues() *UpdateOpts[T] {
	o.Input.ReturnValues = dynamodbtypes.ReturnValueUpdatedNew
	return o
}

// ReturnAllOldValuesOnConditionCheckFailure sets the [dynamodb.UpdateItemInput.ReturnValuesOnConditionCheckFailure] to ALL_OLD.
//
// [dynamodb.UpdateItemInput.ReturnValuesOnConditionCheckFailure] supports NONE or ALL_OLD.
func (o *UpdateOpts[T]) ReturnAllOldValuesOnConditionCheckFailure() *UpdateOpts[T] {
	o.Input.ReturnValuesOnConditionCheckFailure = dynamodbtypes.ReturnValuesOnConditionCheckFailureAllOld
	return o
}
