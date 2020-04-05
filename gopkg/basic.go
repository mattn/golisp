package gopkg

import (
	"reflect"
)

var BasicTypes = map[string]reflect.Type{}

func init() {
	BasicTypes = map[string]reflect.Type{
		"interface": reflect.ValueOf([]interface{}{int64(1)}).Index(0).Type(),
		"bool":      reflect.TypeOf(true),
		"string":    reflect.TypeOf("a"),
		"int":       reflect.TypeOf(int(1)),
		"int32":     reflect.TypeOf(int32(1)),
		"int64":     reflect.TypeOf(int64(1)),
		"uint":      reflect.TypeOf(uint(1)),
		"uint32":    reflect.TypeOf(uint32(1)),
		"uint64":    reflect.TypeOf(uint64(1)),
		"byte":      reflect.TypeOf(byte(1)),
		"rune":      reflect.TypeOf('a'),
		"float32":   reflect.TypeOf(float32(1)),
		"float64":   reflect.TypeOf(float64(1)),
	}
}
