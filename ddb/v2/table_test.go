package v2

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
		Key string `hashkey:"myKey"`
	}
	got, err := New[Test](nil, "myTable")
	assert.NoError(t, err)
	assert.Equal(t, "myTable", got.TableName)
	assert.Equal(t, "myKey", got.HashKeyName)

	item := Test{Key: "myKeyValue"}
	av, err := got.Key(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{"myKey": &dynamodbtypes.AttributeValueMemberS{Value: "myKeyValue"}}, av)
}

func TestNew_ValidCompositeKey(t *testing.T) {
	type Test struct {
		Hash  int    `hashkey:"hash"`
		Range []byte `sortkey:"range"`
	}
	got, err := New[Test](nil, "myTable")
	assert.NoError(t, err)
	assert.Equal(t, "myTable", got.TableName)
	assert.Equal(t, "hash", got.HashKeyName)
	assert.Equal(t, "range", got.SortKeyName)

	item := Test{
		Hash:  1234,
		Range: []byte("hello, world!"),
	}
	av, err := got.Key(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	assert.Equal(t, av, map[string]dynamodbtypes.AttributeValue{
		"hash":  &dynamodbtypes.AttributeValueMemberN{Value: "1234"},
		"range": &dynamodbtypes.AttributeValueMemberB{Value: []byte("hello, world!")},
	})
}

func TestNew_NoHashKey(t *testing.T) {
	type Test struct {
		Key string
	}
	_, err := New[Test](nil, "")
	assert.ErrorContains(t, err, `no field with tag "hashkey" in type "Test"`)
}

func TestNew_DuplicateKeysAndAttributes(t *testing.T) {
	_, err := New[struct {
		KeyA string `hashkey:"a"`
		KeyB string `hashkey:"b"`
	}](nil, "")
	assert.ErrorContains(t, err, `multiple fields with tag "hashkey" found in type ""`)

	_, err = New[struct {
		KeyA string `sortkey:"a"`
		KeyB string `sortkey:"b"`
	}](nil, "")
	assert.ErrorContains(t, err, `multiple fields with tag "sortkey" found in type ""`)
}

func TestNew_ExpectVersionAttributeNotExists(t *testing.T) {
	type Test struct {
		Key     string `hashkey:"key"`
		Version int    `version:"version"`
	}

	table, err := New[Test](nil, "")
	assert.NoError(t, err)

	// with initial version being 0, ExpectVersion will add attribute_not_exists, and NextVersion will always increase by 1.
	item := Test{
		Key:     "123",
		Version: 0,
	}
	cond, err := table.ExpectVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	update, err := table.NextVersion(item, reflect.ValueOf(item))
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
		Key     string `hashkey:"key"`
		Version int    `version:"version"`
	}

	table, err := New[Test](nil, "")
	assert.NoError(t, err)

	// with initial version being non-0, ExpectVersion will have an Equal condition, and NextVersion will always increase by 1.
	item := Test{
		Key:     "123",
		Version: 123,
	}
	cond, err := table.ExpectVersion(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	update, err := table.NextVersion(item, reflect.ValueOf(item))
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
		Key          string                     `hashkey:"key"`
		CreatedTime  timestamp.EpochMillisecond `createdTime:"created"`
		ModifiedTime timestamp.EpochMillisecond `modifiedTime:"modified"`
	}

	table, err := New[Test](nil, "")
	assert.NoError(t, err)

	table.Now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = table.PutTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"},
		"modified": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"},
	}, m)

	update, err := table.UpdateTimestamps(item, reflect.ValueOf(item))
	assert.NoError(t, err)
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	assert.NoError(t, err)
	assert.Equal(t, "SET #0 = :0\n", *expr.Update())
	assert.Equal(t, map[string]string{"#0": "modified"}, expr.Names())
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{":0": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445000"}}, expr.Values())
}

func TestNew_TimestampsUnixTime(t *testing.T) {
	type Test struct {
		Key          string    `hashkey:"key"`
		CreatedTime  time.Time `createdTime:"created" dynamodbav:",unixtime"`
		ModifiedTime time.Time `modifiedTime:"modified" dynamodbav:",unixtime"`
	}

	table, err := New[Test](nil, "")
	assert.NoError(t, err)

	table.Now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = table.PutTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberN{Value: "1136239445"},
		"modified": &dynamodbtypes.AttributeValueMemberN{Value: "1136239445"},
	}, m)

	update, err := table.UpdateTimestamps(item, reflect.ValueOf(item))
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
		Key          string    `hashkey:"key"`
		CreatedTime  time.Time `createdTime:"created"`
		ModifiedTime time.Time `modifiedTime:"modified"`
	}

	table, err := New[Test](nil, "")
	assert.NoError(t, err)

	table.Now = func() time.Time {
		return time.Unix(1136239445, 0)
	}

	item := Test{
		Key: "123",
	}

	m := make(map[string]dynamodbtypes.AttributeValue)
	err = table.PutTimestamps(item, reflect.ValueOf(item), m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]dynamodbtypes.AttributeValue{
		"created":  &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T14:04:05-08:00"},
		"modified": &dynamodbtypes.AttributeValueMemberS{Value: "2006-01-02T14:04:05-08:00"},
	}, m)

	update, err := table.UpdateTimestamps(item, reflect.ValueOf(item))
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
