package gopkg

import (
	"net/http/cookiejar"
	"reflect"
)

func init() {
	Packages["net/http/cookiejar"] = map[string]reflect.Value{
		"New": reflect.ValueOf(cookiejar.New),
	}
	PackageTypes["net/http/cookiejar"] = map[string]reflect.Type{
		"Options": reflect.TypeOf(cookiejar.Options{}),
	}
}
