package ddb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nguyengg/golambda/ddb/delete"
	"github.com/nguyengg/golambda/ddb/expr"
	"github.com/nguyengg/golambda/ddb/load"
	"github.com/nguyengg/golambda/ddb/model"
	. "github.com/nguyengg/golambda/ddb/mutil"
	"github.com/nguyengg/golambda/ddb/save"
	"github.com/nguyengg/golambda/ddb/timestamp"
	"github.com/nguyengg/golambda/ddb/update"
	"time"
)

// Wrapper wraps a dynamodb.Client and provides convenient methods on interfaces provided in this package.
type Wrapper struct {
	Client *dynamodb.Client
}

// Wrap creates a new wrapper from the specified client.
func Wrap(client *dynamodb.Client) *Wrapper {
	return &Wrapper{Client: client}
}

// Save makes a dynamodb.PutItemInput request.
//
// Return the original dynamodb.PutItemOutput output and any error.
//
// See [save.Opts] for the various options that can be passed into this method.
func (w Wrapper) Save(ctx context.Context, item model.Item, options ...func(*save.Opts)) (*dynamodb.PutItemOutput, error) {
	m, err := attributevalue.MarshalMap(item)
	if err != nil {
		return nil, fmt.Errorf("marshal map error: %w", err)
	}

	opts := &save.Opts{
		Item: item,
		Input: &dynamodb.PutItemInput{
			Item:      m,
			TableName: item.GetTableName(),
		},
	}

	for _, opt := range options {
		opt(opts)
	}

	isNew := false

	if !opts.DisableOptimisticLocking {
		switch v := item.(type) {
		case model.Versioned:
			n, oav, ok := First(v.GetVersion())
			if !ok {
				isNew = true
			}

			n, nav, ok := First(v.NextVersion())
			if !ok {
				return nil, fmt.Errorf("NextVersion returns empty map, you can disable Versioned logic with save.DisableOptimisticLocking")
			}

			if isNew {
				opts.Condition = expr.And(opts.Condition, expression.AttributeNotExists(expression.Name(n)))
			} else {
				opts.Condition = expr.And(opts.Condition, expression.Name(n).Equal(expression.Value(oav)))
			}
			opts.Input.Item[n] = nav
		}
	}

	now := time.Now()

	if opts.DisableAutoGenerateTimestamps&timestamp.CreatedTimestamp == 0 && isNew {
		switch v := item.(type) {
		case model.HasCreatedTimestamp:
			n, av, ok := First(v.UpdateCreatedTimestamp(now))
			if !ok {
				return nil, fmt.Errorf("UpdateCreatedTimestamp returns empty map, you can disable HasCreatedTimestamp logic with save.DisableAutoGenerateTimestamps(timestampe.CreatedTimestamp)")
			}

			opts.Input.Item[n] = av
		}
	}

	if opts.DisableAutoGenerateTimestamps&timestamp.ModifiedTimestamp == 0 {
		switch v := item.(type) {
		case model.HasModifiedTimestamp:
			n, av, ok := First(v.UpdateModifiedTimestamp(now))
			if !ok {
				return nil, fmt.Errorf("UpdateModifiedTimestamp returns empty map, you can disable HasModifiedTimestamp logic with save.DisableAutoGenerateTimestamps(timestampe.ModifiedTimestamp)")
			}

			opts.Input.Item[n] = av
		}
	}

	if opts.Condition != nil {
		e, err := expression.NewBuilder().WithCondition(*opts.Condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}
		opts.Input.ConditionExpression = e.Condition()
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ExpressionAttributeValues = e.Values()
	}

	return w.Client.PutItem(ctx, opts.Input)
}

// Load makes a dynamodb.GetItemInput request and loads the response into the specified modeling.
// The modeling must have its key filled out.
// Return the original dynamodb.GetItemOutput output and any error. If it doesn't exist in database, the modeling will not be
// modified, and len(output.Item) will be 0.
func (w Wrapper) Load(ctx context.Context, item model.Item, options ...func(opts *load.Opts)) (*dynamodb.GetItemOutput, error) {
	opts := &load.Opts{
		Item: item,
		Input: &dynamodb.GetItemInput{
			Key:       item.GetKey(),
			TableName: item.GetTableName(),
		},
	}
	for _, opt := range options {
		opt(opts)
	}
	if opts.Projection != nil {
		e, err := expression.NewBuilder().WithProjection(*opts.Projection).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ProjectionExpression = e.Projection()
	}

	output, err := w.Client.GetItem(ctx, opts.Input)
	if err != nil {
		return nil, err
	}
	if len(output.Item) != 0 {
		if err = attributevalue.UnmarshalMap(output.Item, item); err != nil {
			return nil, fmt.Errorf("unmarshal map error: %w", err)
		}
	}

	return output, err
}

