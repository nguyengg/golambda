package ddb

import (
	"encoding/json"
	"fmt"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
	"time"
)

// The string layout as well as DynamoDB string value of Timestamp.
const TimestampLayout = "2006-01-02T15:04:05.000Z"

// Timestamp is epoch second in UTC, formatted by RFC3339 but marshalled as a number.
type TTL time.Time

// Timestamp is epoch second in UTC, formatted by RFC3339, marshalled as a string.
type Timestamp time.Time

func TimestampFromTime(t time.Time) *Timestamp {
	ts := Timestamp(t)
	return &ts
}

func (ts Timestamp) String() string {
	return ts.Format(TimestampLayout)
}

func (ts Timestamp) Format(layout string) string {
	return time.Time(ts).Format(layout)
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(ts.Format(TimestampLayout))
	if err != nil {
		panic(err)
	}
	return data, nil
}

func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("invalid json")
	} else if t, err := time.Parse(TimestampLayout, value); err != nil {
		return fmt.Errorf("timestamp is not in %s format", TimestampLayout)
	} else {
		*ts = Timestamp(t)
	}
	return nil
}

func (ts *Timestamp) UnmarshalDynamoDBAttributeValue(av dynamodbtypes.AttributeValue) error {
	avS, ok := av.(*dynamodbtypes.AttributeValueMemberS)
	if !ok {
		return nil
	}

	s := avS.Value
	if s == "" {
		return nil
	}

	t, err := time.Parse(TimestampLayout, s)
	if err != nil {
		return err
	}

	*ts = Timestamp(t)
	return nil
}

func (ts Timestamp) MarshalDynamoDBAttributeValue() (dynamodbtypes.AttributeValue, error) {
	return &dynamodbtypes.AttributeValueMemberS{Value: time.Time(ts).Format(TimestampLayout)}, nil
}

func TTLFromTime(t time.Time) *TTL {
	ttl := TTL(t)
	return &ttl
}

func (ttl TTL) String() string {
	return time.Time(ttl).Format(time.RFC3339)
}

func (ttl TTL) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(time.Time(ttl).Format(time.RFC3339))
	if err != nil {
		panic(err)
	}
	return data, nil
}

func (ttl *TTL) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("invalid json")
	} else if t, err := time.Parse(time.RFC3339, value); err != nil {
		return fmt.Errorf("ttl is not in RFC3339 format")
	} else {
		*ttl = TTL(t)
	}
	return nil
}

func (ttl *TTL) UnmarshalDynamoDBAttributeValue(av dynamodbtypes.AttributeValue) error {
	avN, ok := av.(*dynamodbtypes.AttributeValueMemberN)
	if !ok {
		return nil
	}

	n := avN.Value
	if n == "" {
		return nil
	}

	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return err
	}

	*ttl = TTL(time.Unix(v, 0).UTC())
	return nil
}

func (ttl TTL) MarshalDynamoDBAttributeValue() (dynamodbtypes.AttributeValue, error) {
	return &dynamodbtypes.AttributeValueMemberN{Value: strconv.FormatInt(time.Time(ttl).Unix(), 10)}, nil
}
