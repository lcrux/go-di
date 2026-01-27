package di

import (
	"fmt"
	"strings"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

// Resolve resolves a service of type T from the container using the provided lifecycle context.
// If the context is nil, it uses the container's background context.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func Resolve[T any](c Container, ctx LifecycleContext) (T, error) {
	// Get the registry key for the service type T
	key := diutils.NameOf[T]()

	// Resolve the service using the registry key and the provided context
	return ResolveWithKey[T](c, key, ctx)
}

// ResolveWithKey resolves a service of type T from the container using the provided key and lifecycle context.
// If the context is nil, it uses the container's background context.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// Key: The key associated with the service to resolve.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func ResolveWithKey[T any](c Container, key string, ctx LifecycleContext) (T, error) {
	var zero T
	if c == nil {
		return zero, fmt.Errorf("container cannot be nil")
	}
	if strings.TrimSpace(key) == "" {
		return zero, fmt.Errorf("key cannot be empty")
	}

	// If the provided context is nil, use the container's background context
	if ctx == nil {
		ctx = c.BackgroundContext()
	}

	inst, err := c.Resolve(key, ctx)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve service with key %v: %w", key, err)
	}

	if inst == nil {
		return zero, fmt.Errorf("resolved instance is nil for key: %v", key)
	}

	val, ok := inst.(T)
	if !ok {
		return zero, fmt.Errorf("resolved instance is not of type %v", diutils.TypeOf[T]())
	}
	return val, nil
}

// MustResolve resolves a service of type T from the container using the provided lifecycle context.
// If the context is nil, it uses the container's background context.
// Panics if the service cannot be resolved or parameters are invalid.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func MustResolve[T any](c Container, ctx LifecycleContext) T {
	instance, err := Resolve[T](c, ctx)
	if err != nil {
		panic(err)
	}

	return instance
}

// MustResolveWithKey resolves a service of type T from the container using the provided key and lifecycle context.
// If the context is nil, it uses the container's background context.
// Panics if the service cannot be resolved or parameters are invalid.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// Key: The key associated with the service to resolve.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func MustResolveWithKey[T any](c Container, key string, ctx LifecycleContext) T {
	instance, err := ResolveWithKey[T](c, key, ctx)
	if err != nil {
		panic(err)
	}
	return instance
}
