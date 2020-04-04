// +build go1.10

package gopkg

import (
	"reflect"
	"strings"
)

func stringsGo110() {
	PackageTypes["strings"] = map[string]reflect.Type{
		"Builder": reflect.TypeOf(strings.Builder{}),
	}
}
