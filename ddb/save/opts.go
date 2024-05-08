package save

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/nguyengg/golambda/ddb/model"
	"github.com/nguyengg/golambda/ddb/timestamp"
)

// Opts provides customisation to the dynamodb.PutItemInput made with [github.com/nguyengg/golambda/ddb.Wrapper.Save].
//
// Opts.Input is guaranteed to exist when passed into the first modifier.
// Opts.Item is the original reference item. Changes to Opts.Item don't automatically update Opts.Input.
// Changes to Opts.Condition will affect the final Opts.Input.
type Opts struct {
	Item                          model.Item
	Input                         *dynamodb.PutItemInput
	Condition                     *expression.ConditionBuilder
	DisableOptimisticLocking      bool
	DisableAutoGenerateTimestamps timestamp.AutoGenerateFlag
}

// WithTableName changes the table name in Opts.Input.
func WithTableName(tableName string) func(*Opts) {
	return func(opts *Opts) {
		opts.Input.TableName = &tableName
	}
}

// DisableOptimisticLocking disables logic around [model.Versioned].
func DisableOptimisticLocking() func(*Opts) {
	return func(opts *Opts) {
		opts.DisableOptimisticLocking = true
	}
}

// DisableAutoGenerateTimestamps disables logic around [model.HasCreatedTimestamp] and [model.HasModifiedTimestamp].
func DisableAutoGenerateTimestamps(flag timestamp.AutoGenerateFlag) func(*Opts) {
	return func(opts *Opts) {
		opts.DisableAutoGenerateTimestamps = flag
	}
}

// ReturnAllOldValues sets the dynamodb.PutItemInput's ReturnValues to ALL_OLD.
//
// dynamodb.PutItemInput's ReturnValues only support NONE or ALL_OLD.
func ReturnAllOldValues(opts *Opts) {
	opts.Input.ReturnValues = types.ReturnValueAllOld
}