// Delete makes a dynamodb.DeleteItemInput request.
//
// Return the original dynamodb.DeleteItemOutput output and any error.
//
// See [delete.Opts] for the various options that can be passed into this method.
func (w Wrapper) Delete(ctx context.Context, item model.Item, options ...func(*delete.Opts)) (*dynamodb.DeleteItemOutput, error) {
	opts := &delete.Opts{
		Item: item,
		Input: &dynamodb.DeleteItemInput{
			Key:       item.GetKey(),
			TableName: item.GetTableName(),
		},
	}

	for _, opt := range options {
		opt(opts)
	}

	if !opts.DisableOptimisticLocking {
		switch v := item.(type) {
		case model.Versioned:
			n, av, ok := First(v.GetVersion())
			if !ok {
				return nil, fmt.Errorf("GetVersion returns empty map, you can disable Versioned logic with delete.DisableOptimisticLocking")
			}

			opts.Condition = expr.And(opts.Condition, expression.Name(n).Equal(expression.Value(av)))
		}
	}

	if opts.Condition != nil {
		e, err := expression.NewBuilder().WithCondition(*opts.Condition).Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}
		opts.Input.ConditionExpression = e.Condition()
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ExpressionAttributeValues = e.Values()
	}

	return w.Client.DeleteItem(ctx, opts.Input)
}

// Update makes a dynamodb.UpdateItemInput request.
//
// At least one update expression must be given such as [update.SetOrRemove]. See [update.Opts] for more options.
// The item being passed in is only used for its [model.Item.GetKey] and [model.Item.GetTableName]; any other attributes
// that are set in the item must be explicitly updated with an opt. This is because there's no easy way to scan through
// (maybe with reflection) all attributes to see which are non-nil or non-zero, and then create the SET or REMOVE
// actions accordingly.
//
// Return the original dynamodb.UpdateItemOutput output and any error.
func (w Wrapper) Update(ctx context.Context, item model.Item, required func(*update.Opts), options ...func(*update.Opts)) (*dynamodb.UpdateItemOutput, error) {
	opts := &update.Opts{
		Item: item,
		Input: &dynamodb.UpdateItemInput{
			Key:       item.GetKey(),
			TableName: item.GetTableName(),
		},
	}

	required(opts)
	for _, opt := range options {
		opt(opts)
	}

	isNew := false

	if !opts.DisableOptimisticLocking {
		switch v := item.(type) {
		case model.Versioned:
			n, av, ok := First(v.GetVersion())
			if !ok {
				isNew = true
				opts.Condition = expr.And(opts.Condition, expression.AttributeNotExists(expression.Name(n)))
			} else {
				opts.Condition = expr.And(opts.Condition, expression.Name(n).Equal(expression.Value(av)))
			}

			n, av, ok = First(v.NextVersion())
			if !ok {
				return nil, fmt.Errorf("NextVersion returns empty map, you can disable Versioned logic with update.DisableOptimisticLocking")
			}

			opts.Update = expr.Set(opts.Update, expression.Name(n), expression.Value(av))
		}
	}

	now := time.Now()

	if opts.DisableAutoGenerateTimestamps&timestamp.CreatedTimestamp == 0 && isNew {
		switch v := item.(type) {
		case model.HasCreatedTimestamp:
			n, av, ok := First(v.UpdateCreatedTimestamp(now))
			if !ok {
				return nil, fmt.Errorf("UpdateCreatedTimestamp returns empty map, you can disable HasCreatedTimestamp logic with update.DisableAutoGenerateTimestamps(timestampe.CreatedTimestamp)")
			}

			opts.Update = expr.Set(opts.Update, expression.Name(n), expression.Value(av))
		}
	}

	if opts.DisableAutoGenerateTimestamps&timestamp.ModifiedTimestamp == 0 {
		switch v := item.(type) {
		case model.HasModifiedTimestamp:
			n, av, ok := First(v.UpdateModifiedTimestamp(now))
			if !ok {
				return nil, fmt.Errorf("UpdateModifiedTimestamp returns empty map, you can disable HasModifiedTimestamp logic with update.DisableAutoGenerateTimestamps(timestampe.ModifiedTimestamp)")
			}

			opts.Update = expr.Set(opts.Update, expression.Name(n), expression.Value(av))
		}
	}

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
		e, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("build expressions error: %w", err)
		}

		opts.Input.ConditionExpression = e.Condition()
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ExpressionAttributeValues = e.Values()
		opts.Input.UpdateExpression = e.Update()
	}

	return w.Client.UpdateItem(ctx, opts.Input)
}
