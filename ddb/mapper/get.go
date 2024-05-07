package mapper

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
)

// Get makes a dynamodb.GetItemInput request.
func (m Mapper[T]) Get(ctx context.Context, item T, opts ...func(*dynamodb.GetItemInput)) (v *T, output *dynamodb.GetItemOutput, err error) {
	key, err := m.getKey(item, reflect.ValueOf(item))
	if err != nil {
		return nil, nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(m.tableName),
		Key:       key,
	}
	for _, f := range opts {
		f(input)
	}

	output, err = m.client.GetItem(ctx, input)
	if err != nil {
		return
	}

	v = new(T)
	err = m.decoder.Decode(&types.AttributeValueMemberM{Value: output.Item}, v)
	return
}
