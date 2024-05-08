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

// SetOrRemove adds either Set or Remove action to the update expression.
//
// If set is true, a SET action will be added.
// If set is false, only when remove is true then a REMOVE action will be added.
//
// | set   | remove | action
// | true  | *      | SET
// | false | true   | REMOVE
// | false | false  | no-op
//
// This is useful for distinguishing between PUT/POST (remove=true) that replaces attributes with clobbering behaviour
// vs. PATCH (remove=false) that will only update attributes that are non-nil. An example is given:
//
//	func PUT(body Request) {
//		// because it's a PUT request, if notes is empty, instead of writing empty string to database, we'll remove it.
//		wrapper.Update(..., SetOrRemove(true, true, "notes", body.Notes))
//	}
//
//	func PATCH(body Request) {
//		// only when notes is non-empty that we'll update it. an empty notes just means caller didn't try to update it.
//		wrapper.Update(..., opts.SetOrRemove(body.Notes != "", false, "notes", body.Notes))
//	}
//
//	func Update(method string, body Request) {
//		// an attempt to unify the methods may look like this.
//		wrapper.Update(..., opts.SetOrRemove(body.Notes != "", method != "PATCH", "notes", body.Notes))
//	}
//
// The name and value will be wrapped with an `expression.Name` and `expression.Value` so don't bother wrapping them
// ahead of time.
func (o *UpdateOpts[T]) SetOrRemove(set, remove bool, name string, value interface{}) *UpdateOpts[T] {
	if set {
		o.Update.Set(expression.Name(name), expression.Value(value))
		return o
	}

	if remove {
		o.Update.Remove(expression.Name(name))
		return o
	}

	return o
}

// SetOrRemoveFunc is a variant of SetOrRemove that accepts a function that is only called to return the value to be set
// if the `set` parameter is true.
//
// Use this if you may need to avoid dereferencing a null pointer. Consider using the various aws.ToString methods which
// is nil-pointer-safe.
func (o *UpdateOpts[T]) SetOrRemoveFunc(set, remove bool, name string, value func() interface{}) *UpdateOpts[T] {
	if set {
		o.Update.Set(expression.Name(name), expression.Value(value()))
		return o
	}

	if remove {
		o.Update.Remove(expression.Name(name))
		return o
	}

	return o
}

// SetOrRemoveStringPointer is a specialization of SetOrRemove for string pointer value.
//
// If ptr is a nil pointer, no action is taken. If ptr dereferences to an empty string, a REMOVE action is used.
// A non-empty string otherwise will result in a SET action.
func (o *UpdateOpts[T]) SetOrRemoveStringPointer(name string, ptr *string) *UpdateOpts[T] {
	if ptr == nil {
		return o
	}

	if v := *ptr; v != "" {
		o.Update.Set(expression.Name(name), expression.Value(v))
		return o
	}

	o.Update.Remove(expression.Name(name))
	return o
}

func (o *UpdateOpts[T]) Set(name string, value interface{}) *UpdateOpts[T] {
	o.Update.Set(expression.Name(name), expression.Value(value))
	return o
}

func (o *UpdateOpts[T]) Remove(name string) *UpdateOpts[T] {
	o.Update.Remove(expression.Name(name))
	return o
}

func (o *UpdateOpts[T]) Add(name string, value interface{}) *UpdateOpts[T] {
	o.Update.Add(expression.Name(name), expression.Value(value))
	return o
}

func (o *UpdateOpts[T]) Delete(name string, value interface{}) *UpdateOpts[T] {
	o.Update.Delete(expression.Name(name), expression.Value(value))
	return o
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
