package mapper

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// Marshal is an alias to attributevalue.Marshal using the internal attributevalue.Encoder.
func (m Mapper[T]) Marshal(in T) (types.AttributeValue, error) {
	return m.encoder.Encode(in)
}

// MarshalMap is an alias to attributevalue.MarshalMap using the internal attributevalue.Encoder.
func (m Mapper[T]) MarshalMap(in T) (map[string]types.AttributeValue, error) {
	av, err := m.encoder.Encode(in)
	avm, ok := av.(*types.AttributeValueMemberM)
	if err == nil && av != nil && ok {
		return avm.Value, nil
	}

	return map[string]types.AttributeValue{}, err
}

// Unmarshal is an alias to attributevalue.Unmarshal using the internal attributevalue.Decoder.
func (m Mapper[T]) Unmarshal(av types.AttributeValue, out T) error {
	return m.decoder.Decode(av, out)
}

// UnmarshalMap is an alias to attributevalue.UnmarshalMap using the internal attributevalue.Decoder.
func (m Mapper[T]) UnmarshalMap(item map[string]types.AttributeValue, out T) error {
	return m.decoder.Decode(&types.AttributeValueMemberM{Value: item}, out)
}
