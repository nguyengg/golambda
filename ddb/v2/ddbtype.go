package v2

import (
	"reflect"
	"time"
)

var byteSliceType = reflect.TypeOf([]byte(nil))
var timeType = reflect.TypeOf(time.Time{})

type ddbType uint

const (
	None ddbType = iota
	S
	N
	B
)

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
