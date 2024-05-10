package mapper

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"reflect"
	"strconv"
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
//
// The rest of the documentation will explain the default logic that is provided from struct tags.
type HasVersion interface {
	// PutVersion is used during Mapper.Put requests.
	//
	// PutVersion is passed the [dynamodb.PutItemInput.Item] and must modify this map to contain the version that will
	// be written to DynamoDB. The method also returns a condition expression to verify the existing version of the item
	// prior to the PutItem request.
	//
	// The tag-based implementation will return an `attribute_not_exists(#hash_key)` condition expression if the version
	// attribute's value is the zero value.
	PutVersion(map[string]dynamodbtypes.AttributeValue) (expression.ConditionBuilder, error)

	// UpdateVersion is used during Mapper.Update requests.
	//
	// PutVersion is passed the update expression that must be used to update the version in DynamoDB. The method also
	// returns a condition expression to verify the existing version of the item prior to the UpdateItem request.
	//
	// The tag-based implementation will return an `attribute_not_exists(#hash_key)` condition expression if the version
	// attribute's value is the zero value.
	UpdateVersion(expression.UpdateBuilder) (expression.UpdateBuilder, expression.ConditionBuilder, error)

	// DeleteVersion is used during Mapper.Delete requests.
	//
	// DeleteVersion must return a condition expression to verify the existing version of the item prior to the
	// DeleteItem request. The tag-based implementation will always add a `#version = :version` condition expression
	// even if the current version attribute's value is the zero value.
	DeleteVersion() (expression.ConditionBuilder, error)
}

