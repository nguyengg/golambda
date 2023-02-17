package ddb

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// A Versioned item is a Loadable and Savable with a numeric attribute that is used for optimistic locking.
type Versioned interface {
	// Returns the table name that is used in dynamodb.GetItemInput.
	GetTableName() string
	// Returns the key map that is used in dynamodb.GetItemInput and dynamodb.PutItemInput.
	GetKey() map[string]dynamodbtypes.AttributeValue
	// Returns the name of the attribute containing the numeric version for optimistic locking.
	GetVersionAttributeName() string
	// Returns the dynamodb map representing this Versioned instance.
	Marshal() (map[string]dynamodbtypes.AttributeValue, error)
}

// Save with the condition that the item doesn't exist in database.
func SaveVersionedWithNoExpectedVersion(
	ctx context.Context,
	v Versioned,
	svc *dynamodb.Client,
	modifiers ...func(*SaveOpts)) (*dynamodb.PutItemOutput, error) {

	return Save(ctx, v, svc, func(opts *SaveOpts) {
		// we only need to set attribute_not_exists(...) on either hash or range key, not both, hence the break.
		for name := range v.GetKey() {
			opts.And(expression.AttributeNotExists(expression.Name(name)))
			break
		}

		for _, modifier := range modifiers {
			modifier(opts)
		}
	})
}

// Save with an explicit expected version which implies the item exists in the database.
func SaveVersionedWithExpectedVersion(
	ctx context.Context,
	v Versioned,
	expectedVersion interface{},
	svc *dynamodb.Client,
	modifiers ...func(*SaveOpts)) (*dynamodb.PutItemOutput, error) {

	return Save(ctx, v, svc, func(opts *SaveOpts) {
		opts.And(expression.Name(v.GetVersionAttributeName()).Equal(expression.Value(expectedVersion)))

		for _, modifier := range modifiers {
			modifier(opts)
		}
	})
}

// Variant of SaveVersionedWithExpectedVersion that uses dynamodb.UpdateItem to prevent clobbering of null attributes.
func UpdateVersionedWithExpectedVersion(
	ctx context.Context,
	v Versioned,
	expectedVersion interface{},
	svc *dynamodb.Client,
	required func(*UpdateOpts),
	modifiers ...func(*UpdateOpts)) (*dynamodb.UpdateItemOutput, error) {

	return Update(ctx, v, svc, func(opts *UpdateOpts) {
		opts.And(expression.Name(v.GetVersionAttributeName()).Equal(expression.Value(expectedVersion)))

		required(opts)
		for _, modifier := range modifiers {
			modifier(opts)
		}
	})
}

// Delete with an explicit expected version.
func DeleteVersionedWithExpectedVersion(
	ctx context.Context,
	v Versioned,
	expectedVersion interface{},
	svc *dynamodb.Client,
	modifiers ...func(*DeleteOpts)) (*dynamodb.DeleteItemOutput, error) {

	return Delete(ctx, v, svc, func(opts *DeleteOpts) {
		opts.And(expression.Name(v.GetVersionAttributeName()).Equal(expression.Value(expectedVersion)))

		for _, modifier := range modifiers {
			modifier(opts)
		}
	})
}
