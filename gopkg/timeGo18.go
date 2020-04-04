// +build go1.8

package gopkg

import (
	"reflect"
	"time"
)

func timeGo18() {
	Packages["time"]["Until"] = reflect.ValueOf(time.Until)
}
