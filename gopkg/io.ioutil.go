package gopkg

import (
	"io/ioutil"
	"reflect"
)

func init() {
	Packages["io/ioutil"] = map[string]reflect.Value{
		"ReadAll":   reflect.ValueOf(ioutil.ReadAll),
		"ReadDir":   reflect.ValueOf(ioutil.ReadDir),
		"ReadFile":  reflect.ValueOf(ioutil.ReadFile),
		"WriteFile": reflect.ValueOf(ioutil.WriteFile),
	}
}
