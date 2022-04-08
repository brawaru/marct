package reflutils

import (
	"reflect"
	"runtime"
)

// GetFunctionName reports the name of the function passed as an argument.
//
// Source: https://stackoverflow.com/a/7053871
func GetFunctionName(i any) string {
	v := reflect.ValueOf(i)
	k := v.Kind()
	if k != reflect.Func {
		return "<not a function>"
	}

	return runtime.FuncForPC(v.Pointer()).Name()
}
