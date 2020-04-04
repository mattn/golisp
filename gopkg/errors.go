package gopkg

import (
	"errors"
	"reflect"
)

func init() {
	Packages["errors"] = map[string]reflect.Value{
		"New": reflect.ValueOf(errors.New),
	}
}
