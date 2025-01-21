package util

import (
	"reflect"
	"runtime"
	"strings"
)

// NameOf returns provided function or method name as string
//
//	tests.NameOf(package.SomeFunc) -> "SomeFunc"
//	tests.NameOf(SomeOtherFunc) -> "SomeOtherFunc"
func NameOf(f interface{}) string {
	v := reflect.ValueOf(f)
	if v.Kind() == reflect.Func {
		if rf := runtime.FuncForPC(v.Pointer()); rf != nil {
			name := rf.Name()

			dotIndex := strings.LastIndex(name, ".")
			name = name[dotIndex+1:]

			// some methods come with -fm extension
			minusIndex := strings.LastIndex(name, "-")
			if minusIndex != -1 {
				name = name[0:minusIndex]
			}
			return name
		}
	}
	return v.String()
}
