package di

import (
	"reflect"
	"sync"

	"github.com/google/uuid"
)

// NewRegistryContext creates a new instance of RegistryContext with a unique ID and an empty scopedInstances map.
//
// It allows storing and retrieving instances of services by their type within the context.
// Once the context is closed, all stored instances are cleaned up and cannot be retrieved.
func NewRegistryContext() RegistryContext {
	DebugLog("Creating new registry context")
	return &registryContextImpl{
		ID:              uuid.New().String(),
		scopedInstances: make(map[reflect.Type]reflect.Value),
	}
}

// RegistryContext defines the interface for managing scoped instances within a registry context.
type RegistryContext interface {
	// GetInstance retrieves an instance of the specified service type from the context.
	GetInstance(serviceType reflect.Type) (reflect.Value, bool)
	// SetInstance stores an instance of the specified service type in the context.
	SetInstance(serviceType reflect.Type, instance reflect.Value)
	// Close cleans up all scoped instances in the context.
	Close()
}

// registryContextImpl is the implementation of the RegistryContext interface.
type registryContextImpl struct {
	ID              string
	scopedInstances map[reflect.Type]reflect.Value
	mutex           sync.RWMutex
}

// GetInstance retrieves an instance of the specified service type from the context.
// Logs the operation and whether the instance was found.
func (ctx *registryContextImpl) GetInstance(serviceType reflect.Type) (reflect.Value, bool) {
	DebugLog("[Context ID: %s] Getting instance for service type: %v", ctx.ID, serviceType)
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	val, exists := ctx.scopedInstances[serviceType]
	if exists {
		DebugLog("[Context ID: %s] Instance found for service type: %v", ctx.ID, serviceType)
	} else {
		DebugLog("[Context ID: %s] No instance found for service type: %v", ctx.ID, serviceType)
	}
	return val, exists
}

// SetInstance stores an instance of the specified service type in the context.
// Logs the operation and confirms the instance has been set.
func (ctx *registryContextImpl) SetInstance(serviceType reflect.Type, instance reflect.Value) {
	DebugLog("[Context ID: %s] Setting instance for service type: %v", ctx.ID, serviceType)
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.scopedInstances[serviceType] = instance
	DebugLog("[Context ID: %s] Instance set for service type: %v", ctx.ID, serviceType)
}

// Close cleans up all scoped instances in the context.
// Logs the operation and confirms the context has been closed.
func (ctx *registryContextImpl) Close() {
	DebugLog("[Context ID: %s] Closing registry context", ctx.ID)
	// Clean up the scoped instances
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	for k := range ctx.scopedInstances {
		DebugLog("[Context ID: %s] Deleting instance for service type: %v", ctx.ID, k)
		delete(ctx.scopedInstances, k)
	}
	DebugLog("[Context ID: %s] Registry context closed", ctx.ID)
}
