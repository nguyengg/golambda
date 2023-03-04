package ddb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/condition"
	"github.com/nguyengg/golambda/ddb/update"
)

// An Updatable provides convenient method Update to update the item in database.
//
// This interface implements Savable and Deletable as well.
type Updatable interface {
	// GetTableName returns the table name that is used in dynamodb.UpdateItemInput.
	GetTableName() string
	// GetKey returns the key map that is used in dynamodb.UpdateItemInput.
	GetKey() map[string]dynamodbtypes.AttributeValue
	// Marshal returns the dynamodb map representing this Updatable.
	Marshal() (map[string]dynamodbtypes.AttributeValue, error)
}

// UpdateOpts contains the various constructs that are used during the process of customising an UpdateItem request.
//
// Update will guarantee that Input exists when passed into the first modifier
type UpdateOpts struct {
	Input     *dynamodb.UpdateItemInput
	Condition *expression.ConditionBuilder
	Update    *expression.UpdateBuilder
}

// Update use DynamoDB UpdateItem to write the Updatable instance to database.
// At least one function to modify UpdateOpts must be provided since the default dynamodb.UpdateItemInput has no update
// expression.
func Update(
	ctx context.Context, u Updatable, svc *dynamodb.Client,
	required func(*UpdateOpts), modifiers ...func(*UpdateOpts)) (*dynamodb.UpdateItemOutput, error) {

	updateOpts := &UpdateOpts{
		Input: &dynamodb.UpdateItemInput{
			Key:       u.GetKey(),
			TableName: aws.String(u.GetTableName()),
		},
	}

	required(updateOpts)
	for _, modifier := range modifiers {
		modifier(updateOpts)
	}

	if err := updateOpts.apply(); err != nil {
		return nil, err
	}

	return svc.UpdateItem(ctx, updateOpts.Input)
}

// UpdateConditionAttributeExists adds a condition that requires the attribute to exist prior to the call.
func UpdateConditionAttributeExists(name string) func(opts *UpdateOpts) {
	return func(opts *UpdateOpts) {
		opts.And(expression.AttributeExists(expression.Name(name)))
	}
}

// UpdateConditionItemExists is a specialization of UpdateConditionAttributeExists that uses the key name to require
// the item to exist prior to the call.
func UpdateConditionItemExists(opts *UpdateOpts) {
	for key := range opts.Input.Key {
		opts.And(expression.AttributeExists(expression.Name(key)))
		break
	}
}

// UpdateReturnAllOldValues sets the dynamodb.UpdateItemInput's ReturnValues to ALL_OLD.
func UpdateReturnAllOldValues(opts *UpdateOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
}

// UpdateReturnUpdatedOldValues sets the dynamodb.UpdateItemInput's ReturnValues to UPDATED_OLD.
func UpdateReturnUpdatedOldValues(opts *UpdateOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueUpdatedOld
}

// UpdateReturnAllNewValues sets the dynamodb.UpdateItemInput's ReturnValues to ALL_NEW.
func UpdateReturnAllNewValues(opts *UpdateOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueAllNew
}

// UpdateReturnUpdatedNewValues sets the dynamodb.UpdateItemInput's ReturnValues to UPDATED_NEW.
func UpdateReturnUpdatedNewValues(opts *UpdateOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueUpdatedNew
}

// And adds a condition.And condition. Return itself for chaining.
func (opts *UpdateOpts) And(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *UpdateOpts {
	opts.Condition = condition.And(opts.Condition, right, other...)
	return opts
}

// Or adds a condition.Or condition. Return itself for chaining.
func (opts *UpdateOpts) Or(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *UpdateOpts {
	opts.Condition = condition.Or(opts.Condition, right, other...)
	return opts
}

// Add adds an update.Add statement. Return itself for chaining.
func (opts *UpdateOpts) Add(name expression.NameBuilder, value expression.ValueBuilder) *UpdateOpts {
	opts.Update = update.Add(opts.Update, name, value)
	return opts
}

// Set adds an update.Set statement. Return itself for chaining.
func (opts *UpdateOpts) Set(name expression.NameBuilder, value expression.ValueBuilder) *UpdateOpts {
	opts.Update = update.Set(opts.Update, name, value)
	return opts
}

// Delete adds an update.Delete statement. Return itself for chaining.
func (opts *UpdateOpts) Delete(name expression.NameBuilder, value expression.ValueBuilder) *UpdateOpts {
	opts.Update = update.Delete(opts.Update, name, value)
	return opts
}

// Remove adds an update.Remove statement. Return itself for chaining.
func (opts *UpdateOpts) Remove(name expression.NameBuilder) *UpdateOpts {
	opts.Update = update.Remove(opts.Update, name)
	return opts
}

// SetOrRemove adds either Set or Remove statement for the attribute specified by the name parameter.
//
// If set is true, UpdateOpts.Update will always receive an update.Set to set the name to the value.
//
// If set is false, only when remove is true then an update.Remove will be added. Otherwise, nothing is done.
//
// | set   | remove | effect
// | true  | *      | update.Set
// | false | true   | update.Remove
// | false | false  | no-op
//
// This is useful for distinguishing between PUT/POST (remove=true) that replaces attributes vs. PATCH (remove=false)
// that will only update attributes that are non-nil.
func (opts *UpdateOpts) SetOrRemove(set, remove bool, name expression.NameBuilder, value func() expression.ValueBuilder) *UpdateOpts {
	if set {
		opts.Update = update.Set(opts.Update, name, value())
		return opts
	}

	if remove {
		opts.Update = update.Remove(opts.Update, name)
		return opts
	}

	return opts
}

func (opts *UpdateOpts) apply() error {
	builder := expression.NewBuilder()
	hasExpressions := false

	if opts.Condition != nil {
		hasExpressions = true
		builder = builder.WithCondition(*opts.Condition)
	}
	if opts.Update != nil {
		hasExpressions = true
		builder = builder.WithUpdate(*opts.Update)
	}

	if hasExpressions {
		expr, err := builder.Build()
		if err != nil {
			return err
		}

		opts.Input.ConditionExpression = expr.Condition()
		opts.Input.ExpressionAttributeNames = expr.Names()
		opts.Input.ExpressionAttributeValues = expr.Values()
		opts.Input.UpdateExpression = expr.Update()
	}

	return nil
}
