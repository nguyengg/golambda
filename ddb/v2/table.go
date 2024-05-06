package v2

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"strings"
	"time"
)

// Table contains metadata about items in a DynamoDB table computed from several struct tags.
type Table[T interface{}] struct {
	// TableName is given at initialisation.
	TableName string

	// HashKeyName is detected from the field with tag `hashkey:"HashKeyName"`.
	HashKeyName string
	// SortKeyName is detected from the field with tag `sortKey:"SortKeyName"`.
	//
	// This field is empty if the table does not use a composite primary key.
	SortKeyName string
	// Key creates the key attribute that is used in many DynamoDB requests.
	//
	// Override this if you are using a custom key type.
	Key func(T, reflect.Value) (map[string]dynamodbtypes.AttributeValue, error)

	// VersionName is detected from the field with tag `version:"VersionName"`.
	//
	// Currently, only type int64 is supported out of the box.
	// If VersionName is available, ExpectVersion and NextVersion will both be non-nil. You may override these functions
	// if your version field is not of type int64. To turn this feature off, assign empty string to this field.
	VersionName string
	// ExpectVersion creates a condition that expects the version to equal the value in the item being passed in.
	//
	// The function is passed an item of type T and its `reflect.ValueOf(item)` value.
	//
	// If the version is at zero, an `attribute_not_exists` condition will be created instead.
	ExpectVersion func(item T, value reflect.Value) (expression.ConditionBuilder, error)
	// NextVersion creates an update expression that sets the version attribute to a new value.
	//
	// The function is passed an item of type T and its `reflect.ValueOf(item)` value.
	NextVersion func(item T, value reflect.Value) (expression.UpdateBuilder, error)

	// CreatedTimeName is detected from the field with tag `createdTime:"CreatedTimeName"` and type time.Time.
	//
	// Created timestamp is only set if the item's created timestamp field is a zero-value time.Time.
	//
	// time.Time by default is marshaled as `time.RFC3339Nano` format. Supports marshalling as Unix epoch second
	// (by adding tag `dynamodb:",unixtime"`) out of the box.
	CreatedTimeName string
	// ModifiedTimeName is detected from the field with tag `modifiedTime:"ModifiedTimeName"` and type time.Time.
	//
	// time.Time by default is marshaled as `time.RFC3339Nano` format. Supports marshalling as Unix epoch second
	// (by adding tag `dynamodb:",unixtime"`) out of the box.
	ModifiedTimeName string
	// PutTimestamps is used during PutItem requests to create new timestamps.
	//
	// The function is passed an item of type T, its `reflect.ValueOf(item)` value, and the [dynamodb.PutItemInput.Item]
	// to be modified to add timestamps.
	PutTimestamps func(T, reflect.Value, map[string]dynamodbtypes.AttributeValue) error
	// UpdateTimestamps is used during UpdateItem requests to update modified timestamps.
	//
	// The function is passed an item of type T and its `reflect.ValueOf(item)` value.
	//
	// time.Time by default is marshaled as `time.RFC3339Nano` format. Supports marshalling as Unix epoch second
	// (by adding tag `dynamodb:",unixtime"`) out of the box.
	UpdateTimestamps func(T, reflect.Value) (expression.UpdateBuilder, error)

	client  *dynamodb.Client
	encoder *attributevalue.Encoder
	decoder *attributevalue.Decoder
	now     func() time.Time
}

// TableOpts allows customisation of the logic to create Table.
type TableOpts struct {
	// HashKeyTagKey defaults to "hashkey".
	HashKeyTagKey string
	// SortKeyTagKey defaults to "sortkey".
	SortKeyTagKey string
	// VersionTagKey defaults to "version".
	VersionTagKey string
	// CreatedTimeTagKey defaults to "createdTime".
	CreatedTimeTagKey string
	// ModifiedTimeTagKey defaults to "modifiedTime".
	ModifiedTimeTagKey string
	// DynamoDBAttributeValueTagKey defaults to "dynamodbav".
	DynamoDBAttributeValueTagKey string
	// Encoder is the attributevalue.Encoder to marshal structs into DynamoDB items.
	//
	// If nil, a default one will be created with the DynamoDBAttributeValueTagKey as the [attributevalue.EncoderOptions.TagKey].
	Encoder *attributevalue.Encoder
	// Decoder is the attributevalue.Decoder to unmarshal results from DynamoDB.
	//
	// If nil, a default one will be created with the DynamoDBAttributeValueTagKey as the [attributevalue.EncoderOptions.TagKey].
	Decoder *attributevalue.Decoder
}

