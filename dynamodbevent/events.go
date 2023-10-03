package dynamodbevent

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// StreamToDynamoDBAttributeValue converts a DynamoDB Stream event attribute to an equivalent DynamoDB attribute.
// TODO replace recursive implementation.
func StreamToDynamoDBAttributeValue(av events.DynamoDBAttributeValue) dynamodbtypes.AttributeValue {
	switch av.DataType() {
	case events.DataTypeBinary:
		return &dynamodbtypes.AttributeValueMemberB{Value: av.Binary()}
	case events.DataTypeBoolean:
		return &dynamodbtypes.AttributeValueMemberBOOL{Value: av.Boolean()}
	case events.DataTypeBinarySet:
		return &dynamodbtypes.AttributeValueMemberBS{Value: av.BinarySet()}
	case events.DataTypeList:
		l := av.List()
		value := make([]dynamodbtypes.AttributeValue, len(l))
		for i, v := range l {
			value[i] = StreamToDynamoDBAttributeValue(v)
		}
		return &dynamodbtypes.AttributeValueMemberL{Value: value}
	case events.DataTypeMap:
		value := make(map[string]dynamodbtypes.AttributeValue)
		for k, v := range av.Map() {
			value[k] = StreamToDynamoDBAttributeValue(v)
		}
		return &dynamodbtypes.AttributeValueMemberM{Value: value}
	case events.DataTypeNumber:
		return &dynamodbtypes.AttributeValueMemberN{Value: av.Number()}
	case events.DataTypeNumberSet:
		return &dynamodbtypes.AttributeValueMemberNS{Value: av.NumberSet()}
	case events.DataTypeNull:
		return &dynamodbtypes.AttributeValueMemberNULL{Value: av.IsNull()}
	case events.DataTypeString:
		return &dynamodbtypes.AttributeValueMemberS{Value: av.String()}
	case events.DataTypeStringSet:
		return &dynamodbtypes.AttributeValueMemberSS{Value: av.StringSet()}
	default:
		panic(UnsupportedDynamoDBTypeError{DataType: av.DataType()})
	}
}

// StreamToDynamoDBItem uses StreamToDynamoDBAttributeValue to convert an item from a DynamoDB Stream event to an item in
// DynamoDB.
func StreamToDynamoDBItem(item map[string]events.DynamoDBAttributeValue) map[string]dynamodbtypes.AttributeValue {
	res := make(map[string]dynamodbtypes.AttributeValue)
	for k, v := range item {
		res[k] = StreamToDynamoDBAttributeValue(v)
	}
	return res
}

type UnsupportedDynamoDBTypeError struct {
	DataType events.DynamoDBDataType
}

func (e UnsupportedDynamoDBTypeError) Error() string {
	return fmt.Sprintf("unsupported DynamoDB attribute type, %v", e.DataType)
}
