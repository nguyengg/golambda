package mapper

import (
	"reflect"
	"time"
)

type attribute struct {
	name      string
	field     reflect.StructField
	omitempty bool
	unixtime  bool
}

func (a *attribute) isValidKey() bool {
	switch parseType(a.field) {
	case S, N, B:
		return true
	default:
		return false
	}
}

func (a *attribute) isValidVersionAttribute() bool {
	switch parseType(a.field) {
	case N:
		return true
	default:
		return false
	}
}

func (a *attribute) isValidTimestampAttribute() bool {
	return a.field.Type.ConvertibleTo(timeType)
}

func (a *attribute) get(value reflect.Value) (reflect.Value, error) {
	return value.FieldByIndexErr(a.field.Index)
}

func (a *attribute) typeName() string {
	return a.field.Type.Name()
}

var byteSliceType = reflect.TypeOf([]byte(nil))
var timeType = reflect.TypeOf(time.Time{})

type ddbType uint

const (
	None ddbType = iota
	S
	N
	B
)

// parseType uses the [reflect.StructField.Kind] to guess the DynamoDB type.
func parseType(f reflect.StructField) ddbType {
	switch ft := f.Type; ft.Kind() {
	case reflect.String:
		return S
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return N
	case reflect.Array, reflect.Slice:
		if f.Type == byteSliceType || ft.Elem().Kind() == reflect.Uint8 {
			return B
		}
		fallthrough
	default:
		return None
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
