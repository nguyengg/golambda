// Package model defines a few interfaces that structs representing an item in a DynamoDB table should implement to
// make use of [.ddb.Wrapper] features. The name of the package was carefully chosen to avoid collision with the various
// types packages that AWS SDK for Go v2 exports.

package model

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

// Item provides basic methods to retrieve the table name and key.
type Item interface {
	// GetTableName returns the table name.
	//
	// Method wasn't named TableName to avoid collision with possible field.
	GetTableName() *string
	// GetKey returns the key map that includes both its name and value.
	//
	// Method wasn't named Key to avoid collision with possible field.
	GetKey() map[string]types.AttributeValue
}

// Versioned is an optional interface for items that expose a version attribute for optimistic locking.
type Versioned interface {
	// GetVersion returns the current version for optimistic locking.
	//
	// Method wasn't named Version to avoid collision with possible field.
	//
	// If the item currently does not exist, the method must return an empty map instead of the zero value. This will
	// allow Save and Update operations to use attribute_not_exists instead.
	GetVersion() map[string]types.AttributeValue
	// NextVersion returns the new version, usually by incrementing the value by 1 if numeric.
	//
	// Implementations should update the field in the item to match what will be written to database.
	NextVersion() map[string]types.AttributeValue
}

// HasCreatedTimestamp is an optional interface for items that has timestamp for creation.
type HasCreatedTimestamp interface {
	// UpdateCreatedTimestamp sets the creation timestamp.
	//
	// Implementations should update the field in the item to match what will be written to database, or can also choose
	// to ignore the update request and return the existing created time (check [time.Time.IsZero] for example).
	UpdateCreatedTimestamp(now time.Time) map[string]types.AttributeValue
}

// HasModifiedTimestamp is an optional interface for items that has timestamp for last modification.
//
// This interface expects timestamp.Timestamp as the timestamp values.
type HasModifiedTimestamp interface {
	// UpdateModifiedTimestamp sets the last-modified timestamp.
	//
	// Implementations should update the field in the item to match what will be written to database.
	UpdateModifiedTimestamp(now time.Time) map[string]types.AttributeValue
}