// New creates a new DynamoDB client wrapper around a table.
func New[T interface{}](client *dynamodb.Client, tableName string, optFns ...func(*TableOpts)) (*Table[T], error) {
	opts := &TableOpts{
		HashKeyTagKey:                "hashkey",
		SortKeyTagKey:                "sortkey",
		VersionTagKey:                "version",
		CreatedTimeTagKey:            "createdTime",
		ModifiedTimeTagKey:           "modifiedTime",
		DynamoDBAttributeValueTagKey: "dynamodbav",
	}
	for _, fn := range optFns {
		fn(opts)
	}
	if opts.HashKeyTagKey == "" {
		return nil, fmt.Errorf("cannot specify empty string as HashKeyTagKey")
	}
	if opts.Encoder == nil {
		opts.Encoder = attributevalue.NewEncoder(func(options *attributevalue.EncoderOptions) {
			options.TagKey = opts.DynamoDBAttributeValueTagKey
		})
	}
	if opts.Decoder == nil {
		opts.Decoder = attributevalue.NewDecoder(func(options *attributevalue.DecoderOptions) {
			options.TagKey = opts.DynamoDBAttributeValueTagKey
		})
	}

	table := &Table[T]{
		TableName: tableName,
		encoder:   attributevalue.NewEncoder(),
		decoder:   attributevalue.NewDecoder(),
		client:    client,
		now:       time.Now,
	}

	t := reflect.TypeFor[T]()
	hashKeyIndex := -1
	sortKeyIndex := -1
	versionIndex := -1
	createdTimeIndex := -1
	createdTimeAsUnixTime := false
	modifiedTimeIndex := -1
	modifiedTimeAsUnixTime := false

	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)

		if v := f.Tag.Get(opts.HashKeyTagKey); v != "" {
			if table.HashKeyName != "" {
				return nil, fmt.Errorf(`multiple fields with tag "%s" found in type "%s"`, opts.HashKeyTagKey, t.Name())
			}

			if ft := parseType(f); ft == None {
				return nil, fmt.Errorf(`unsupported "%s" field with type "%s"`, opts.HashKeyTagKey, f.Type.Name())
			}

			table.HashKeyName = v
			hashKeyIndex = i
		}

		if opts.SortKeyTagKey != "" {
			if v := f.Tag.Get(opts.SortKeyTagKey); v != "" {
				if table.SortKeyName != "" {
					return nil, fmt.Errorf(`multiple fields with tag "%s" found in type "%s"`, opts.SortKeyTagKey, t.Name())
				}

				if ft := parseType(f); ft == None {
					return nil, fmt.Errorf(`unsupported "%s" field with type "%s"`, opts.SortKeyTagKey, f.Type.Name())
				}

				table.SortKeyName = v
				sortKeyIndex = i
			}
		}

		if opts.VersionTagKey != "" {
			if v := f.Tag.Get(opts.VersionTagKey); v != "" {
				if table.VersionName != "" {
					return nil, fmt.Errorf(`multiple fields with tag "%s" found in type "%s"`, opts.VersionTagKey, t.Name())
				}

				if ft := parseType(f); ft != N {
					return nil, fmt.Errorf(`unsupported "%s" field with type "%s"`, opts.VersionTagKey, f.Type.Name())
				}

				table.VersionName = v
				versionIndex = i
			}
		}

		if opts.CreatedTimeTagKey != "" {
			if v := f.Tag.Get(opts.CreatedTimeTagKey); v != "" {
				if table.CreatedTimeName != "" {
					return nil, fmt.Errorf(`multiple fields with tag "%s" found in type "%s"`, opts.CreatedTimeTagKey, t.Name())
				}

				if !f.Type.ConvertibleTo(timeType) {
					return nil, fmt.Errorf(`unsupported "%s" field with type "%s"`, opts.CreatedTimeTagKey, f.Type.Name())
				}

				table.CreatedTimeName = v
				createdTimeIndex = i

				for _, p := range strings.Split(f.Tag.Get(opts.DynamoDBAttributeValueTagKey), ",") {
					if p == "unixtime" {
						createdTimeAsUnixTime = true
						break
					}
				}
			}
		}

		if opts.ModifiedTimeTagKey != "" {
			if v := f.Tag.Get(opts.ModifiedTimeTagKey); v != "" {
				if table.ModifiedTimeName != "" {
					return nil, fmt.Errorf(`multiple fields with tag "%s" found in type "%s"`, opts.ModifiedTimeTagKey, t.Name())
				}

				if !f.Type.ConvertibleTo(timeType) {
					return nil, fmt.Errorf(`unsupported "%s" field with type "%s"`, opts.ModifiedTimeTagKey, f.Type.Name())
				}

				table.ModifiedTimeName = v
				modifiedTimeIndex = i

				for _, p := range strings.Split(f.Tag.Get(opts.DynamoDBAttributeValueTagKey), ",") {
					if p == "unixtime" {
						modifiedTimeAsUnixTime = true
						break
					}
				}
			}
		}
	}

	// hash key name is required.
	switch {
	case table.HashKeyName != "" && table.SortKeyName != "":
		table.Key = func(item T, v reflect.Value) (map[string]dynamodbtypes.AttributeValue, error) {
			hashAv, err := table.encoder.Encode(v.Field(hashKeyIndex).Interface())
			if err != nil {
				return nil, fmt.Errorf("encode hash key error: %w", err)
			}

			sortAv, err := table.encoder.Encode(v.Field(sortKeyIndex).Interface())
			if err != nil {
				return nil, fmt.Errorf("encode sort key error: %w", err)
			}

			return map[string]dynamodbtypes.AttributeValue{
				table.HashKeyName: hashAv,
				table.SortKeyName: sortAv,
			}, nil
		}
	case table.HashKeyName != "":
		table.Key = func(_ T, v reflect.Value) (map[string]dynamodbtypes.AttributeValue, error) {
			hashAv, err := table.encoder.Encode(v.Field(hashKeyIndex).Interface())
			if err != nil {
				return nil, err
			}

			return map[string]dynamodbtypes.AttributeValue{
				table.HashKeyName: hashAv,
			}, nil
		}
	default:
		return nil, fmt.Errorf(`no field with tag "%s" in type "%s"`, opts.HashKeyTagKey, t.Name())
	}

	if table.VersionName != "" {
		table.ExpectVersion = func(_ T, v reflect.Value) (cb expression.ConditionBuilder, err error) {
			f := v.Field(versionIndex)
			if f.IsZero() {
				return expression.AttributeNotExists(expression.Name(table.HashKeyName)), nil
			}

			av, err := table.encoder.Encode(f.Interface())
			if err != nil {
				return cb, fmt.Errorf("encode version error: %w", err)
			}

			return expression.Equal(expression.Name(table.VersionName), expression.Value(av)), nil
		}

		table.NextVersion = func(_ T, _ reflect.Value) (expression.UpdateBuilder, error) {
			return expression.Set(expression.Name(table.VersionName), expression.Plus(expression.Name(table.VersionName), expression.Value(1))), nil
		}
	}

	// UpdateTimestamps doesn't care for the created timestamp.
	// PutTimestamps, however, behaves differently if the item only has created timestamp for example.
	if table.ModifiedTimeName != "" {
		table.UpdateTimestamps = func(_ T, v reflect.Value) (ub expression.UpdateBuilder, err error) {
			var av dynamodbtypes.AttributeValue
			now := table.now()

			if modifiedTimeAsUnixTime {
				av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
			} else {
				f := v.Field(modifiedTimeIndex)
				updateValue := reflect.ValueOf(now).Convert(f.Type())
				av, err = table.encoder.Encode(updateValue.Interface())
			}
			if err != nil {
				return ub, fmt.Errorf("encode modified timestamp error: %w", err)
			}

			return expression.Set(expression.Name(table.ModifiedTimeName), expression.Value(av)), nil
		}
	}
	if table.CreatedTimeName != "" || table.ModifiedTimeName != "" {
		table.PutTimestamps = func(_ T, v reflect.Value, m map[string]dynamodbtypes.AttributeValue) (err error) {
			var av dynamodbtypes.AttributeValue
			now := table.now()

			if table.CreatedTimeName != "" {
				f := v.Field(createdTimeIndex)
				if f.IsZero() {
					if createdTimeAsUnixTime {
						av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
					} else {
						updateValue := reflect.ValueOf(now).Convert(f.Type())
						av, err = table.encoder.Encode(updateValue.Interface())
					}
					if err != nil {
						return fmt.Errorf("encode created timestamp error: %w", err)
					}
					m[table.CreatedTimeName] = av
				}
			}

			if table.ModifiedTimeName != "" {
				f := v.Field(modifiedTimeIndex)
				if f.IsZero() {
					if modifiedTimeAsUnixTime {
						av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
					} else {
						updateValue := reflect.ValueOf(now).Convert(f.Type())
						av, err = table.encoder.Encode(updateValue.Interface())
					}
					if err != nil {
						return fmt.Errorf("encode modified timestamp error: %w", err)
					}
					m[table.ModifiedTimeName] = av
				}
			}

			return nil
		}
	}

	return table, nil
}

// Marshal is an alias to attributevalue.Marshal using the internal Tabe.encoder.
func (t Table[T]) Marshal(in T) (dynamodbtypes.AttributeValue, error) {
	return t.encoder.Encode(in)
}

// MarshalMap is an alias to attributevalue.MarshalMap using the internal Tabe.encoder.
func (t Table[T]) MarshalMap(in T) (map[string]dynamodbtypes.AttributeValue, error) {
	av, err := t.encoder.Encode(in)
	avm, ok := av.(*dynamodbtypes.AttributeValueMemberM)
	if err == nil && av != nil && ok {
		return avm.Value, nil
	}

	return map[string]dynamodbtypes.AttributeValue{}, err
}

// Unmarshal is an alias to attributevalue.Unmarshal using the internal Table.decoder.
func (t Table[T]) Unmarshal(av dynamodbtypes.AttributeValue, out T) error {
	return t.decoder.Decode(av, out)
}

// UnmarshalMap is an alias to attributevalue.UnmarshalMap using the internal Table.decoder.
func (t Table[T]) UnmarshalMap(m map[string]dynamodbtypes.AttributeValue, out T) error {
	return t.decoder.Decode(&dynamodbtypes.AttributeValueMemberM{Value: m}, out)
}
