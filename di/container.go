package di

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

const backgroundContextKey = "BACKGROUND_CONTEXT_KEY"

var containerReflectedKey string = diutils.NameOfType(diutils.TypeOf[Container]())
var lifecycleContextReflectedKey = diutils.NameOfType(diutils.TypeOf[LifecycleContext]())

// Resolve resolves a service of type T from the container using the provided lifecycle context.
// If the context is nil, it uses the container's background context.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func Resolve[T any](c Container, ctx LifecycleContext) T {
	// Get the type of the service to resolve and its registry key
	serviceType := diutils.TypeOf[T]()
	key := diutils.NameOfType(serviceType)

	// Resolve the service using the registry key and the provided context
	instance := ResolveWithKey[T](c, key, ctx)

	return instance
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
func ResolveWithKey[T any](c Container, key string, ctx LifecycleContext) T {
	if c == nil {
		panic("container cannot be nil")
	}
	if strings.TrimSpace(key) == "" {
		panic("key cannot be empty")
	}

	// If the provided context is nil, use the container's background context
	if ctx == nil {
		ctx = c.BackgroundContext()
	}

	inst, err := c.Resolve(key, ctx)
	if err != nil {
		diutils.DebugLog("[Container] Error resolving service with key: %v, error: %v", key, err)
		panic(fmt.Sprintf("failed to resolve service with key: %v, error: %v", key, err))
	}

	if inst == nil {
		panic(fmt.Sprintf("resolved instance is nil for key: %v", key))
	}

	val, ok := inst.(T)
	if !ok {
		panic(fmt.Sprintf("resolved instance is not of type %v", diutils.TypeOf[T]()))
	}
	return val
}

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

// Container represents a dependency injection container that manages the lifecycle of services.
type Container interface {
	NewContext() LifecycleContext
	CloseContext(ctx LifecycleContext) []error
	BackgroundContext() LifecycleContext
	Shutdown() []error
	Resolve(key string, ctx LifecycleContext) (interface{}, error)
	Register(serviceType reflect.Type, key string, scope LifecycleScope, factoryFn interface{}) error
	Validate() error
}

type containerEntry struct {
	serviceType     reflect.Type   // The type of the service
	key             string         // The key associated with the service type
	factoryFn       reflect.Value  // The factory function to create instances of the service
	factoryFnParams []reflect.Type // The parameter types of the factory function
	scope           LifecycleScope // The scope of the service (Transient, Singleton, Scoped)
	mutex           sync.Mutex     // Mutex to protect access to the container entry
}

func NewContainer() Container {
	container := &containerImpl{
		registry:          make(map[string]*containerEntry),
		lifecycleContexts: make(map[string]LifecycleContext),
		mutex:             sync.RWMutex{},
	}
	// Create the background lifecycle context
	container.lifecycleContexts[backgroundContextKey] = NewLifecycleContext()
	return container
}

type containerImpl struct {
	registry          map[string]*containerEntry
	mutex             sync.RWMutex
	lifecycleContexts map[string]LifecycleContext
}

func (c *containerImpl) NewContext() LifecycleContext {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ctx := NewLifecycleContext()
	c.lifecycleContexts[ctx.ID()] = ctx
	return ctx
}

func (c *containerImpl) BackgroundContext() LifecycleContext {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lifecycleContexts[backgroundContextKey]
}

func (c *containerImpl) CloseContext(ctx LifecycleContext) []error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.lifecycleContexts, ctx.ID())
	return ctx.Shutdown()
}

func (c *containerImpl) Shutdown() []error {
	semaphore := diutils.NewSemaphore(10)
	defer semaphore.Done()

	wg := sync.WaitGroup{}
	var allErrors []error
	var allErrorsMux sync.Mutex

	c.mutex.RLock()
	lcKeys := make([]string, 0, len(c.lifecycleContexts))
	for k := range c.lifecycleContexts {
		lcKeys = append(lcKeys, k)
	}
	c.mutex.RUnlock()

	for _, key := range lcKeys {
		c.mutex.RLock()
		ctx := c.lifecycleContexts[key]
		c.mutex.RUnlock()

		wg.Add(1)
		semaphore.Acquire()
		go func(ctx LifecycleContext) {
			defer wg.Done()
			defer semaphore.Release()
			errors := ctx.Shutdown()

			allErrorsMux.Lock()
			allErrors = append(allErrors, errors...)
			allErrorsMux.Unlock()
		}(ctx)
	}
	wg.Wait()

	// Reset the lifecycle contexts after shutdown, keeps a clean background context to avoid nil references
	c.mutex.Lock()
	c.lifecycleContexts = make(map[string]LifecycleContext)
	c.lifecycleContexts[backgroundContextKey] = NewLifecycleContext()
	c.mutex.Unlock()

	return allErrors
}

