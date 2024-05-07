package mapper

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

// HasKey allows items to provide its own implementation for returning its key attribute values.
type HasKey interface {
	// GetKey returns the key map that includes both its name and value.
	//
	// Method wasn't named Key to avoid collision with possible field.
	GetKey() (map[string]dynamodbtypes.AttributeValue, error)
}

// HasVersion allows items to provide its own implementation for optimistic locking.
type HasVersion interface {
	// ExpectVersion creates a condition that expects the version to equal the value in the item being passed in.
	ExpectVersion() (expression.ConditionBuilder, error)
	// UpdateVersion creates an update expression that sets the version attribute to a new value.
	UpdateVersion() (expression.UpdateBuilder, error)
}

// HasTimestamps allows items to provide its own implementation for timestamp generation.
type HasTimestamps interface {
	// PutTimestamps is used during PutItem requests to create new timestamps.
	PutTimestamps(map[string]dynamodbtypes.AttributeValue) error
	// UpdateTimestamps is used during UpdateItem requests to update modified timestamps.
	UpdateTimestamps() (expression.UpdateBuilder, error)
}

// MapOpts allows customisation of the New logic to create Mapper.
type MapOpts struct {
	// TagKey defaults to "dynamodbav".
	TagKey string

	// MustHaveVersion if is true, New will fail if the item does not implement HasVersion, or it fails to detect a
	// valid version attribute.
	MustHaveVersion bool
	// MustHaveTimestamps if is true, New will fail if the item does not implement HasTimestamps, or it fails to detect
	// any valid timestamp attributes.
	MustHaveTimestamps bool

	// Encoder is the attributevalue.Encoder to marshal structs into DynamoDB items.
	//
	// If nil, a default one will be created with the TagKey as the [attributevalue.EncoderOptions.TagKey]; generally
	// you don't want the encoder's tag key to be different.
	Encoder *attributevalue.Encoder
	// Decoder is the attributevalue.Decoder to unmarshal results from DynamoDB.
	//
	// If nil, a default one will be created with the TagKey as the [attributevalue.EncoderOptions.TagKey]; generally
	// you don't want the encoder's tag key to be different.
	Decoder *attributevalue.Decoder
}

const (
	hashKeyName      = "hashkey"
	sortKeyName      = "sortkey"
	versionName      = "version"
	createdTimeName  = "createdTime"
	modifiedTimeName = "modifiedTime"
)

