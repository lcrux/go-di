package diutils

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
	return NameOfType(to)
}

// NameOfType returns the fully qualified name of a reflect.Type.
func NameOfType(t reflect.Type) string {
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}
