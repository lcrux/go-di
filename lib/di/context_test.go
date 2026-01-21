package di

import (
	"reflect"
	"testing"
)

func TestNewRegistryContext(t *testing.T) {
	ctx := NewRegistryContext()

	if ctx == nil {
		t.Fatal("Expected non-nil RegistryContext")
	}
}

func TestRegistryContext_SetAndGetInstance(t *testing.T) {
	ctx := NewRegistryContext()
	serviceType := reflect.TypeOf("string")
	expected := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, expected)
	val, exists := ctx.GetInstance(serviceType)

	if !exists {
		t.Fatal("Expected instance to exist")
	}

	if expected.Interface() != val.Interface() {
		t.Errorf("Expected instance to match, expected %v got %v", expected.Interface(), val.Interface())
	}
}

func TestRegistryContext_Close(t *testing.T) {
	ctx := NewRegistryContext()
	serviceType := reflect.TypeOf("string")
	instance := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, instance)
	ctx.Close()

	_, exists := ctx.GetInstance(serviceType)
	if exists {
		t.Fatal("Expected instance to be cleaned up after Close")
	}
}
