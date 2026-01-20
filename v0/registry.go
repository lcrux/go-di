package godi

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

// InstanceScope defines the lifetime of a service instance.
// Transient: A new instance is created every time the service is resolved.
// Singleton: A single instance is shared across the application lifetime.
// Scoped: A single instance is shared within a specific context.
type InstanceScope int

const (
	Transient InstanceScope = iota
	Singleton
	Scoped
)

// registryEntry represents a registered service in the registry.
type registryEntry struct {
	factoryFn        reflect.Value  // The factory function to create instances of the service
	factoryFnParams  []reflect.Type // The parameter types of the factory function
	scope            InstanceScope  // The scope of the service (Transient, Singleton, Scoped)
	singletonCache   interface{}    // Cache for singleton instances
	setSingletonOnce sync.Once      // Ensures singleton initialization happens only once
}

// servicesRegistry is the global registry for all registered services.
var servicesRegistry map[reflect.Type]*registryEntry = make(map[reflect.Type]*registryEntry)
var mutex sync.RWMutex = sync.RWMutex{}

// defatultRegistryContext is the default context used for resolving services.
var defatultRegistryContext = NewRegistryContext()

// TypeOf returns the reflect.Type of a generic type T.
func TypeOf[T interface{}]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// Register registers a service with the given factory function and scope.
// The factory function must return exactly one value of the service type.
func Register[T any](factoryFn interface{}, scope InstanceScope) error {
	serviceType := TypeOf[T]()

	if factoryFn == nil {
		return fmt.Errorf("factoryFn cannot be nil")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := servicesRegistry[serviceType]; exists {
		return fmt.Errorf("service already registered: %s", serviceType.String())
	}

	// Convert the factory function to a reflect.Value and get its type
	factoryFnValue := reflect.ValueOf(factoryFn)
	factoryFnType := factoryFnValue.Type()

	// Ensure the factory function is a valid function and returns exactly one value
	if factoryFnValue.Kind() != reflect.Func || factoryFnType.NumOut() != 1 {
		return fmt.Errorf("factoryFn must be a function that returns exactly one value")
	}

	// Ensure the factory function returns the correct type
	if factoryFnType.Out(0) != serviceType {
		return fmt.Errorf("factoryFn must return a value of type %s, returning %s", serviceType.String(), factoryFnType.Out(0).String())
	}

	// Create a new registry entry for the service
	registryEntry := &registryEntry{
		factoryFn:       factoryFnValue,
		factoryFnParams: make([]reflect.Type, factoryFnType.NumIn()),
		scope:           scope,
	}
	servicesRegistry[serviceType] = registryEntry

	// Store the parameter types of the factory function
	for i := 0; i < factoryFnType.NumIn(); i++ {
		registryEntry.factoryFnParams[i] = factoryFnType.In(i)
	}

	DebugLog("Registered service: %s with scope: %v", serviceType.String(), scope)
	return nil
}

// Resolve resolves a service of type T using the default registry context.
func Resolve[T interface{}]() (T, error) {
	return resolve[T](defatultRegistryContext)
}

// ResolveWithContext resolves a service of type T using the provided registry context.
func ResolveWithContext[T interface{}](ctx RegistryContext) (T, error) {
	return resolve[T](ctx)
}

// resolve resolves a service of type T using the provided registry context.
func resolve[T interface{}](ctx RegistryContext) (T, error) {
	var zero T
	serviceType := TypeOf[T]()

	mutex.RLock()
	_, exists := servicesRegistry[serviceType]
	mutex.RUnlock()
	if !exists {
		return zero, fmt.Errorf("service not registered: %s", serviceType.String())
	}

	DebugLog("Resolving service: %s", serviceType.String())

	dependencies, err := getDependencyTree(serviceType)
	if err != nil {
		return zero, fmt.Errorf("failed to get dependency tree for %s: %w", serviceType.String(), err)
	}

	DebugLog("Dependencies for service %s: %v", serviceType.String(), dependencies)

	resolved, err := resolveDependencies(dependencies, ctx)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve dependencies for %s: %w", serviceType.String(), err)
	}

	value, exists := resolved[serviceType]
	if !exists {
		return zero, fmt.Errorf("failed to resolve service: %s", serviceType.String())
	}

	ok := value.Type().AssignableTo(serviceType)
	if !ok {
		return zero, fmt.Errorf("resolved service is not of expected type %s, got %s", serviceType.String(), value.Type().String())
	}

	DebugLog("Successfully resolved service: %s", serviceType.String())
	return value.Interface().(T), nil
}

// getDependencyTree returns the list of dependencies for the given service type in the order they should be resolved.
func getDependencyTree(serviceType reflect.Type) ([]reflect.Type, error) {
	fifoQueue := []reflect.Type{serviceType}
	dependencies := make([]reflect.Type, 0)
	for len(fifoQueue) > 0 {
		serviceType := fifoQueue[0]

		// Check for circular dependencies
		if slices.Contains(dependencies, serviceType) {
			return nil, fmt.Errorf("circular dependency detected for service: %s", serviceType.String())
		}
		fifoQueue = fifoQueue[1:]

		mutex.RLock()
		entry, exists := servicesRegistry[serviceType]
		mutex.RUnlock()
		if !exists {
			return nil, fmt.Errorf("service not found: %s", serviceType.String())
		}
		// Add the current service to the list of dependencies
		dependencies = append(dependencies, serviceType)
		// Add the dependencies of the current service to the queue
		fifoQueue = append(fifoQueue, entry.factoryFnParams...)
	}
	slices.Reverse(dependencies)
	DebugLog("Dependency tree for service %s: %v", serviceType.String(), dependencies)
	return dependencies, nil
}

// resolveDependencies resolves the dependencies for the given list of service types.
// It returns a map of service types to their resolved reflect.Value instances.
func resolveDependencies(dependencies []reflect.Type, ctx RegistryContext) (map[reflect.Type]reflect.Value, error) {
	resolved := make(map[reflect.Type]reflect.Value)
	for _, dep := range dependencies {
		mutex.RLock()
		entry, exists := servicesRegistry[dep]
		mutex.RUnlock()
		if !exists {
			return nil, fmt.Errorf("service not found: %s", dep.String())
		}

		DebugLog("Resolving dependency: %s", dep.String())

		switch entry.scope {
		case Singleton:
			cachedInstance := entry.singletonCache
			if cachedInstance != nil {
				DebugLog("Using cached singleton instance for: %s", dep.String())
				resolved[dep] = reflect.ValueOf(cachedInstance)
				continue
			}
		case Scoped:
			cachedInstance, exists := ctx.GetInstance(dep)
			if exists {
				DebugLog("Using cached scoped instance for: %s", dep.String())
				resolved[dep] = cachedInstance
				continue
			}
		}

		params := make([]reflect.Value, 0, len(entry.factoryFnParams))
		for _, paramType := range entry.factoryFnParams {
			paramValue, exists := resolved[paramType]
			if !exists {
				return nil, fmt.Errorf("dependency %s for service %s not resolved", paramType.String(), dep.String())
			}
			params = append(params, paramValue)
		}
		instance := entry.factoryFn.Call(params)[0]
		if instance == (reflect.Value{}) {
			return nil, fmt.Errorf("factory for service %s returned an invalid instance", dep.String())
		}

		DebugLog("Created new instance for: %s", dep.String())

		switch entry.scope {
		case Singleton:
			mutex.Lock()
			entry.singletonCache = instance.Interface()
			mutex.Unlock()
		case Scoped:
			ctx.SetInstance(dep, instance)
		}
		resolved[dep] = instance
	}
	return resolved, nil
}
