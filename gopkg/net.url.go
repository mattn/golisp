// +build !appengine

package gopkg

import (
	"net/url"
	"reflect"
)

func init() {
	Packages["net/url"] = map[string]reflect.Value{
		"QueryEscape":     reflect.ValueOf(url.QueryEscape),
		"QueryUnescape":   reflect.ValueOf(url.QueryUnescape),
		"Parse":           reflect.ValueOf(url.Parse),
		"ParseRequestURI": reflect.ValueOf(url.ParseRequestURI),
		"User":            reflect.ValueOf(url.User),
		"UserPassword":    reflect.ValueOf(url.UserPassword),
		"ParseQuery":      reflect.ValueOf(url.ParseQuery),
	}
	PackageTypes["net/url"] = map[string]reflect.Type{
		"Error":  reflect.TypeOf(url.Error{}),
		"URL":    reflect.TypeOf(url.URL{}),
		"Values": reflect.TypeOf(url.Values{}),
	}
}
