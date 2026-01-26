package diUtils

import (
	"fmt"
	"reflect"
)

// TypeOf returns the reflect.Type of a generic type T.
func TypeOf[T interface{}]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// NameOf returns the fully qualified name of a generic type T.
func NameOf[T interface{}]() string {
	to := TypeOf[T]()
	return fmt.Sprintf("%s/%s", to.PkgPath(), to.Name())
}
