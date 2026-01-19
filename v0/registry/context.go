package registry

import (
	"reflect"
	"sync"

	"github.com/google/uuid"
)

// NewRegistryContext creates a new instance of RegistryContext with a unique ID and an empty scopedInstances map.
func NewRegistryContext() RegistryContext {
	debugLog("Creating new registry context")
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
	debugLog("[Context ID: %s] Getting instance for service type: %v", ctx.ID, serviceType)
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	val, exists := ctx.scopedInstances[serviceType]
	if exists {
		debugLog("[Context ID: %s] Instance found for service type: %v", ctx.ID, serviceType)
	} else {
		debugLog("[Context ID: %s] No instance found for service type: %v", ctx.ID, serviceType)
	}
	return val, exists
}

// SetInstance stores an instance of the specified service type in the context.
// Logs the operation and confirms the instance has been set.
func (ctx *registryContextImpl) SetInstance(serviceType reflect.Type, instance reflect.Value) {
	debugLog("[Context ID: %s] Setting instance for service type: %v", ctx.ID, serviceType)
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.scopedInstances[serviceType] = instance
	debugLog("[Context ID: %s] Instance set for service type: %v", ctx.ID, serviceType)
}

// Close cleans up all scoped instances in the context.
// Logs the operation and confirms the context has been closed.
func (r *registryContextImpl) Close() {
	debugLog("[Context ID: %s] Closing registry context", r.ID)
	// Clean up the scoped instances
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for k := range r.scopedInstances {
		debugLog("[Context ID: %s] Deleting instance for service type: %v", r.ID, k)
		delete(r.scopedInstances, k)
	}
	debugLog("[Context ID: %s] Registry context closed", r.ID)
}
