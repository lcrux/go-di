package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

// backgroundContextKey is the key used to identify the background lifecycle context in the container.
const backgroundContextKey = "__BACKGROUND_CONTEXT_KEY__"

// containerReflectedKey is the reflected key for the Container type.
var containerReflectedKey string = diutils.NameOfType(diutils.TypeOf[Container]())

// lifecycleContextReflectedKey is the reflected key for the LifecycleContext type.
var lifecycleContextReflectedKey = diutils.NameOfType(diutils.TypeOf[LifecycleContext]())

// Container represents a dependency injection container that manages the lifecycle of services.
type Container interface {
	NewContext() LifecycleContext
	RemoveContext(ctx LifecycleContext) error
	BackgroundContext() LifecycleContext
	Shutdown(...context.Context) []error
	Resolve(key string, ctx LifecycleContext) (interface{}, error)
	Register(serviceType reflect.Type, key string, scope LifecycleScope, factoryFn interface{}) error
	Validate() error
}

// containerEntry represents a registered service in the container.
type containerEntry struct {
	serviceType         reflect.Type      // The type of the service
	key                 string            // The key associated with the service type
	factoryFn           reflect.Value     // The factory function to create instances of the service
	factoryFnParams     []reflect.Type    // The parameter types of the factory function
	scope               LifecycleScope    // The scope of the service (Transient, Singleton, Scoped)
	mutex               sync.Mutex        // Mutex to protect access to the container entry
	dependencyTreeCache []*containerEntry // Cache for the dependency tree of this service
}

// NewContainer creates a new dependency injection container.
// It initializes the container's registry and lifecycle contexts, including the background context.
func NewContainer() Container {
	container := &containerImpl{
		registry:          diutils.NewMap[string, *containerEntry](),
		lifecycleContexts: diutils.NewMap[string, LifecycleContext](),
	}
	// Create the background lifecycle context
	container.lifecycleContexts.Set(backgroundContextKey, NewLifecycleContext())
	return container
}

// containerImpl is the concrete implementation of the Container interface.
type containerImpl struct {
	registry          *diutils.Map[string, *containerEntry]  // Map to store registered services, keyed by their unique string keys
	lifecycleContexts *diutils.Map[string, LifecycleContext] // Map to store lifecycle contexts, keyed by their unique string keys (including the background context)
	mutex             sync.RWMutex                           // Mutex to protect access when registering and validating services
}

// NewContext creates a new lifecycle context and adds it to the container.
// It returns the newly created lifecycle context.
func (c *containerImpl) NewContext() LifecycleContext {
	ctx := NewLifecycleContext()
	c.lifecycleContexts.Set(ctx.ID(), ctx)
	return ctx
}

// BackgroundContext returns the background lifecycle context.
func (c *containerImpl) BackgroundContext() LifecycleContext {
	if value, exists := c.lifecycleContexts.Get(backgroundContextKey); exists {
		return value
	}
	return nil
}

// RemoveContext removes the given lifecycle context from the container and shuts it down.
func (c *containerImpl) RemoveContext(lctx LifecycleContext) error {
	if lctx == nil || lctx.IsClosed() {
		return nil
	}

	c.lifecycleContexts.Delete(lctx.ID())

	if errs := lctx.Shutdown(); len(errs) > 0 {
		return fmt.Errorf(
			"failed to shutdown lifecycle context %s: %v", lctx.ID(),
			errors.Join(errs...),
		)
	}
	return nil
}

// Shutdown gracefully shuts down the container and all its lifecycle contexts.
//
// It returns a slice of errors encountered during the shutdown process, if any.
// If the provided context is nil, a background context will be used.
func (c *containerImpl) Shutdown(ctxs ...context.Context) []error {
	// If no context is provided, use a background context
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	// errors stores the errors encountered during the shutdown process
	var errors []error
	var errorsMutex sync.Mutex
	setErrors := func(errs ...error) {
		errorsMutex.Lock()
		defer errorsMutex.Unlock()
		errors = append(errors, errs...)
	}

	if checkIfCanceled(ctx) {
		setErrors(fmt.Errorf("shutdown canceled before starting"))
		return errors
	}

	diutils.DebugLog("Shutting down container and all lifecycle contexts...")

	semaphore := diutils.NewSemaphore(10)
	defer semaphore.Done()

	lcKeys := c.lifecycleContexts.Keys()

	wg := sync.WaitGroup{}
	for _, lck := range lcKeys {
		if checkIfCanceled(ctx) {
			setErrors(fmt.Errorf("shutdown canceled before starting"))
			return errors
		}

		semaphore.Acquire()

		lcc, _ := c.lifecycleContexts.Get(lck)

		wg.Add(1)
		go func(lc LifecycleContext) {
			defer wg.Done()
			defer semaphore.Release()

			if checkIfCanceled(ctx) {
				setErrors(fmt.Errorf("shutdown canceled for lifecycle context %s", lc.ID()))
				return
			}

			errors := lc.Shutdown(ctx)
			setErrors(errors...)
		}(lcc)
	}
	wg.Wait()

	if !checkIfCanceled(ctx) {
		// Reset the lifecycle contexts after shutdown, keeps a clean background context to avoid nil references
		c.lifecycleContexts = diutils.NewMap[string, LifecycleContext]()
		c.lifecycleContexts.Set(backgroundContextKey, NewLifecycleContext())
	}

	return errors
}

