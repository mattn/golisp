package gopkg

import (
	"os/signal"
	"reflect"
)

func init() {
	Packages["os/signal"] = map[string]reflect.Value{
		"Notify": reflect.ValueOf(signal.Notify),
		"Stop":   reflect.ValueOf(signal.Stop),
	}
}
