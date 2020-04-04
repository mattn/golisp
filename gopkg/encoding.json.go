package gopkg

import (
	"encoding/json"
	"reflect"
)

func init() {
	Packages["encoding/json"] = map[string]reflect.Value{
		"Marshal":   reflect.ValueOf(json.Marshal),
		"Unmarshal": reflect.ValueOf(json.Unmarshal),
	}
}
