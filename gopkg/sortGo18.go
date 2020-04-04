// +build go1.8

package gopkg

import (
	"reflect"
	"sort"
)

func sortGo18() {
	Packages["sort"]["Slice"] = reflect.ValueOf(sort.Slice)
	Packages["sort"]["SliceIsSorted"] = reflect.ValueOf(sort.SliceIsSorted)
	Packages["sort"]["SliceStable"] = reflect.ValueOf(sort.SliceStable)
}