// New creates a new DynamoDB client wrapper for a specific type and table.
//
// While type T can implement HasKey, HasVersion, or HasTimestamps, the T struct can also use `dynamodbav:"..."` tags
// with additional semantics recognized by this mapper:
//
//	// Hash key is required. Its type must be marshalled to a valid key type (S, N, or B).
//	Field string `dynamodbav:"myName,hashkey"`
//
//	// Sort key is optional. Its type must be marshalled to a valid key type (S, N, or B).
//	Field string `dynamodbav:"myName,sortkey"`
//
//	// Versioned attribute. Its type must be marshalled to type N. Otherwise, the item must implement HasVersion to
//	// provide its own optimistic locking mechanism.
//	Version int64 `dynamodbav:"myName,version"`
//
//	// Created and modified timestamp attributes. Their underlying dynamodbtypes must be time.Time, and by default are
//	// marshalled as `time.RFC3339Nano` format and can be formatted as Unix epoch second by adding tag
//	//	`dynamodb:",unixtime"`. The timestamp constructs from timestamp package are also natively supported here.
//	CreatedTime time.Time `dynamodbav:"myName,createdTime"`
//	ModifiedTime time.Time `dynamodbav:"myName,modifiedTime,unixtime"`
func New[T interface{}](client *dynamodb.Client, tableName string, optFns ...func(*MapOpts)) (*Mapper[T], error) {
	opts := &MapOpts{TagKey: "dynamodbav"}
	for _, fn := range optFns {
		fn(opts)
	}
	if opts.TagKey == "" {
		return nil, fmt.Errorf("cannot specify empty string as TagKey")
	}
	if opts.Encoder == nil {
		opts.Encoder = attributevalue.NewEncoder(func(options *attributevalue.EncoderOptions) {
			options.TagKey = opts.TagKey
		})
	}
	if opts.Decoder == nil {
		opts.Decoder = attributevalue.NewDecoder(func(options *attributevalue.DecoderOptions) {
			options.TagKey = opts.TagKey
		})
	}

	m := &Mapper[T]{
		client:    client,
		tableName: tableName,
		encoder:   attributevalue.NewEncoder(),
		decoder:   attributevalue.NewDecoder(),
		now:       time.Now,
	}

	// one pass through the fields first, picking out known attributes without validation here.
	typ := reflect.TypeFor[T]()
	attributes := make(map[string][]attribute)
	for i, n := 0, typ.NumField(); i < n; i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}

		tag := f.Tag.Get(opts.TagKey)
		if tag == "" {
			continue
		}

		tags := strings.Split(tag, ",")
		name := tags[0]
		if name == "-" || name == "" {
			continue
		}

		var isHashKey, isSortKey, isVersion, isCreatedTime, isModifiedTime bool
		attr := attribute{name: name, field: f}
		for _, tag = range tags[1:] {
			switch tag {
			case hashKeyName:
				isHashKey = true
			case sortKeyName:
				isSortKey = true
			case versionName:
				isVersion = true
			case createdTimeName:
				isCreatedTime = true
			case modifiedTimeName:
				isModifiedTime = true
			case "omitempty":
				attr.omitempty = true
			case "unixtime":
				attr.unixtime = true
			}
		}

		// this will allow same field to be sort key and timestamp for example.
		if isHashKey {
			attributes[hashKeyName] = append(attributes[hashKeyName], attr)
		}
		if isSortKey {
			attributes[sortKeyName] = append(attributes[sortKeyName], attr)
		}
		if isVersion {
			attributes[versionName] = append(attributes[versionName], attr)
		}
		if isCreatedTime {
			attributes[createdTimeName] = append(attributes[createdTimeName], attr)
		}
		if isModifiedTime {
			attributes[modifiedTimeName] = append(attributes[modifiedTimeName], attr)
		}
	}

	// hash key and optional sort key.
	var hashKey attribute
	if typ.Implements(reflect.TypeFor[HasKey]()) {
		m.getKey = func(item T, value reflect.Value) (map[string]dynamodbtypes.AttributeValue, error) {
			switch i := any(item).(type) {
			case HasKey:
				return i.GetKey()
			default:
				return nil, fmt.Errorf("item does not implement HasKey")
			}
		}
	} else {
		hashKeys := attributes[hashKeyName]
		sortKeys := attributes[sortKeyName]
		switch x, y := len(hashKeys), len(sortKeys); {
		case x == 1 && y == 1:
			hashKey = hashKeys[0]
			if !hashKey.isValidKey() {
				return nil, fmt.Errorf(`unsupported hashkey field type "%s"`, hashKey.typeName())
			}

			sortKey := sortKeys[0]
			if !sortKey.isValidKey() {
				return nil, fmt.Errorf(`unsupported sortkey field type "%s"`, sortKey.typeName())
			}

			m.getKey = func(_ T, value reflect.Value) (_ map[string]dynamodbtypes.AttributeValue, err error) {
				var hashAv, sortAv dynamodbtypes.AttributeValue
				var v reflect.Value

				v, err = hashKey.get(value)
				if err != nil {
					return
				}
				hashAv, err = m.encoder.Encode(v.Interface())
				if err != nil {
					return
				}

				v, err = sortKey.get(value)
				if err != nil {
					return
				}
				sortAv, err = m.encoder.Encode(v.Interface())
				if err != nil {
					return
				}

				return map[string]dynamodbtypes.AttributeValue{hashKey.name: hashAv, sortKey.name: sortAv}, nil
			}
		case x == 1:
			hashKey = hashKeys[0]
			if !hashKey.isValidKey() {
				return nil, fmt.Errorf(`unsupported hashkey field type "%s"`, hashKey.typeName())
			}

			m.getKey = func(_ T, value reflect.Value) (_ map[string]dynamodbtypes.AttributeValue, err error) {
				var hashAv dynamodbtypes.AttributeValue
				var v reflect.Value

				v, err = hashKey.get(value)
				if err != nil {
					return
				}
				hashAv, err = m.encoder.Encode(v.Interface())
				if err != nil {
					return
				}

				return map[string]dynamodbtypes.AttributeValue{hashKey.name: hashAv}, nil
			}
		case x > 1:
			return nil, fmt.Errorf(`found multiple hashkey fields (%d) in type "%s"`, x, typ.Name())
		case y > 1:
			return nil, fmt.Errorf(`found multiple sortkey fields (%d) in type "%s"`, y, typ.Name())
		default:
			return nil, fmt.Errorf(`no hashkey field found in type "%s"`, typ.Name())
		}
	}

	// version.
	if typ.Implements(reflect.TypeFor[HasVersion]()) {
		m.expectVersion = func(item T, _ reflect.Value) (cb expression.ConditionBuilder, err error) {
			switch i := any(item).(type) {
			case HasVersion:
				return i.ExpectVersion()
			default:
				return cb, fmt.Errorf("item does not implement HasVersion")
			}
		}
		m.nextVersion = func(item T, _ reflect.Value) (ub expression.UpdateBuilder, err error) {
			switch i := any(item).(type) {
			case HasVersion:
				return i.UpdateVersion()
			default:
				return ub, fmt.Errorf("item does not implement HasVersion")
			}
		}
	} else {
		versionAttributes := attributes[versionName]
		switch n := len(versionAttributes); {
		case n == 1:
			versionAttribute := versionAttributes[0]
			if !versionAttribute.isValidVersionAttribute() {
				return nil, fmt.Errorf(`unsupported version field type "%s"`, versionAttribute.typeName())
			}

			m.expectVersion = func(_ T, value reflect.Value) (_ expression.ConditionBuilder, err error) {
				var av dynamodbtypes.AttributeValue
				var v reflect.Value

				v, err = versionAttribute.get(value)
				if v.IsZero() {
					return expression.AttributeNotExists(expression.Name(hashKey.name)), nil
				}

				av, err = m.encoder.Encode(v.Interface())
				if err != nil {
					return
				}

				return expression.Equal(expression.Name(versionAttribute.name), expression.Value(av)), nil
			}
			m.nextVersion = func(_ T, _ reflect.Value) (expression.UpdateBuilder, error) {
				return expression.Set(expression.Name(versionAttribute.name), expression.Plus(expression.Name(versionAttribute.name), expression.Value(1))), nil
			}
		case n > 1:
			return nil, fmt.Errorf(`found multiple version fields (%d) in type "%s"`, n, typ.Name())
		case n == 0:
			fallthrough
		default:
			if opts.MustHaveVersion {
				return nil, fmt.Errorf(`no version field found in type "%s"`, typ.Name())
			}
		}
	}

	// timestamps.
	if typ.Implements(reflect.TypeFor[HasTimestamps]()) {
		m.putTimestamps = func(item T, _ reflect.Value, m map[string]dynamodbtypes.AttributeValue) error {
			switch i := any(item).(type) {
			case HasTimestamps:
				return i.PutTimestamps(m)
			default:
				return fmt.Errorf("item does not implement HasTimestamps")
			}
		}
		m.updateTimestamps = func(item T, _ reflect.Value) (ub expression.UpdateBuilder, err error) {
			switch i := any(item).(type) {
			case HasTimestamps:
				return i.UpdateTimestamps()
			default:
				return ub, fmt.Errorf("item does not implement HasTimestamps")
			}
		}
	} else {
		createdTimeAttributes := attributes[createdTimeName]
		modifiedTimeAttributes := attributes[modifiedTimeName]
		switch x, y := len(createdTimeAttributes), len(modifiedTimeAttributes); {
		case x > 1:
			return nil, fmt.Errorf(`found multiple createdTime fields (%d) in type "%s"`, x, typ.Name())
		case y > 1:
			return nil, fmt.Errorf(`found multiple modifiedTime fields (%d) in type "%s"`, x, typ.Name())
		case x == 0 && y == 0:
			if opts.MustHaveTimestamps {
				return nil, fmt.Errorf(`no timestamp fields found in type "%s"`, typ.Name())
			}
		default:
			var createdTimeAttribute, modifiedTimeAttribute attribute
			if x == 1 {
				if createdTimeAttribute = createdTimeAttributes[0]; !createdTimeAttribute.isValidTimestampAttribute() {
					return nil, fmt.Errorf(`unsupported createdTime field type "%s"`, createdTimeAttribute.typeName())
				}
			}
			if y == 1 {
				if modifiedTimeAttribute = modifiedTimeAttributes[0]; !modifiedTimeAttribute.isValidTimestampAttribute() {
					return nil, fmt.Errorf(`unsupported modifiedTime field type "%s"`, modifiedTimeAttribute.typeName())
				}
			}

			// updateTimestamps only cares about modifiedTime so we can skip this if the item doesn't have any modifiedTime.
			if y == 1 {
				m.updateTimestamps = func(_ T, value reflect.Value) (_ expression.UpdateBuilder, err error) {
					var av dynamodbtypes.AttributeValue
					var v reflect.Value
					now := m.now()

					if modifiedTimeAttribute.unixtime {
						av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
					} else {
						v, err = modifiedTimeAttribute.get(value)
						if err != nil {
							return
						}
						updateValue := reflect.ValueOf(now).Convert(v.Type())
						av, err = m.encoder.Encode(updateValue.Interface())
					}
					if err != nil {
						return
					}

					return expression.Set(expression.Name(modifiedTimeAttribute.name), expression.Value(av)), nil
				}
			}

			// putTimestamps cares about either createdTime and/or modifiedTime, and each field may be marshalled
			// differently (no idea why you would do that, but this library supports it anyway).
			m.putTimestamps = func(_ T, value reflect.Value, item map[string]dynamodbtypes.AttributeValue) (err error) {
				var av dynamodbtypes.AttributeValue
				var v reflect.Value
				now := m.now()

				if x == 1 {
					v, err = createdTimeAttribute.get(value)
					if err != nil {
						return
					}

					if v.IsZero() {
						if createdTimeAttribute.unixtime {
							av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
						} else {
							updateValue := reflect.ValueOf(now).Convert(v.Type())
							av, err = m.encoder.Encode(updateValue.Interface())
						}
						if err != nil {
							return
						}
						item[createdTimeAttribute.name] = av
					}
				}

				if y == 1 {
					v, err = modifiedTimeAttribute.get(value)
					if err != nil {
						return
					}

					if v.IsZero() {
						if modifiedTimeAttribute.unixtime {
							av, err = attributevalue.UnixTime(now).MarshalDynamoDBAttributeValue()
						} else {
							updateValue := reflect.ValueOf(now).Convert(v.Type())
							av, err = m.encoder.Encode(updateValue.Interface())
						}
						if err != nil {
							return
						}
						item[modifiedTimeAttribute.name] = av
					}
				}

				return
			}
		}
	}

	return m, nil
}

// Mapper contains metadata about items in a DynamoDB table computed from struct tags.
type Mapper[T interface{}] struct {
	getKey           func(T, reflect.Value) (map[string]dynamodbtypes.AttributeValue, error)
	expectVersion    func(T, reflect.Value) (expression.ConditionBuilder, error)
	nextVersion      func(T, reflect.Value) (expression.UpdateBuilder, error)
	putTimestamps    func(T, reflect.Value, map[string]dynamodbtypes.AttributeValue) error
	updateTimestamps func(T, reflect.Value) (expression.UpdateBuilder, error)

	client    *dynamodb.Client
	tableName string
	encoder   *attributevalue.Encoder
	decoder   *attributevalue.Decoder
	now       func() time.Time
}
