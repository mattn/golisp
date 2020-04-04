// +build go1.7

package gopkg

import (
	"bytes"
	"reflect"
)

func bytesGo17() {
	Packages["bytes"]["ContainsRune"] = reflect.ValueOf(bytes.ContainsRune)
}
