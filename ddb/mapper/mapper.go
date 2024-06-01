package mapper

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// New creates a new Mapper instance after parsing a Model from struct T's tags.
//
// See NewModel on what struct tags are supported.
func New[T interface{}](client *dynamodb.Client, tableName string, optFns ...func(*MapOpts)) (*Mapper[T], error) {
	model, err := NewModel[T](tableName, optFns...)
	if err != nil {
		return nil, err
	}

	return &Mapper[T]{model, client}, nil
}

// Mapper contains a Model and a DynamoDB client.
type Mapper[T interface{}] struct {
	model  *Model[T]
	client *dynamodb.Client
}
