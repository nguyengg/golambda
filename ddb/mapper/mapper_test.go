package mapper

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/timestamp"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestNew_ValidPrimaryKey(t *testing.T) {
	type Test struct {
		Key string `dynamodbav:"myKey,hashkey"`
	}
	mapper, err := New[Test](nil, "myTable")
	assert.NoError(t, err)

	item := Test{Key: "myKeyValue"}
	av, err := mapper.getKey(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{"myKey": &dynamodbtypes.AttributeValueMemberS{Value: "myKeyValue"}}, av)
}

func TestNew_ValidCompositeKey(t *testing.T) {
	type Test struct {
		Hash  int    `dynamodbav:"hash,hashkey"`
		Range []byte `dynamodbav:"range,sortkey"`
	}
	mapper, err := New[Test](nil, "myTable")
	assert.NoError(t, err)

	item := Test{
		Hash:  1234,
		Range: []byte("hello, world!"),
	}
	av, err := mapper.getKey(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	assert.Equal(t, av, map[string]dynamodbtypes.AttributeValue{
		"hash":  &dynamodbtypes.AttributeValueMemberN{Value: "1234"},
		"range": &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
	})
}

func TestNew_ExpectVersionAttributeNotExists(t *testing.T) {
	type Test struct {
		Key     string `dynamodbav:"key,hashkey"`
		Version int    `dynamodbav:"version,version"`
	}

	mapper, err := New[Test](nil, "", func(opts *MapOpts) {
		opts.MustHaveVersion = true
	})
	assert.NoError(t, err)

	// with initial version being 0, ExpectVersion will add attribute_not_exists, and NextVersion will always increase by 1.
	item := Test{
		Key:     "123",
		Version: 0,
	}
	cond, err := mapper.expectVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	update, err := mapper.nextVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)

	expr, err := expression.NewBuilder().
		WithCondition(cond).
		WithUpdate(update).
		Build()
	assert.NoError(t, err)
	assert.Equal(t, "attribute_not_exists (#0)", *expr.Condition())
	assert.Equal(t, "SET #1 = #1 + :0\n", *expr.Update())
	assert.Equal(t, map[string]string{
		"#0": "key",
		"#1": "version",
	}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		":0": &dynamodbtypes.AttributeValueMemberN{Value: "1"},
	}, expr.Values())
}

func TestNew_ExpectVersionIncrease(t *testing.T) {
	type Test struct {
		Key     string `dynamodbav:"key,hashkey"`
		Version int    `dynamodbav:"version,version"`
	}

	mapper, err := New[Test](nil, "", func(opts *MapOpts) {
		opts.MustHaveVersion = true
	})
	assert.NoError(t, err)

	// with initial version being non-0, ExpectVersion will have an Equal condition, and NextVersion will always increase by 1.
	item := Test{
		Key:     "123",
		Version: 123,
	}
	cond, err := mapper.expectVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	update, err := mapper.nextVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)

	expr, err := expression.NewBuilder().
		WithCondition(cond).
		WithUpdate(update).
		Build()
	assert.NoError(t, err)
	assert.Equal(t, "#0 = :0", *expr.Condition())
	assert.Equal(t, "SET #0 = #0 + :1\n", *expr.Update())
	assert.Equal(t, map[string]string{
		"#0": "version",
	}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		":0": &dynamodbtypes.AttributeValueMemberN{Value: "123"},
		":1": &dynamodbtypes.AttributeValueMemberN{Value: "1"},
	}, expr.Values())
}

func TestNew_TimestampsEpochMillisecond(t *testing.T) {
	type Test struct {
		Key          string                     `dynamodbav:"key,hashkey"`
		CreatedTime  timestamp.EpochMillisecond `dynamodbav:"created,createdTime"`
		ModifiedTime timestamp.EpochMillisecond `dynamodbav:"modified,modifiedTime"`
	}

	mapper, err := New[Test](nil, "", func(opts *MapOpts) {
		opts.MustHaveTimestamps = true
	})
	assert.NoError(t, err)

	mapper.now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = mapper.putTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"},
		"modified": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"},
	}, m)

	update, err := mapper.updateTimestamps(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	assert.NoError(t, err)
	assert.Equal(t, "SET #0 = :0\n", *expr.Update())
	assert.Equal(t, map[string]string{"#0": "modified"}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{":0": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"}}, expr.Values())
}

func TestNew_TimestampsUnixTime(t *testing.T) {
	type Test struct {
		Key          string    `dynamodbav:"key,hashkey"`
		CreatedTime  time.Time `dynamodbav:"created,unixtime,createdTime"`
		ModifiedTime time.Time `dynamodbav:"modified,unixtime,modifiedTime"`
	}

	mapper, err := New[Test](nil, "", func(opts *MapOpts) {
		opts.MustHaveTimestamps = true
	})
	assert.NoError(t, err)

	mapper.now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = mapper.putTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberN{Value: "1136239445"},
		"modified": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445"},
	}, m)

	update, err := mapper.updateTimestamps(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	assert.NoError(t, err)
	assert.Equal(t, "SET #0 = :0\n", *expr.Update())
	assert.Equal(t, map[string]string{
		"#0": "modified",
	}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		":0": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445"},
	}, expr.Values())
}

func TestNew_TimestampsRFC3339Nano(t *testing.T) {
	type Test struct {
		Key          string    `dynamodbav:"key,hashkey"`
		CreatedTime  time.Time `dynamodbav:"created,createdTime"`
		ModifiedTime time.Time `dynamodbav:"modified,modifiedTime"`
	}

	mapper, err := New[Test](nil, "", func(opts *MapOpts) {
		opts.MustHaveTimestamps = true
	})
	assert.NoError(t, err)

	mapper.now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = mapper.putTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T14:04:05-08:00"},
		"modified": &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T14:04:05-08:00"},
	}, m)

	update, err := mapper.updateTimestamps(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	assert.NoError(t, err)
	assert.Equal(t, "SET #0 = :0\n", *expr.Update())
	assert.Equal(t, map[string]string{
		"#0": "modified",
	}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		":0": &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T14:04:05-08:00"},
	}, expr.Values())
}
