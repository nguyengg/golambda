package v2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Get makes a dynamodb.GetItemInput request.
func (t Table[T]) Get(ctx context.Context, key string, opts ...func(*dynamodb.GetItemInput)) (v *T, output *dynamodb.GetItemOutput, err error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(t.TableName),
		Key:       map[string]types.AttributeValue{t.HashKeyName: &types.AttributeValueMemberS{Value: key}},
	}
	for _, f := range opts {
		f(input)
	}

	output, err = t.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: nil,
		Key:       nil,
	})
	if err != nil {
		return
	}

	v = new(T)
	err = t.Decoder.Decode(&types.AttributeValueMemberM{Value: output.Item}, v)
	return
}
