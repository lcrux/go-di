package di

import (
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
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

func (l *listenerOk) EndLifecycle() error {
	atomic.AddInt32(l.called, 1)
	return nil
}

type listenerErr struct{}

func (l *listenerErr) EndLifecycle() error {
	return errors.New("end lifecycle failed")
}

type listenerPanic struct{}

func (l *listenerPanic) EndLifecycle() error {
	panic("boom")
}

func TestLifecycleContext_SetAndGetInstance(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf("")
	expected := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, expected)
	val, exists := ctx.GetInstance(serviceType)

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
	first := reflect.ValueOf("first")
	second := reflect.ValueOf("second")

	ctx.SetInstance(serviceType, first)
	ctx.SetInstance(serviceType, second)
	val, exists := ctx.GetInstance(serviceType)

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
	instance := reflect.ValueOf("test-instance")

	ctx.SetInstance(serviceType, instance)
	ctx.Shutdown()

	_, exists := ctx.GetInstance(serviceType)
	if exists {
		t.Fatal("Expected instance to be cleaned up after Shutdown")
	}
}

func TestLifecycleContext_Shutdown_InvokesLifecycleListener(t *testing.T) {
	ctx := NewLifecycleContext()
	called := int32(0)
	serviceType := reflect.TypeOf(&listenerOk{})
	instance := reflect.ValueOf(&listenerOk{called: &called})

	ctx.SetInstance(serviceType, instance)
	errs := ctx.Shutdown()

	if len(errs) != 0 {
		t.Fatalf("Expected no errors, got %d", len(errs))
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("Expected EndLifecycle to be called once, got %d", called)
	}

	_, exists := ctx.GetInstance(serviceType)
	if exists {
		t.Fatal("Expected listener instance to be removed after Shutdown")
	}
}

func TestLifecycleContext_Shutdown_CollectsErrors(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf(&listenerErr{})
	instance := reflect.ValueOf(&listenerErr{})

	ctx.SetInstance(serviceType, instance)
	errs := ctx.Shutdown()

	if len(errs) != 1 {
		t.Fatalf("Expected one error, got %d", len(errs))
	}

	// Instance should remain in cache when EndLifecycle returns error
	if _, exists := ctx.GetInstance(serviceType); !exists {
		t.Fatal("Expected instance to remain after EndLifecycle error")
	}
}

func TestLifecycleContext_Shutdown_RecoversFromPanics(t *testing.T) {
	ctx := NewLifecycleContext()
	serviceType := reflect.TypeOf(&listenerPanic{})
	instance := reflect.ValueOf(&listenerPanic{})

	ctx.SetInstance(serviceType, instance)
	errs := ctx.Shutdown()

	if len(errs) != 1 {
		t.Fatalf("Expected one error from panic recovery, got %d", len(errs))
	}

	// Instance should remain in cache when EndLifecycle panics
	if _, exists := ctx.GetInstance(serviceType); !exists {
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
