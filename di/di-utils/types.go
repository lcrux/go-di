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
	var pkgPath string
	var tName string

	if t.Kind() == reflect.Ptr {
		pkgPath = t.Elem().PkgPath()
		tName = t.Elem().Name()
	} else {
		pkgPath = t.PkgPath()
		tName = t.Name()
	}

	if tName == "" {
		return t.String()
	}

	if pkgPath == "" {
		return tName
	}

	return fmt.Sprintf("%s/%s", pkgPath, tName)
}
