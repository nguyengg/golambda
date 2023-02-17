package ddb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/condition"
)

// A Deletable provides convenient method Delete to delete the item from database.
type Deletable interface {
	// Returns the table name that is used in dynamodb.DeleteItemInput.
	GetTableName() string
	// Returns the key map that is used in dynamodb.DeleteItemInput.
	GetKey() map[string]dynamodbtypes.AttributeValue
}

// Input is guaranteed to exist when passed into the first modifier.
type DeleteOpts struct {
	Input     *dynamodb.DeleteItemInput
	Condition *expression.ConditionBuilder
}

// Delete uses DynamoDB DeleteItem to remove the Deletable instance from database.
func Delete(ctx context.Context, d Deletable, svc *dynamodb.Client, opts ...func(*DeleteOpts)) (*dynamodb.DeleteItemOutput, error) {
	deleteOpts := &DeleteOpts{
		Input: &dynamodb.DeleteItemInput{
			Key:       d.GetKey(),
			TableName: aws.String(d.GetTableName()),
		},
	}
	for _, o := range opts {
		o(deleteOpts)
	}

	if err := deleteOpts.apply(); err != nil {
		return nil, err
	}

	return svc.DeleteItem(ctx, deleteOpts.Input)
}

// DeleteConditionAttributeExists adds a condition that requires the attribute to exist prior to the call.
func DeleteConditionAttributeExists(name string) func(opts *DeleteOpts) {
	return func(opts *DeleteOpts) {
		opts.And(expression.AttributeExists(expression.Name(name)))
	}
}

// DeleteConditionItemExists is a specialization of DeleteConditionAttributeExists that uses the key name to require
// the item to exist prior to the call.
func DeleteConditionItemExists(opts *DeleteOpts) {
	for key := range opts.Input.Key {
		opts.And(expression.AttributeExists(expression.Name(key)))
		break
	}
}

// DeleteReturnAllOldValues sets the dynamodb.DeleteItemInput's ReturnValues to ALL_OLD.
// Because Delete uses DynamoDB DeleteItem, the ReturnValues only support NONE or ALL_OLD.
func DeleteReturnAllOldValues(opts *DeleteOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
}

// See condition.And. Return itself for chaining.
func (opts *DeleteOpts) And(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *DeleteOpts {
	opts.Condition = condition.And(opts.Condition, right, other...)
	return opts
}

// See condition.And. Return itself for chaining.
func (opts *DeleteOpts) Or(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *DeleteOpts {
	opts.Condition = condition.Or(opts.Condition, right, other...)
	return opts
}

func (opts DeleteOpts) apply() error {
	if opts.Condition != nil {
		expr, err := expression.NewBuilder().WithCondition(*opts.Condition).Build()
		if err != nil {
			return err
		}
		opts.Input.ConditionExpression = expr.Condition()
		opts.Input.ExpressionAttributeNames = expr.Names()
		opts.Input.ExpressionAttributeValues = expr.Values()
	}

	return nil
}