func (c *containerImpl) Register(serviceType reflect.Type, key string, scope LifecycleScope, factoryFn interface{}) error {
	if serviceType == nil {
		return fmt.Errorf("serviceType cannot be nil")
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("key cannot be empty")
	}
	if factoryFn == nil {
		return fmt.Errorf("factoryFn cannot be nil")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.registry[key]; exists {
		return fmt.Errorf("service already registered with key: %s", key)
	}

	// Convert the factory function to a reflect.Value and get its type
	factoryFnValue := reflect.ValueOf(factoryFn)
	factoryFnType := factoryFnValue.Type()

	// Ensure the factory function is a valid function and returns exactly one value
	if factoryFnValue.Kind() != reflect.Func || factoryFnType.NumOut() != 1 {
		return fmt.Errorf("factoryFn must be a function that returns exactly one value")
	}

	// Ensure the factory function returns a value that is assignable to the service type
	if !factoryFnType.Out(0).AssignableTo(serviceType) {
		return fmt.Errorf("factoryFn must return a value of type %s, returning %s", serviceType.String(), factoryFnType.Out(0).String())
	}

	// Create a new registry entry for the service
	entry := &containerEntry{
		serviceType:     serviceType,
		key:             key,
		factoryFn:       factoryFnValue,
		factoryFnParams: make([]reflect.Type, factoryFnType.NumIn()),
		scope:           scope,
	}
	c.registry[key] = entry

	// Store the parameter types of the factory function
	for i := 0; i < factoryFnType.NumIn(); i++ {
		entry.factoryFnParams[i] = factoryFnType.In(i)
	}

	diutils.DebugLog("Registered service: %s with key: %s scope: %v", serviceType.String(), key, scope)
	return nil
}

func (c *containerImpl) Resolve(key string, ctx LifecycleContext) (interface{}, error) {
	var zero interface{}

	// If no context is provided, use the background context
	if ctx == nil {
		ctx = c.BackgroundContext()
	}

	// If the key corresponds to the Container, return the container itself
	if key == containerReflectedKey {
		return c, nil
	}

	// If the key corresponds to the LifecycleContext, return the provided context
	if key == lifecycleContextReflectedKey {
		return ctx, nil
	}

	// Check if the service is registered
	c.mutex.RLock()
	entry, exists := c.registry[key]
	c.mutex.RUnlock()
	if !exists {
		return zero, fmt.Errorf("service with key '%s' not registered", key)
	}
	serviceType := entry.serviceType

	diutils.DebugLog("Resolving service: %s with key: %s", serviceType.String(), key)

	// Get the dependency tree for the service
	dependencies, err := c.getDependencyTree(key)
	if err != nil {
		return zero, fmt.Errorf("failed to get dependency tree for %s: %w", serviceType.String(), err)
	}

	diutils.DebugLog("Dependencies for service %s: %v", serviceType.String(), dependencies)

	// Resolve the dependencies for the service
	resolved, err := c.resolveDependencies(dependencies, ctx)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve dependencies for %s: %w", serviceType.String(), err)
	}

	// Retrieve the resolved instance for the requested service
	value, exists := resolved[key]
	if !exists {
		return zero, fmt.Errorf("failed to resolve service: %s", serviceType.String())
	}

	diutils.DebugLog("Successfully resolved service: %s", serviceType.String())
	return value.Interface(), nil
}

func (c *containerImpl) getDependencyTree(key string) ([]*containerEntry, error) {
	seen := make(map[*containerEntry]bool)
	visiting := make(map[*containerEntry]bool)
	order := make([]*containerEntry, 0)

	var visit func(string) error
	visit = func(k string) error {
		// If the type is Container or LifecycleContext, we don't need to resolve its dependencies
		if k == containerReflectedKey || k == lifecycleContextReflectedKey {
			var typ reflect.Type
			switch k {
			case containerReflectedKey:
				typ = diutils.TypeOf[Container]()
			case lifecycleContextReflectedKey:
				typ = diutils.TypeOf[LifecycleContext]()
			}
			fakeEntry := &containerEntry{
				serviceType: typ,
				key:         k,
				scope:       Transient,
			}
			order = append(order, fakeEntry)
			visiting[fakeEntry] = false
			seen[fakeEntry] = true
			return nil
		}

		// Retrieve the container entry for the current key
		c.mutex.RLock()
		entry, exists := c.registry[k]
		c.mutex.RUnlock()
		if !exists {
			return fmt.Errorf("service not found: %s", k)
		}

		if visiting[entry] {
			return fmt.Errorf("circular dependency detected for service: %s", entry.serviceType.String())
		}
		if seen[entry] {
			return nil
		}
		visiting[entry] = true

		for _, dep := range entry.factoryFnParams {
			if err := visit(diutils.NameOfType(dep)); err != nil {
				return err
			}
		}
		visiting[entry] = false
		seen[entry] = true
		order = append(order, entry)
		return nil
	}
	if err := visit(key); err != nil {
		return nil, err
	}
	return order, nil
}

