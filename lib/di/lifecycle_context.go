package di

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	libutils "github.com/lcrux/go-di/lib_utils"
)

// LifecycleScope defines the lifetime of a service instance.
type LifecycleScope int

const (
	// Transient: A new instance is created every time the service is resolved.
	Transient LifecycleScope = iota
	// Singleton: A single instance is shared across the application lifetime.
	Singleton
	// Scoped: A single instance is shared, like a singleton, within a specific context.
	Scoped
)

type LifecycleListener interface {
	EndLifecycle() error
}

// NewLifecycleContext creates a new instance of RegistryContext with a unique ID and an empty scopedInstances map.
//
// It allows storing and retrieving instances of services by their type within the context.
// Once the context is closed, all stored instances are cleaned up and cannot be retrieved.
func NewLifecycleContext() LifecycleContext {
	libutils.DebugLog("Creating new lifecycle context")
	ctx := &lifecycleContextImpl{
		id:    uuid.New().String(),
		cache: make(map[reflect.Type]reflect.Value),
	}
	return ctx
}

// LifecycleContext defines the interface for managing scoped instances within a lifecycle context.
type LifecycleContext interface {
	ID() string
	Shutdown() []error
	GetInstance(serviceType reflect.Type) (reflect.Value, bool)
	SetInstance(serviceType reflect.Type, instance reflect.Value)
}

// lifecycleContextImpl is the implementation of the LifecycleContext interface.
type lifecycleContextImpl struct {
	id    string
	cache map[reflect.Type]reflect.Value
	mutex sync.RWMutex
}

// ID returns the unique identifier of the lifecycle context.
func (ctx *lifecycleContextImpl) ID() string {
	return ctx.id
}

// Shutdown cleans up all scoped instances in the context.
// Logs the operation and confirms the context has been closed.
func (ctx *lifecycleContextImpl) Shutdown() []error {
	libutils.DebugLog("[Context ID: %s] Closing lifecycle context...", ctx.ID())

	// Use a semaphore to limit the number of concurrent EndLifecycle calls
	semaphore := libutils.NewSemaphore(10)
	defer semaphore.Done()

	var errors []error
	var errorsMux sync.Mutex
	wg := sync.WaitGroup{}

	// Acquire a read lock to safely access the cache and get the keys
	ctx.mutex.RLock()
	cacheKeys := make([]reflect.Type, 0, len(ctx.cache))
	for k := range ctx.cache {
		cacheKeys = append(cacheKeys, k)
	}
	ctx.mutex.RUnlock()
	for _, k := range cacheKeys {
		libutils.DebugLog("[Context ID: %s] Deleting instance for service type: %v", ctx.ID(), k)

		// Acquire a read lock to safely access the cache
		ctx.mutex.RLock()
		instance, exists := ctx.cache[k]
		ctx.mutex.RUnlock()
		if !exists {
			continue
		}

		// Check if the instance implements the LifecycleListener interface, if not, skip it
		lm, ok := instance.Interface().(LifecycleListener)
		if !ok {
			ctx.mutex.Lock()
			delete(ctx.cache, k)
			ctx.mutex.Unlock()
			continue
		}

		// Call EndLifecycle in a separate goroutine to avoid blocking
		wg.Add(1)
		semaphore.Acquire()
		go func(lm LifecycleListener, k reflect.Type, ctx *lifecycleContextImpl) {
			defer wg.Done()
			defer semaphore.Release()
			defer func() {
				if r := recover(); r != nil {
					libutils.DebugLog("[Context ID: %s] Recovered from panic in EndLifecycle for service type: %v, panic: %v", ctx.ID(), k, r)

					errorsMux.Lock()
					errors = append(errors, fmt.Errorf("panic in EndLifecycle for service type: %v, panic: %v", k, r))
					errorsMux.Unlock()
				}
			}()
			libutils.DebugLog("[Context ID: %s] Ending lifecycle for service type: %v...", ctx.ID(), k)
			if err := lm.EndLifecycle(); err != nil {
				libutils.DebugLog("[Context ID: %s] Error ending lifecycle for service type: %v, error: %v", ctx.ID(), k, err)

				errorsMux.Lock()
				errors = append(
					errors,
					fmt.Errorf("error in EndLifecycle for service type: %v, error: %v", k, err),
				)
				errorsMux.Unlock()
			} else {
				// Remove the instance from the cache
				libutils.DebugLog("[Context ID: %s] Removing instance for service type: %v", ctx.ID(), k)
				ctx.mutex.Lock()
				delete(ctx.cache, k)
				ctx.mutex.Unlock()
			}
		}(lm, k, ctx)
	}
	wg.Wait() // Wait for all EndLifecycle calls to complete
	libutils.DebugLog("[Context ID: %s] Lifecycle context closed", ctx.ID())
	return errors
}

// GetInstance retrieves an instance of the specified service type from the context.
// Logs the operation and whether the instance was found.
func (ctx *lifecycleContextImpl) GetInstance(serviceType reflect.Type) (reflect.Value, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	libutils.DebugLog("[Context ID: %s] Getting instance for service type: %v", ctx.ID(), serviceType)
	instance, exists := ctx.cache[serviceType]
	if exists {
		libutils.DebugLog("[Context ID: %s] Instance found for service type: %v", ctx.ID(), serviceType)
	} else {
		libutils.DebugLog("[Context ID: %s] No instance found for service type: %v", ctx.ID(), serviceType)
	}

	return instance, exists
}

// SetInstance stores an instance of the specified service type in the context.
// Logs the operation and confirms the instance has been set.
//
// Any existing instance, of the specified type, will be overwritten.
func (ctx *lifecycleContextImpl) SetInstance(serviceType reflect.Type, instance reflect.Value) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	libutils.DebugLog("[Context ID: %s] Setting instance for service type: %v", ctx.ID(), serviceType)
	if _, exists := ctx.cache[serviceType]; exists {
		libutils.DebugLog("[Context ID: %s] Overwriting existing instance for service type: %v", ctx.ID(), serviceType)
	}
	ctx.cache[serviceType] = instance
	libutils.DebugLog("[Context ID: %s] Instance set for service type: %v", ctx.ID(), serviceType)
}
