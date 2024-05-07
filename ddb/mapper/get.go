package mapper

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/expr"
	"reflect"
)

// Get makes a DynamoDB GetItem request.
//
// The returned item is nil if there is no matching item ([dynamodb.GetItemOutput.Item] is empty).
func (m Mapper[T]) Get(ctx context.Context, item T, optFns ...func(*GetOpts[T])) (*T, *dynamodb.GetItemOutput, error) {
	key, err := m.getKey(item, reflect.ValueOf(item))
	if err != nil {
		return nil, nil, fmt.Errorf("create GetItem's Key error: %w", err)
	}

	opts := &GetOpts[T]{
		Item: item,
		Input: &dynamodb.GetItemInput{
			Key:       key,
			TableName: &m.tableName,
		},
	}
	for _, fn := range optFns {
		fn(opts)
	}

	if opts.Projection != nil {
		e, err := expression.NewBuilder().WithProjection(*opts.Projection).Build()
		if err != nil {
			return nil, nil, fmt.Errorf("build projection expression error: %w", err)
		}
		opts.Input.ExpressionAttributeNames = e.Names()
		opts.Input.ProjectionExpression = e.Projection()
	}

	output, err := m.client.GetItem(ctx, opts.Input)
	if err != nil || len(output.Item) == 0 {
		return nil, output, err
	}

	res := new(T)
	err = m.decoder.Decode(&types.AttributeValueMemberM{Value: output.Item}, res)
	if err != nil {
		return nil, output, fmt.Errorf("unmarshal item error: %w", err)
	}

	return res, output, nil
}

// GetOpts provides customisation to the Mapper.Get operation.
//
// GetOpts.Input is guaranteed to exist when passed into the first modifier.
// GetOpts.Item is the original reference item. Changes to GetOpts.Item don't automatically update GetOpts.Input.
// Changes to GetOpts.Projection will affect the final PutOpts.Input.
type GetOpts[T interface{}] struct {
	Item       T
	Input      *dynamodb.GetItemInput
	Projection *expression.ProjectionBuilder
}

// WithTableName changes the table.
func (o *GetOpts[T]) WithTableName(tableName string) *GetOpts[T] {
	o.Input.TableName = &tableName
	return o
}

// WithProjection changes the projection expression.
func (o *GetOpts[T]) WithProjection(name string, names ...string) *GetOpts[T] {
	switch len(names) {
	case 0:
		o.Projection = expr.AddNames(o.Projection, expression.Name(name))
	case 1:
		o.Projection = expr.AddNames(o.Projection, expression.Name(name), expression.Name(names[0]))
	case 2:
		o.Projection = expr.AddNames(o.Projection, expression.Name(name), expression.Name(names[0]), expression.Name(names[1]))
	default:
		nameBuilders := make([]expression.NameBuilder, len(names))
		for i, n := range names {
			nameBuilders[i] = expression.Name(n)
		}
		o.Projection = expr.AddNames(o.Projection, expression.Name(name), nameBuilders...)
	}
	return o
}
