package di

import (
	"fmt"
	"strings"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

// Register registers a service of type T with the container using the provided factory function and lifecycle scope.
//
// The factory function must be a function that returns exactly one value of type T.
// The scope determines the lifetime of the service instance (Transient, Singleton, Scoped).
//
// Parameters:
//
// Container: The container instance in which to register the service.
//
// Scope: The lifecycle scope of the service (Transient, Singleton, Scoped).
//
// FactoryFn: The factory function used to create instances of the service.
func Register[T any](c Container, scope LifecycleScope, factoryFn interface{}) error {
	serviceType := diutils.TypeOf[T]()
	key := diutils.NameOfType(serviceType)
	return RegisterWithKey[T](c, key, scope, factoryFn)
}

// RegisterWithKey registers a service of type T with the container using the provided key, factory function, and lifecycle scope.
//
// The factory function must be a function that returns exactly one value of type T.
// The scope determines the lifetime of the service instance (Transient, Singleton, Scoped).
//
// Parameters:
//
// Container: The container instance in which to register the service.
//
// Key: The key associated with the service to register.
//
// Scope: The lifecycle scope of the service (Transient, Singleton, Scoped).
//
// FactoryFn: The factory function used to create instances of the service.
func RegisterWithKey[T any](c Container, key string, scope LifecycleScope, factoryFn interface{}) error {
	if c == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("key cannot be empty")
	}

	serviceType := diutils.TypeOf[T]()
	return c.Register(serviceType, key, scope, factoryFn)
}