func (c *containerImpl) resolveDependencies(dependencies []*containerEntry, ctx LifecycleContext) (map[string]reflect.Value, error) {
	resolved := make(map[string]reflect.Value)
	for _, entry := range dependencies {
		depType := entry.serviceType
		// If the dependency is of type LifecycleContext, use the provided context
		if entry.key == lifecycleContextReflectedKey {
			resolved[entry.key] = reflect.ValueOf(ctx)
			continue
		}
		// If the dependency is of type Container, use the current container instance
		if entry.key == containerReflectedKey {
			resolved[entry.key] = reflect.ValueOf(c)
			continue
		}

		diutils.DebugLog("Resolving dependency: %s", depType.String())
		// Resolve the current dependency within a locked context to ensure thread safety
		instance, err := func() (reflect.Value, error) {
			if entry.scope == Singleton || entry.scope == Scoped {
				entry.mutex.Lock()
				defer entry.mutex.Unlock()
			}

			var zero reflect.Value
			// Check if the instance is already cached for Singleton or Scoped scope
			cached, ok := c.loadInstance(ctx, entry)
			if ok {
				diutils.DebugLog("Using cached instance for: %s", depType.String())
				return cached, nil
			}

			// Resolve the dependencies for the factory function
			params := make([]reflect.Value, 0, len(entry.factoryFnParams))
			for _, paramType := range entry.factoryFnParams {
				paramValue, exists := resolved[diutils.NameOfType(paramType)]
				if !exists {
					return zero, fmt.Errorf("dependency %s for service %s not resolved", paramType.String(), depType.String())
				}
				params = append(params, paramValue)
			}

			// Call the factory function to create a new instance
			instance := entry.factoryFn.Call(params)[0]

			// Verify that the created instance is valid and of the expected type
			if !instance.IsValid() || !instance.Type().AssignableTo(entry.serviceType) {
				return zero, fmt.Errorf(
					"factory for service %s returned an instance of type %s, expected %s",
					depType.String(),
					instance.Type().String(),
					entry.serviceType.String(),
				)
			}

			// Persist the created instance based on its lifecycle scope
			c.persistInstance(ctx, entry, instance)

			diutils.DebugLog("Created new instance for: %s", depType.String())
			return instance, nil
		}()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s: %w", depType.String(), err)
		}

		// Add the created instance to the resolved map
		resolved[entry.key] = instance
	}
	return resolved, nil
}

// loadInstance attempts to load a cached instance of the given service type based on its scope.
//
// It returns the cached instance and a boolean indicating whether the instance was found in the cache.
func (c *containerImpl) loadInstance(ctx LifecycleContext, entry *containerEntry) (reflect.Value, bool) {
	switch entry.scope {
	case Singleton:
		// For Singleton scope, use the container's background lifecycle context
		bgCtx := c.BackgroundContext()
		// If the instance is already cached in the container background lifecycle context, return it
		if cached, found := bgCtx.GetInstance(entry.key); found {
			return cached, true
		}
	case Scoped:
		// For Scoped scope, use the provided lifecycle context or fall back to the container's background lifecycle context
		if ctx == nil {
			ctx = c.BackgroundContext()
		}
		// If the instance is already cached in the current lifecycle context, return it
		instance, exists := ctx.GetInstance(entry.key)
		if exists {
			return instance, true
		}
	case Transient:
		// For Transient scope, do not cache the instance; it will be created anew each time
	}
	return reflect.Value{}, false
}

// persistInstance stores the given instance in the appropriate cache based on its scope.
func (c *containerImpl) persistInstance(ctx LifecycleContext, entry *containerEntry, instance reflect.Value) {
	switch entry.scope {
	case Singleton:
		// For Singleton scope, use the container's background lifecycle context
		bgCtx := c.BackgroundContext()
		// Store the singleton instance in the container background lifecycle context if it doesn't already exist
		if _, exists := bgCtx.GetInstance(entry.key); !exists {
			bgCtx.SetInstance(entry.key, instance)
		}
	case Scoped:
		// For Scoped scope, use the provided lifecycle context or fall back to the container's background lifecycle context
		if ctx == nil {
			ctx = c.BackgroundContext()
		}
		// Store the scoped instance in the current lifecycle context
		ctx.SetInstance(entry.key, instance)
	case Transient:
		// For Transient scope, do not cache the instance; it will be created anew each time
	}
}

// Validate checks that all registered services have their dependencies (factory function parameters) also registered.
// It returns an error if any service depends on an unregistered type.
func (c *containerImpl) Validate() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, entry := range c.registry {
		for _, dep := range entry.factoryFnParams {
			depKey := diutils.NameOfType(dep)
			if depKey == containerReflectedKey || depKey == lifecycleContextReflectedKey {
				continue
			}
			if _, ok := c.registry[depKey]; !ok {
				return fmt.Errorf("service %s depends on unregistered type %s",
					entry.serviceType.String(), dep.String())
			}
		}
	}
	return nil
}
