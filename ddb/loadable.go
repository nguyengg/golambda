package ddb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/projection"
)

// A Loadable provides convenient method Load to load the item from database.
type Loadable interface {
	// Returns the table name that is used in dynamodb.GetItemInput.
	GetTableName() string
	// Returns the key map that is used in dynamodb.GetItemInput.
	GetKey() map[string]dynamodbtypes.AttributeValue
	// Loads the contents of the given dynamodb map into this Loadable.
	Unmarshal(item map[string]dynamodbtypes.AttributeValue) error
}

// Unmarshal the content of the map of attribute values into the Loadable.
func Unmarshal(l Loadable, item map[string]dynamodbtypes.AttributeValue) error {
	return attributevalue.UnmarshalMap(item, l)
}

// A panicky variant of Unmarshal.
func MustUnmarshal(l Loadable, item map[string]dynamodbtypes.AttributeValue) {
	if err := Unmarshal(l, item); err != nil {
		panic(err)
	}
}

// Input is guaranteed to exist when passed into the first modifier.
type LoadOpts struct {
	Input      *dynamodb.GetItemInput
	Projection *expression.ProjectionBuilder
}

// Loads the item.
func Load(ctx context.Context, l Loadable, svc *dynamodb.Client, modifiers ...func(*LoadOpts)) (*dynamodb.GetItemOutput, error) {
	loadOpts := &LoadOpts{
		Input: &dynamodb.GetItemInput{
			Key:       l.GetKey(),
			TableName: aws.String(l.GetTableName()),
		},
	}
	for _, modifier := range modifiers {
		modifier(loadOpts)
	}

	if err := loadOpts.apply(); err != nil {
		return nil, err
	}

	output, err := svc.GetItem(ctx, loadOpts.Input)
	if err == nil {
		err = l.Unmarshal(output.Item)
	}
	return output, err
}

// Batch load several items.
// Two callbacks are given. The first one is passed each item returned from DynamoDB, if the callback returns a non-nil
// error, the method will stop and return that error immediately. The second is passed two key slices: the remaining
// keys and the unprocessed keys, either of which can be empty. The second callback must return the next slice of keys
// to be loaded. Use BatchLoadRetryUnprocessed for the default keys callback.
func BatchLoad(
	ctx context.Context,
	svc *dynamodb.Client,
	items []Loadable,
	itemCallback func(item map[string]dynamodbtypes.AttributeValue) error,
	keysCallback func(remaining, unprocessed []map[string]dynamodbtypes.AttributeValue) []map[string]dynamodbtypes.AttributeValue) error {

	var tableName string
	keys := make([]map[string]dynamodbtypes.AttributeValue, len(items))
	for i, item := range items {
		keys[i] = item.GetKey()

		switch tableName {
		case "":
			tableName = item.GetTableName()
		case item.GetTableName():
		default:
			return fmt.Errorf("item at index %d has different table name (%s) instead of %s", i, item.GetTableName(), tableName)
		}
	}

	for n := len(keys); n != 0; n = len(keys) {
		if n > 100 {
			n = 100
		}

		output, err := svc.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{RequestItems: map[string]dynamodbtypes.KeysAndAttributes{tableName: {Keys: keys[:n]}}})
		if err != nil {
			return err
		}

		for _, item := range output.Responses[tableName] {
			itemCallback(item)
		}

		keys = keys[n:]
		unprocessed, ok := output.UnprocessedKeys[tableName]
		if ok {
			keys = keysCallback(keys, unprocessed.Keys)
			continue
		}
		keys = keysCallback(keys, emptyUnprocessedKeys)
	}

	return nil
}

// By default, unprocessed keys will be appended to the remaining slice for retry.
func BatchLoadRetryUnprocessed(remaining, unprocessed []map[string]dynamodbtypes.AttributeValue) []map[string]dynamodbtypes.AttributeValue {
	return append(remaining, unprocessed...)
}

var emptyUnprocessedKeys []map[string]dynamodbtypes.AttributeValue

// Checks whether the item exists or not.
func Exists(ctx context.Context, l Loadable, svc *dynamodb.Client) (bool, error) {
	loadOpts := &LoadOpts{
		Input: &dynamodb.GetItemInput{
			Key:       l.GetKey(),
			TableName: aws.String(l.GetTableName()),
		},
	}
	for k := range l.GetKey() {
		WithProjection(k)(loadOpts)
		break
	}

	output, err := svc.GetItem(ctx, loadOpts.Input)
	if err != nil {
		return false, err
	}

	return len(output.Item) != 0, nil
}

func WithProjection(name string, names ...string) func(*LoadOpts) {
	return func(opts *LoadOpts) {
		nameBuilders := make([]expression.NameBuilder, len(names))
		for i, n := range names {
			nameBuilders[i] = expression.Name(n)
		}
		opts.AddNames(expression.Name(name), nameBuilders...)
	}
}

// See projection.AddNames.
func (opts *LoadOpts) AddNames(nameBuilder expression.NameBuilder, nameBuilders ...expression.NameBuilder) *LoadOpts {
	opts.Projection = projection.AddNames(opts.Projection, nameBuilder, nameBuilders...)
	return opts
}

// Apply the LoadOpts.Projection to LoadOpts.Input.
func (opts LoadOpts) apply() error {
	if opts.Projection != nil {
		expr, err := expression.NewBuilder().WithProjection(*opts.Projection).Build()
		if err != nil {
			return err
		}
		opts.Input.ExpressionAttributeNames = expr.Names()
		opts.Input.ProjectionExpression = expr.Projection()
	}

	return nil
}
