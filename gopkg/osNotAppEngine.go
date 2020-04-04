// +build !appengine

package gopkg

import (
	"os"
	"reflect"
)

func osNotAppEngine() {
	Packages["os"]["Getppid"] = reflect.ValueOf(os.Getppid)
}
