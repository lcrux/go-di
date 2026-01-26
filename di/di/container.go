package di

import (
	"fmt"
	"reflect"
	"sync"

	diUtils "github.com/lcrux/go-di/di-utils"
)

const backgroundContextKey = "BACKGROUND_CONTEXT_KEY"

// Resolve resolves a service of type T from the container using the provided lifecycle context.
// If the context is nil, it uses the container's background context.
//
// Parameters:
//
// Container: The container instance from which to resolve the service.
//
// LifecycleContext: The lifecycle context to use for resolving the service. If nil, the container's background context is used.
func Resolve[T any](c Container, ctx LifecycleContext) T {
	if c == nil {
		panic("container cannot be nil")
	}

	// If the provided context is nil, use the container's background context
	if ctx == nil {
		ctx = c.BackgroundContext()
	}

	serviceType := diUtils.TypeOf[T]()
	inst, err := c.Resolve(serviceType, ctx)
	if err != nil {
		diUtils.DebugLog("[Container] Error resolving service of type: %v, error: %v", serviceType, err)
		panic(fmt.Sprintf("failed to resolve service of type: %v, error: %v", serviceType, err))
	}
	return inst.(T)
}

// Register registers a service of type T with the container using the provided factory function and lifecycle scope.
//
// The factory function must be a function that returns exactly one value of type T.
// The scope determines the lifetime of the service instance (Transient, Singleton, Scoped).
func Register[T any](c Container, factoryFn interface{}, scope LifecycleScope) error {
	return c.Register(diUtils.TypeOf[T](), factoryFn, scope)
}

type Container interface {
	NewContext() LifecycleContext
	CloseContext(ctx LifecycleContext) []error
	BackgroundContext() LifecycleContext
	Shutdown() []error
	Resolve(serviceType reflect.Type, ctx LifecycleContext) (interface{}, error)
	Register(serviceType reflect.Type, factoryFn interface{}, scope LifecycleScope) error
}

type containerEntry struct {
	serviceType     reflect.Type   // The type of the service
	factoryFn       reflect.Value  // The factory function to create instances of the service
	factoryFnParams []reflect.Type // The parameter types of the factory function
	scope           LifecycleScope // The scope of the service (Transient, Singleton, Scoped)
	mutex           sync.Mutex     // Mutex to protect access to the container entry
}

func NewContainer() Container {
	container := &containerImpl{
		registry:          make(map[reflect.Type]*containerEntry),
		lifecycleContexts: make(map[string]LifecycleContext),
		mutex:             sync.RWMutex{},
	}
	container.lifecycleContexts[backgroundContextKey] = NewLifecycleContext()
	return container
}

