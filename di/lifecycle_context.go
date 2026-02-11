package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/google/uuid"
	diutils "github.com/lcrux/go-di/di/di-utils"
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
	EndLifecycle(...context.Context) error
}

// NewLifecycleContext creates a new instance of RegistryContext with a unique ID and an empty scopedInstances map.
//
// It allows storing and retrieving instances of services by their type within the context.
// Once the context is closed, all stored instances are cleaned up and cannot be retrieved.
func NewLifecycleContext() LifecycleContext {
	diutils.DebugLog("Creating new lifecycle context")
	ctx := &lifecycleContextImpl{
		id:    uuid.New().String(),
		cache: diutils.NewAsyncMap[string, reflect.Value](),
	}
	return ctx
}

// LifecycleContext defines the interface for managing scoped instances within a lifecycle context.
type LifecycleContext interface {
	// ID returns the unique identifier of the lifecycle context.
	ID() string
	// IsClosed indicates whether the lifecycle context has been closed.
	IsClosed() bool
	// Shutdown cleans up all scoped instances in the context.
	// It returns a slice of errors encountered during the shutdown process.
	Shutdown(...context.Context) []error
	// GetInstance retrieves an instance of the specified service type from the context.
	// It returns the instance and a boolean indicating whether the instance was found.
	GetInstance(key string) (reflect.Value, bool)
	// SetInstance stores an instance of the specified service type in the context.
	// Any existing instance of the specified type will be overwritten.
	SetInstance(key string, instance reflect.Value) error
}

// lifecycleContextImpl is the implementation of the LifecycleContext interface.
type lifecycleContextImpl struct {
	id     string
	cache  diutils.AsyncMap[string, reflect.Value]
	mutex  sync.RWMutex
	closed bool
}

// ID returns the unique identifier of the lifecycle context.
func (lctx *lifecycleContextImpl) ID() string {
	return lctx.id
}

func (lctx *lifecycleContextImpl) IsClosed() bool {
	lctx.mutex.RLock()
	defer lctx.mutex.RUnlock()
	return lctx.closed
}

func checkIfCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func setContextClosed(lctx *lifecycleContextImpl) {
	lctx.mutex.Lock()
	defer lctx.mutex.Unlock()
	lctx.closed = true
}

// Shutdown cleans up all scoped instances in the context.
// Logs the operation and confirms the context has been closed.
func (lctx *lifecycleContextImpl) Shutdown(ctxs ...context.Context) []error {
	diutils.DebugLog("[Context ID: %s] Closing lifecycle context...", lctx.ID())

	// If a context is provided, use it; otherwise, use a background context
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}
	if checkIfCanceled(ctx) {
		return []error{fmt.Errorf("context canceled before shutdown")}
	}

	defer func() {
		if !checkIfCanceled(ctx) {
			// Mark the context as closed
			setContextClosed(lctx)
		}
	}()

	// To collect errors from EndLifecycle calls
	var errors []error
	var errorsMux sync.Mutex
	setError := func(err error) {
		errorsMux.Lock()
		defer errorsMux.Unlock()
		errors = append(errors, err)
	}

	// Use a semaphore to limit the number of concurrent EndLifecycle calls
	semaphore := diutils.NewSemaphore()
	defer semaphore.Done()

	// Acquire a read lock to safely access the cache and get the keys
	cacheKeys := lctx.cache.Keys()

	wg := sync.WaitGroup{}
	for _, k := range cacheKeys {
		diutils.DebugLog("[Context ID: %s] Deleting instance for service type: %v", lctx.ID(), k)

		instance, exists := lctx.cache.Get(k)
		if !exists {
			continue
		}

		// Check if the instance implements the LifecycleListener interface, if not, skip it
		lm, ok := instance.Interface().(LifecycleListener)
		if !ok {
			diutils.DebugLog("[Context ID: %s] Instance for service type: %v does not implement LifecycleListener, skipping EndLifecycle", lctx.ID(), k)
			lctx.cache.Delete(k)
			continue
		}

		if checkIfCanceled(ctx) {
			setError(fmt.Errorf("context canceled during shutdown"))
			return errors
		}

		// Call EndLifecycle in a separate goroutine to avoid blocking
		wg.Add(1)
		semaphore.Acquire()
		go func(lm LifecycleListener, k string, lctx *lifecycleContextImpl, ctx context.Context) {
			defer wg.Done()
			defer semaphore.Release()
			defer func() {
				if r := recover(); r != nil {
					diutils.DebugLog("[Context ID: %s] Recovered from panic in EndLifecycle for service type: %v, panic: %v", lctx.ID(), k, r)

					setError(fmt.Errorf("panic in EndLifecycle for service type: %v, panic: %v", k, r))
				}
			}()

			diutils.DebugLog("[Context ID: %s] Ending lifecycle for service type: %v...", lctx.ID(), k)

			if err := lm.EndLifecycle(ctx); err != nil {
				diutils.DebugLog("[Context ID: %s] Error ending lifecycle for service type: %v, error: %v", lctx.ID(), k, err)
				setError(fmt.Errorf("error in EndLifecycle for service type: %v: %w", k, err))
			} else {
				// Remove the instance from the cache
				diutils.DebugLog("[Context ID: %s] Removing instance for service type: %v", lctx.ID(), k)
				lctx.cache.Delete(k)
			}
		}(lm, k, lctx, ctx)
	}
	wg.Wait() // Wait for all EndLifecycle calls to complete

	diutils.DebugLog("[Context ID: %s] Lifecycle context closed", lctx.ID())
	return errors
}

// GetInstance retrieves an instance of the specified service type from the context.
// Logs the operation and whether the instance was found.
func (lctx *lifecycleContextImpl) GetInstance(key string) (reflect.Value, bool) {
	if key == "" {
		diutils.DebugLog("[Context ID: %s] GetInstance called with empty service type key", lctx.ID())
		return reflect.Value{}, false
	}
	if lctx.IsClosed() {
		diutils.DebugLog("[Context ID: %s] Cannot get instance from closed lifecycle context", lctx.ID())
		return reflect.Value{}, false
	}

	lctx.mutex.RLock()
	defer lctx.mutex.RUnlock()

	diutils.DebugLog("[Context ID: %s] Getting instance for service type: %v", lctx.ID(), key)
	instance, exists := lctx.cache.Get(key)
	if exists {
		diutils.DebugLog("[Context ID: %s] Instance found for service type: %v", lctx.ID(), key)
	} else {
		diutils.DebugLog("[Context ID: %s] No instance found for service type: %v", lctx.ID(), key)
	}

	return instance, exists
}

// SetInstance stores an instance of the specified service type in the context.
// Logs the operation and confirms the instance has been set.
//
// Any existing instance, of the specified type, will be overwritten.
func (lctx *lifecycleContextImpl) SetInstance(key string, instance reflect.Value) error {
	if key == "" {
		return fmt.Errorf("service type key cannot be empty")
	}
	if !instance.IsValid() {
		return fmt.Errorf("instance value is not valid")
	}
	if lctx.IsClosed() {
		return fmt.Errorf("cannot set instance on closed lifecycle context")
	}

	lctx.mutex.Lock()
	defer lctx.mutex.Unlock()

	diutils.DebugLog("[Context ID: %s] Setting instance for service type: %v", lctx.ID(), key)
	if _, exists := lctx.cache.Get(key); exists {
		diutils.DebugLog("[Context ID: %s] Overwriting existing instance for service type: %v", lctx.ID(), key)
	}

	lctx.cache.Set(key, instance)
	diutils.DebugLog("[Context ID: %s] Instance set for service type: %v", lctx.ID(), key)
	return nil
}
