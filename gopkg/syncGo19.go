// +build go1.9

package gopkg

import (
	"reflect"
	"sync"
)

func syncGo19() {
	PackageTypes["sync"]["Map"] = reflect.TypeOf(sync.Map{})
}
