package tests

import (
	"reflect"
	"testing"

	godi "github.com/lcrux/go-di"
	"github.com/stretchr/testify/assert"
)

func TestNewRegistryContext(t *testing.T) {
	ctx := godi.NewRegistryContext()

	assert.NotNil(t, ctx, "Expected non-nil RegistryContext")
}

func TestRegistryContext_SetAndGetInstance(t *testing.T) {
	ctx := godi.NewRegistryContext()
	serviceType := reflect.TypeOf("string")
	instance := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, instance)
	val, exists := ctx.GetInstance(serviceType)

	assert.True(t, exists, "Expected instance to exist")

	assert.Equal(t, instance.Interface(), val.Interface(), "Expected instance to match")
}

func TestRegistryContext_Close(t *testing.T) {
	ctx := godi.NewRegistryContext()
	serviceType := reflect.TypeOf("string")
	instance := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, instance)
	ctx.Close()

	_, exists := ctx.GetInstance(serviceType)
	assert.False(t, exists, "Expected instance to be cleaned up after Close")
}