// Register registers a service with the given type, key, scope, and factory function in the container.
// It returns an error if the service cannot be registered.
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

	if _, exists := c.registry.Get(key); exists {
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
	c.registry.Set(key, entry)

	// Store the parameter types of the factory function
	for i := 0; i < factoryFnType.NumIn(); i++ {
		entry.factoryFnParams[i] = factoryFnType.In(i)
	}

	diutils.DebugLog("Registered service: %s with key: %s scope: %v", serviceType.String(), key, scope)
	return nil
}

// Validate checks that all registered services have their dependencies (factory function parameters) also registered.
// It returns an error if any service depends on an unregistered type.
func (c *containerImpl) Validate() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	registryEntries := c.registry.Values()

	for _, entry := range registryEntries {
		for _, dep := range entry.factoryFnParams {
			depKey := diutils.NameOfType(dep)
			if depKey == containerReflectedKey || depKey == lifecycleContextReflectedKey {
				continue
			}
			if _, ok := c.registry.Get(depKey); !ok {
				return fmt.Errorf("service %s depends on unregistered type %s",
					entry.serviceType.String(), dep.String())
			}
		}
	}
	return nil
}

// Resolve resolves the service identified by the given key within the provided lifecycle context.
// If no context is provided, the background context is used.
// It returns the resolved service instance or an error if the service cannot be resolved.
func (c *containerImpl) Resolve(key string, ctx LifecycleContext) (interface{}, error) {
	ctx = c.resolveContext(ctx)

	if v, ok := c.resolveSpecial(key, ctx); ok {
		return v, nil
	}

	entry, err := c.getEntry(key)
	if err != nil {
		return nil, err
	}

	return c.resolveEntryWithDeps(key, entry, ctx)
}

// resolveContext returns the provided lifecycle context if it is not nil.
// Otherwise, it returns the container's background context.
func (c *containerImpl) resolveContext(ctx LifecycleContext) LifecycleContext {
	if ctx == nil {
		return c.BackgroundContext()
	}
	return ctx
}

// resolveSpecial checks if the given key corresponds to a special service (Container or LifecycleContext).
// If it does, it returns the corresponding instance and true. Otherwise, it returns nil and false.
func (c *containerImpl) resolveSpecial(key string, ctx LifecycleContext) (interface{}, bool) {
	switch key {
	case containerReflectedKey:
		return c, true
	case lifecycleContextReflectedKey:
		return ctx, true
	default:
		return nil, false
	}
}

// getEntry retrieves the container entry for the given key.
// It returns an error if the entry does not exist.
func (c *containerImpl) getEntry(key string) (*containerEntry, error) {
	entry, exists := c.registry.Get(key)
	if !exists {
		return nil, fmt.Errorf("service with key '%s' not registered", key)
	}
	return entry, nil
}

// resolveEntryWithDeps resolves the service identified by the given key along with its dependencies.
// It first resolves all dependencies of the service and then invokes the factory function to create the service instance.
func (c *containerImpl) resolveEntryWithDeps(
	key string,
	entry *containerEntry,
	ctx LifecycleContext,
) (interface{}, error) {
	serviceType := entry.serviceType
	diutils.DebugLog("Resolving service: %s with key: %s", serviceType.String(), key)

	// Get the dependency tree for the service
	dependencies, err := c.getDependencyTree(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency tree for %s: %w", serviceType.String(), err)
	}

	diutils.DebugLog("Dependencies for service %s: %v", serviceType.String(), dependencies)

	// Resolve the dependencies for the service
	resolved, err := c.resolveDependencies(dependencies, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies for %s: %w", serviceType.String(), err)
	}

	// Retrieve the resolved instance for the requested service
	value, exists := resolved[key]
	if !exists {
		return nil, fmt.Errorf("failed to resolve service: %s", serviceType.String())
	}

	diutils.DebugLog("Successfully resolved service: %s", serviceType.String())
	return value.Interface(), nil
}

// getDependencyTree returns the dependency tree for the service identified by the given key.
// It performs a depth-first search to determine the order in which services should be resolved.
// It detects circular dependencies and returns an error if any are found.
func (c *containerImpl) getDependencyTree(key string) ([]*containerEntry, error) {

	if entry, exists := c.registry.Get(key); exists && entry.dependencyTreeCache != nil {
		return entry.dependencyTreeCache, nil
	}
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
		entry, exists := c.registry.Get(k)
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

	if entry, exists := c.registry.Get(key); exists {
		entry.dependencyTreeCache = order
	}

	return order, nil
}

// resolveDependencies resolves the dependencies for the given container entries within the provided lifecycle context.
// It returns a map of resolved instances keyed by their service keys, or an error if any dependency cannot be resolved.
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
			if err := c.persistInstance(ctx, entry, instance); err != nil {
				return zero, err
			}

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
		if cached, exists := bgCtx.GetInstance(entry.key); exists {
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
func (c *containerImpl) persistInstance(ctx LifecycleContext, entry *containerEntry, instance reflect.Value) error {
	switch entry.scope {
	case Singleton:
		// For Singleton scope, use the container's background lifecycle context
		bgCtx := c.BackgroundContext()
		// Store the singleton instance in the container background lifecycle context if it doesn't already exist
		if _, exists := bgCtx.GetInstance(entry.key); !exists {
			if err := bgCtx.SetInstance(entry.key, instance); err != nil {
				return err
			}
		}
	case Scoped:
		// For Scoped scope, use the provided lifecycle context or fall back to the container's background lifecycle context
		if ctx == nil {
			ctx = c.BackgroundContext()
		}
		// Store the scoped instance in the current lifecycle context
		if err := ctx.SetInstance(entry.key, instance); err != nil {
			return err
		}
	case Transient:
		// For Transient scope, do not cache the instance; it will be created anew each time
	}
	return nil
}