type containerImpl struct {
	registry          map[reflect.Type]*containerEntry
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
	semaphore := diUtils.NewSemaphore(10)
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

func (c *containerImpl) Register(serviceType reflect.Type, factoryFn interface{}, scope LifecycleScope) error {
	if serviceType == nil {
		return fmt.Errorf("serviceType cannot be nil")
	}
	if factoryFn == nil {
		return fmt.Errorf("factoryFn cannot be nil")
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.registry[serviceType]; exists {
		return fmt.Errorf("service already registered: %s", serviceType.String())
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
		factoryFn:       factoryFnValue,
		factoryFnParams: make([]reflect.Type, factoryFnType.NumIn()),
		scope:           scope,
	}
	c.registry[serviceType] = entry

	// Store the parameter types of the factory function
	for i := 0; i < factoryFnType.NumIn(); i++ {
		entry.factoryFnParams[i] = factoryFnType.In(i)
	}

	diUtils.DebugLog("Registered service: %s with scope: %v", serviceType.String(), scope)
	return nil
}

func (c *containerImpl) Resolve(serviceType reflect.Type, ctx LifecycleContext) (interface{}, error) {
	if ctx == nil {
		ctx = c.BackgroundContext()
	}
	var zero interface{}

	diUtils.DebugLog("Resolving service: %s", serviceType.String())

	// Check if the service is registered
	c.mutex.RLock()
	_, exists := c.registry[serviceType]
	c.mutex.RUnlock()
	if !exists {
		return zero, fmt.Errorf("service not registered: %s", serviceType.String())
	}

	// Get the dependency tree for the service
	dependencies, err := c.getDependencyTree(serviceType)
	if err != nil {
		return zero, fmt.Errorf("failed to get dependency tree for %s: %w", serviceType.String(), err)
	}

	diUtils.DebugLog("Dependencies for service %s: %v", serviceType.String(), dependencies)

	// Resolve the dependencies for the service
	resolved, err := c.resolveDependencies(dependencies, ctx)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve dependencies for %s: %w", serviceType.String(), err)
	}

	// Retrieve the resolved instance for the requested service
	value, exists := resolved[serviceType]
	if !exists {
		return zero, fmt.Errorf("failed to resolve service: %s", serviceType.String())
	}

	diUtils.DebugLog("Successfully resolved service: %s", serviceType.String())
	return value.Interface(), nil
}

func (c *containerImpl) getDependencyTree(serviceType reflect.Type) ([]reflect.Type, error) {
	seen := make(map[reflect.Type]bool)
	visiting := make(map[reflect.Type]bool)
	order := make([]reflect.Type, 0)

	var visit func(reflect.Type) error
	visit = func(t reflect.Type) error {
		if visiting[t] {
			return fmt.Errorf("circular dependency detected for service: %s", t.String())
		}
		if seen[t] {
			return nil
		}
		visiting[t] = true
		c.mutex.RLock()
		entry, exists := c.registry[t]
		c.mutex.RUnlock()
		if !exists {
			return fmt.Errorf("service not found: %s", t.String())
		}
		for _, dep := range entry.factoryFnParams {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[t] = false
		seen[t] = true
		order = append(order, t)
		return nil
	}
	if err := visit(serviceType); err != nil {
		return nil, err
	}
	return order, nil
}

func (c *containerImpl) resolveDependencies(dependencies []reflect.Type, ctx LifecycleContext) (map[reflect.Type]reflect.Value, error) {
	resolved := make(map[reflect.Type]reflect.Value)
	for _, depType := range dependencies {
		// Retrieve the container entry for the current dependency
		c.mutex.RLock()
		entry, exists := c.registry[depType]
		c.mutex.RUnlock()
		if !exists {
			return nil, fmt.Errorf("service not found: %s", depType.String())
		}

		diUtils.DebugLog("Resolving dependency: %s", depType.String())
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
				diUtils.DebugLog("Using cached instance for: %s", depType.String())
				return cached, nil
			}

			// Resolve the dependencies for the factory function
			params := make([]reflect.Value, 0, len(entry.factoryFnParams))
			for _, paramType := range entry.factoryFnParams {
				paramValue, exists := resolved[paramType]
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

			diUtils.DebugLog("Created new instance for: %s", depType.String())

			return instance, nil
		}()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s: %w", depType.String(), err)
		}

		// Add the created instance to the resolved map
		resolved[depType] = instance
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
		if cached, found := bgCtx.GetInstance(entry.serviceType); found {
			return cached, true
		}
	case Scoped:
		// For Scoped scope, use the provided lifecycle context or fall back to the container's background lifecycle context
		if ctx == nil {
			ctx = c.BackgroundContext()
		}
		// If the instance is already cached in the current lifecycle context, return it
		instance, exists := ctx.GetInstance(entry.serviceType)
		if exists {
			return instance, true
		}
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
		if _, exists := bgCtx.GetInstance(entry.serviceType); !exists {
			bgCtx.SetInstance(entry.serviceType, instance)
		}
	case Scoped:
		// For Scoped scope, use the provided lifecycle context or fall back to the container's background lifecycle context
		if ctx == nil {
			ctx = c.BackgroundContext()
		}
		// Store the scoped instance in the current lifecycle context
		ctx.SetInstance(entry.serviceType, instance)
	}
}
