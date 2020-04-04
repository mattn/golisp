// +build go1.10

package gopkg

import (
	"reflect"
	"time"
)

func timeGo110() {
	Packages["time"]["LoadLocationFromTZData"] = reflect.ValueOf(time.LoadLocationFromTZData)
}
