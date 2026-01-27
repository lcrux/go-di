package di

import (
	"sync/atomic"
	"testing"
)

type depA struct {
	name string
}

type depB struct {
	name string
}

type depC struct {
	a *depA
	b *depB
}

type depD struct {
	c *depC
}

type depWithContainer struct {
	c Container
}

type depWithContext struct {
	ctx LifecycleContext
}

type depWithContainerAndContext struct {
	c   Container
	ctx LifecycleContext
}

type listenerDep struct {
	called *int32
}

func (l *listenerDep) EndLifecycle() error {
	if l.called != nil {
		atomic.AddInt32(l.called, 1)
	}
	return nil
}

func TestContainer_Validate_MissingDependency(t *testing.T) {
	c := NewContainer()

	if err := Register[*depC](c, Transient, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depA](c, Transient, func() *depA { return &depA{name: "a"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if err := c.Validate(); err == nil {
		t.Fatal("expected validation error for missing dependency")
	}
}

func TestContainer_Validate_AllDependenciesRegistered(t *testing.T) {
	c := NewContainer()

	if err := Register[*depA](c, Transient, func() *depA { return &depA{name: "a"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, Transient, func() *depB { return &depB{name: "b"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depC](c, Transient, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if err := c.Validate(); err != nil {
		t.Fatalf("expected no validation error, got: %v", err)
	}
}

func TestContainer_CloseContext_ShutdownsScopedInstances(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()
	called := int32(0)

	if err := Register[*listenerDep](c, Scoped, func() *listenerDep {
		return &listenerDep{called: &called}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if _, err := Resolve[*listenerDep](c, ctx); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	errs := c.CloseContext(ctx)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
	if called != 1 {
		t.Fatalf("expected EndLifecycle to be called once, got %d", called)
	}
}

func TestContainer_Shutdown_CollectsContextErrors(t *testing.T) {
	c := NewContainer()
	ctx1 := c.NewContext()
	ctx2 := c.NewContext()

	if err := Register[*listenerErr](c, Scoped, func() *listenerErr {
		return &listenerErr{}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if _, err := Resolve[*listenerErr](c, ctx1); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if _, err := Resolve[*listenerErr](c, ctx2); err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	errs := c.Shutdown()
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestNewContainer_BackgroundContextInitialized(t *testing.T) {
	c := NewContainer()

	bg1 := c.BackgroundContext()
	if bg1 == nil {
		t.Fatal("expected background context to be initialized")
	}

	bg2 := c.BackgroundContext()
	if bg2 == nil {
		t.Fatal("expected background context to be non-nil on subsequent call")
	}

	if bg1.ID() != bg2.ID() {
		t.Fatal("expected background context to be stable across calls")
	}
}

func TestContainer_Shutdown_ResetsBackgroundContext(t *testing.T) {
	c := NewContainer()

	bg1 := c.BackgroundContext()
	if bg1 == nil {
		t.Fatal("expected background context to be initialized")
	}

	_ = c.Shutdown()

	bg2 := c.BackgroundContext()
	if bg2 == nil {
		t.Fatal("expected background context to be re-initialized after shutdown")
	}

	if bg1.ID() == bg2.ID() {
		t.Fatal("expected background context to be reset after shutdown")
	}
}
