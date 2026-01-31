package di

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

func TestNewLifecycleContext(t *testing.T) {
	ctx := NewLifecycleContext()

	if ctx == nil {
		t.Fatal("Expected non-nil LifecycleContext")
	}
}

type listenerOk struct {
	called *int32
}

func (l *listenerOk) EndLifecycle(_ ...context.Context) error {
	atomic.AddInt32(l.called, 1)
	return nil
}

type listenerErr struct{}

func (l *listenerErr) EndLifecycle(_ ...context.Context) error {
	return errors.New("end lifecycle failed")
}

type listenerPanic struct{}

func (l *listenerPanic) EndLifecycle(_ ...context.Context) error {
	panic("boom")
}

func TestLifecycleContext_SetAndGetInstance(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf("")
	key := diutils.NameOfType(serviceType)
	expected := reflect.ValueOf("test-instance")

	if err := ctx.SetInstance(key, expected); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	val, exists := ctx.GetInstance(key)

	if !exists {
		t.Fatal("Expected instance to exist")
	}

	if expected.Interface() != val.Interface() {
		t.Fatalf("Expected instance to match, expected %v got %v", expected.Interface(), val.Interface())
	}
}

func TestLifecycleContext_SetInstance_Overwrite(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf("")
	key := diutils.NameOfType(serviceType)
	first := reflect.ValueOf("first")
	second := reflect.ValueOf("second")

	if err := ctx.SetInstance(key, first); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	if err := ctx.SetInstance(key, second); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	val, exists := ctx.GetInstance(key)

	if !exists {
		t.Fatal("Expected instance to exist")
	}
	if val.Interface() != "second" {
		t.Fatalf("Expected overwritten instance, got %v", val.Interface())
	}
}

func TestLifecycleContext_Shutdown_RemovesNonListenerInstances(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf("")
	key := diutils.NameOfType(serviceType)
	instance := reflect.ValueOf("test-instance")

	if err := ctx.SetInstance(key, instance); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	ctx.Shutdown()

	_, exists := ctx.GetInstance(key)
	if exists {
		t.Fatal("Expected instance to be cleaned up after Shutdown")
	}
}

func TestLifecycleContext_Shutdown_InvokesLifecycleListener(t *testing.T) {
	ctx := NewLifecycleContext()
	called := int32(0)
	serviceType := reflect.TypeOf(&listenerOk{})
	key := diutils.NameOfType(serviceType)
	instance := reflect.ValueOf(&listenerOk{called: &called})

	if err := ctx.SetInstance(key, instance); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	errs := ctx.Shutdown()

	if len(errs) != 0 {
		t.Fatalf("Expected no errors, got %d", len(errs))
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("Expected EndLifecycle to be called once, got %d", called)
	}

	_, exists := ctx.GetInstance(key)
	if exists {
		t.Fatal("Expected listener instance to be removed after Shutdown")
	}
}

func TestLifecycleContext_Shutdown_CollectsErrors(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf(&listenerErr{})
	key := diutils.NameOfType(serviceType)
	instance := reflect.ValueOf(&listenerErr{})

	if err := ctx.SetInstance(key, instance); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	errs := ctx.Shutdown()

	if len(errs) != 1 {
		t.Fatalf("Expected one error, got %d", len(errs))
	}

	// Instance should remain in cache when EndLifecycle returns error
	if _, exists := ctx.GetInstance(key); exists {
		t.Fatal("Expected instance to remain after EndLifecycle error")
	}
}

func TestLifecycleContext_Shutdown_RecoversFromPanics(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf(&listenerPanic{})
	key := diutils.NameOfType(serviceType)
	instance := reflect.ValueOf(&listenerPanic{})

	if err := ctx.SetInstance(key, instance); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}
	errs := ctx.Shutdown()

	if len(errs) != 1 {
		t.Fatalf("Expected one error from panic recovery, got %d", len(errs))
	}

	// Instance should remain in cache when EndLifecycle panics
	if _, exists := ctx.GetInstance(key); exists {
		t.Fatal("Expected instance to remain after EndLifecycle panic")
	}
}

func TestLifecycleContext_Shutdown_EmptyContext(t *testing.T) {
	ctx := NewLifecycleContext()

	errs := ctx.Shutdown()
	if len(errs) != 0 {
		t.Fatalf("Expected no errors, got %d", len(errs))
	}
}

func TestLifecycleContext_Shutdown_ContextCanceledBeforeStart(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf(&listenerOk{})
	key := diutils.NameOfType(serviceType)
	called := int32(0)
	instance := reflect.ValueOf(&listenerOk{called: &called})

	if err := ctx.SetInstance(key, instance); err != nil {
		t.Fatalf("Failed to set instance: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	errs := ctx.Shutdown(cancelCtx)
	if len(errs) != 1 {
		t.Fatalf("Expected one error, got %d", len(errs))
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("Expected EndLifecycle not to be called, got %d", called)
	}
	if _, exists := ctx.GetInstance(key); !exists {
		t.Fatal("Expected instance to remain after canceled shutdown")
	}
}
