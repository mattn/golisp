package gopkg

import (
	"reflect"
	"strconv"
)

func init() {
	Packages["strconv"] = map[string]reflect.Value{
		"FormatBool":  reflect.ValueOf(strconv.FormatBool),
		"FormatFloat": reflect.ValueOf(strconv.FormatFloat),
		"FormatInt":   reflect.ValueOf(strconv.FormatInt),
		"FormatUint":  reflect.ValueOf(strconv.FormatUint),
		"ParseBool":   reflect.ValueOf(strconv.ParseBool),
		"ParseFloat":  reflect.ValueOf(strconv.ParseFloat),
		"ParseInt":    reflect.ValueOf(strconv.ParseInt),
		"ParseUint":   reflect.ValueOf(strconv.ParseUint),
	}
}
