package ddb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/condition"
)

// A Savable provides convenient method Save to save the item to database.
type Savable interface {
	// Returns the table name that is used in dynamodb.PutItemInput.
	GetTableName() string
	// Returns the dynamodb map representing this Savable.
	Marshal() (map[string]dynamodbtypes.AttributeValue, error)
}

// Marshals the content of Savable into a map of attribute values.
func Marshal(s Savable) (map[string]dynamodbtypes.AttributeValue, error) {
	return attributevalue.MarshalMap(&s)
}

// A panicky variant of Marshal.
func MustMarshal(s Savable) map[string]dynamodbtypes.AttributeValue {
	if m, err := Marshal(s); err != nil {
		panic(err)
	} else {
		return m
	}
}

// Input is guaranteed to exist when passed into the first modifier.
type SaveOpts struct {
	Input     *dynamodb.PutItemInput
	Condition *expression.ConditionBuilder
}

// Save uses DynamoDB PutItem to save the Savable instance into database.
// To use UpdateItem, look into the Update methods instead.
func Save(ctx context.Context, s Savable, svc *dynamodb.Client, modifiers ...func(*SaveOpts)) (*dynamodb.PutItemOutput, error) {
	item, err := s.Marshal()
	if err != nil {
		return nil, err
	}

	saveOpts := &SaveOpts{
		Input: &dynamodb.PutItemInput{
			Item:      item,
			TableName: aws.String(s.GetTableName()),
		},
	}
	for _, modifier := range modifiers {
		modifier(saveOpts)
	}

	if err := saveOpts.apply(); err != nil {
		return nil, err
	}

	return svc.PutItem(ctx, saveOpts.Input)
}

// SaveConditionAttributeExists adds a condition that requires the attribute to exist prior to the call.
func SaveConditionAttributeExists(name string) func(opts *SaveOpts) {
	return func(opts *SaveOpts) {
		opts.And(expression.AttributeExists(expression.Name(name)))
	}
}

// SaveReturnAllOldValues sets the dynamodb.PutItemInput's ReturnValues to ALL_OLD.
// Because Save uses DynamoDB PutItem, the ReturnValues only support NONE or ALL_OLD.
func SaveReturnAllOldValues(opts *SaveOpts) {
	opts.Input.ReturnValues = dynamodbtypes.ReturnValueAllOld
}

// Batch save several items.
// Two callbacks are given. The first one is passed each item returned from DynamoDB, if the callback returns false, the
// method will stop and return immediately. The second is passed two key slices: the remaining keys and the unprocessed
// keys, either of which can be empty. The second callback must return the next slice of keys to be loaded. Use
// BatchSaveRetryUnprocessed for the default keys callback.
func BatchSave(
	ctx context.Context,
	svc *dynamodb.Client,
	items []Savable,
	callback func(remaining, unprocessed []dynamodbtypes.WriteRequest) []dynamodbtypes.WriteRequest) error {

	var tableName string
	writeRequests := make([]dynamodbtypes.WriteRequest, 0)
	for i, s := range items {
		item, err := s.Marshal()
		if err != nil {
			return err
		}

		switch tableName {
		case "":
			tableName = s.GetTableName()
		case s.GetTableName():
		default:
			return fmt.Errorf("item at index %d has different table name (%s) instead of %s", i, s.GetTableName(), tableName)
		}

		writeRequests = append(writeRequests, dynamodbtypes.WriteRequest{
			PutRequest: &dynamodbtypes.PutRequest{Item: item},
		})
	}

	for n := len(writeRequests); n != 0; n = len(writeRequests) {
		if n > 25 {
			n = 25
		}

		output, err := svc.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{RequestItems: map[string][]dynamodbtypes.WriteRequest{tableName: writeRequests[0:n]}})
		if err != nil {
			return err
		}

		writeRequests = writeRequests[n:]
		unprocessed, ok := output.UnprocessedItems[tableName]
		if ok {
			writeRequests = callback(writeRequests, unprocessed)
			continue
		}
		writeRequests = callback(writeRequests, emptyUnprocessedWriteRequests)
	}

	return nil
}

// By default, unprocessed write requests will be appended to the remaining slice for retry.
func BatchSaveRetryUnprocessed(remaining, unprocessed []dynamodbtypes.WriteRequest) []dynamodbtypes.WriteRequest {
	return append(remaining, unprocessed...)
}

var emptyUnprocessedWriteRequests []dynamodbtypes.WriteRequest

// See condition.And. Return itself for chaining.
func (opts *SaveOpts) And(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *SaveOpts {
	opts.Condition = condition.And(opts.Condition, right, other...)
	return opts
}

// See condition.And. Return itself for chaining.
func (opts *SaveOpts) Or(right expression.ConditionBuilder, other ...expression.ConditionBuilder) *SaveOpts {
	opts.Condition = condition.Or(opts.Condition, right, other...)
	return opts
}

func (opts SaveOpts) apply() error {
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