// HasTimestamps allows items to provide its own implementation for timestamp generation.
//
// The rest of the documentation will explain the default logic that is provided from struct tags.
type HasTimestamps interface {
	// PutTimestamps is used during Mapper.Put requests.
	//
	// PutVersion is passed the [dynamodb.PutItemInput.Item] and must modify this map to contain the timestamps that
	// will be written to DynamoDB.
	PutTimestamps(map[string]dynamodbtypes.AttributeValue) error

	// UpdateTimestamps is used during Mapper.Update requests.
	//
	// UpdateTimestamps is passed the update expression that must be used to update the timestamps in DynamoDB.
	UpdateTimestamps(expression.UpdateBuilder) (expression.UpdateBuilder, error)
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
		m.putVersion = func(item T, _ reflect.Value, in map[string]dynamodbtypes.AttributeValue) (expression.ConditionBuilder, error) {
			switch i := any(item).(type) {
			case HasVersion:
				return i.PutVersion(in)
			default:
				return expression.ConditionBuilder{}, fmt.Errorf("item does not implement HasVersion")
			}
		}
		m.updateVersion = func(item T, _ reflect.Value, update expression.UpdateBuilder) (expression.UpdateBuilder, expression.ConditionBuilder, error) {
			switch i := any(item).(type) {
			case HasVersion:
				return i.UpdateVersion(update)
			default:
				return expression.UpdateBuilder{}, expression.ConditionBuilder{}, fmt.Errorf("item does not implement HasVersion")
			}
		}
		m.deleteVersion = func(item T, _ reflect.Value) (expression.ConditionBuilder, error) {
			switch i := any(item).(type) {
			case HasVersion:
				return i.DeleteVersion()
			default:
				return expression.ConditionBuilder{}, fmt.Errorf("item does not implement HasVersion")
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

			m.putVersion = func(_ T, value reflect.Value, in map[string]dynamodbtypes.AttributeValue) (expression.ConditionBuilder, error) {
				v, err := versionAttribute.get(value)
				if err != nil {
					return expression.ConditionBuilder{}, err
				}

				switch {
				case v.IsZero():
					in[versionAttribute.name] = &dynamodbtypes.AttributeValueMemberN{Value: "1"}
					return expression.AttributeNotExists(expression.Name(hashKey.name)), nil
				case v.CanInt():
					in[versionAttribute.name] = &dynamodbtypes.AttributeValueMemberN{Value: strconv.FormatInt(v.Int()+1, 10)}
				case v.CanUint():
					in[versionAttribute.name] = &dynamodbtypes.AttributeValueMemberN{Value: strconv.FormatInt(int64(v.Uint()+1), 10)}
				case v.CanFloat():
					in[versionAttribute.name] = &dynamodbtypes.AttributeValueMemberN{Value: strconv.FormatFloat(v.Float(), 'f', -1, 64)}
				default:
					return expression.ConditionBuilder{}, fmt.Errorf("version attribute is not numeric, item must implement HasVersion")
				}

				av, err := m.encoder.Encode(v.Interface())
				if err != nil {
					return expression.ConditionBuilder{}, err
				}

				return expression.Equal(expression.Name(versionAttribute.name), expression.Value(av)), nil
			}

			m.updateVersion = func(_ T, value reflect.Value, update expression.UpdateBuilder) (expression.UpdateBuilder, expression.ConditionBuilder, error) {
				v, err := versionAttribute.get(value)
				if err != nil {
					return expression.UpdateBuilder{}, expression.ConditionBuilder{}, err
				}

				if v.IsZero() {
					update = update.Set(expression.Name(versionAttribute.name), expression.Value(1))
					cond := expression.AttributeNotExists(expression.Name(hashKey.name))
					return update, cond, nil
				}

				av, err := m.encoder.Encode(v.Interface())
				if err != nil {
					return expression.UpdateBuilder{}, expression.ConditionBuilder{}, err
				}

				update = update.Set(expression.Name(versionAttribute.name), expression.Plus(expression.Name(versionAttribute.name), expression.Value(1)))
				cond := expression.Equal(expression.Name(versionAttribute.name), expression.Value(av))
				return update, cond, nil
			}

			m.deleteVersion = func(_ T, value reflect.Value) (expression.ConditionBuilder, error) {
				v, err := versionAttribute.get(value)
				if err != nil {
					return expression.ConditionBuilder{}, err
				}

				av, err := m.encoder.Encode(v.Interface())
				if err != nil {
					return expression.ConditionBuilder{}, err
				}

				cond := expression.Equal(expression.Name(versionAttribute.name), expression.Value(av))
				return cond, nil
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
		m.updateTimestamps = func(item T, _ reflect.Value, update expression.UpdateBuilder) (expression.UpdateBuilder, error) {
			switch i := any(item).(type) {
			case HasTimestamps:
				return i.UpdateTimestamps(update)
			default:
				return expression.UpdateBuilder{}, fmt.Errorf("item does not implement HasTimestamps")
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
				m.updateTimestamps = func(_ T, value reflect.Value, update expression.UpdateBuilder) (expression.UpdateBuilder, error) {
					if modifiedTimeAttribute.unixtime {
						av, err := attributevalue.UnixTime(m.now()).MarshalDynamoDBAttributeValue()
						if err != nil {
							return update, err
						}

						update = update.Set(expression.Name(modifiedTimeAttribute.name), expression.Value(av))
						return update, nil
					}

					v, err := modifiedTimeAttribute.get(value)
					if err != nil {
						return update, err
					}
					updateValue := reflect.ValueOf(m.now()).Convert(v.Type())
					av, err := m.encoder.Encode(updateValue.Interface())
					if err == nil {
						update = update.Set(expression.Name(modifiedTimeAttribute.name), expression.Value(av))
					}
					return update, err
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
	putVersion       func(T, reflect.Value, map[string]dynamodbtypes.AttributeValue) (expression.ConditionBuilder, error)
	updateVersion    func(T, reflect.Value, expression.UpdateBuilder) (expression.UpdateBuilder, expression.ConditionBuilder, error)
	deleteVersion    func(T, reflect.Value) (expression.ConditionBuilder, error)
	putTimestamps    func(T, reflect.Value, map[string]dynamodbtypes.AttributeValue) error
	updateTimestamps func(T, reflect.Value, expression.UpdateBuilder) (expression.UpdateBuilder, error)

	client    *dynamodb.Client
	tableName string
	encoder   *attributevalue.Encoder
	decoder   *attributevalue.Decoder
	now       func() time.Time
}
