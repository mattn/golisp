package gopkg

import (
	"reflect"
)

var (
	Packages     = map[string]map[string]reflect.Value{}
	PackageTypes = make(map[string]map[string]reflect.Type)
)
