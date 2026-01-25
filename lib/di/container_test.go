package di

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	libUtils "github.com/lcrux/go-di/lib_utils"
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

type listenerDep struct {
	called *int32
}

func (l *listenerDep) EndLifecycle() error {
	if l.called != nil {
		atomic.AddInt32(l.called, 1)
	}
	return nil
}

func TestContainer_RegisterAndResolve_Transient(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	err := Register[*depA](c, func() *depA { return &depA{name: "a"} }, Transient)
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first := Resolve[*depA](c, ctx)
	second := Resolve[*depA](c, ctx)
	if first == second {
		t.Fatal("expected transient instances to differ")
	}
}

func TestContainer_Resolve_SingletonAcrossContexts(t *testing.T) {
	c := NewContainer()
	ctx1 := c.NewContext()
	ctx2 := c.NewContext()

	created := 0
	err := Register[*depA](c, func() *depA {
		created++
		return &depA{name: "singleton"}
	}, Singleton)
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first := Resolve[*depA](c, ctx1)
	second := Resolve[*depA](c, ctx2)
	if first != second {
		t.Fatal("expected singleton instance to be shared across contexts")
	}
	if created != 1 {
		t.Fatalf("expected factory to be called once, got %d", created)
	}
}

func TestContainer_Resolve_ScopedPerContext(t *testing.T) {
	c := NewContainer()
	ctx1 := c.NewContext()
	ctx2 := c.NewContext()

	created := 0
	err := Register[*depA](c, func() *depA {
		created++
		return &depA{name: "scoped"}
	}, Scoped)
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first := Resolve[*depA](c, ctx1)
	second := Resolve[*depA](c, ctx1)
	third := Resolve[*depA](c, ctx2)
	if first != second {
		t.Fatal("expected scoped instance to be reused within the same context")
	}
	if first == third {
		t.Fatal("expected scoped instances to differ across contexts")
	}
	if created != 2 {
		t.Fatalf("expected factory to be called twice, got %d", created)
	}
}

func TestContainer_Resolve_MultipleDependencies(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, func() *depA { return &depA{name: "a"} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, func() *depB { return &depB{name: "b"} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depC](c, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depD](c, func(ca *depC) *depD { return &depD{c: ca} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	service := Resolve[*depD](c, ctx)
	if service == nil || service.c == nil || service.c.a == nil || service.c.b == nil {
		t.Fatal("expected all dependencies to be resolved")
	}
}

func TestContainer_Resolve_CircularDependencies(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, func(b *depB) *depA { return &depA{name: b.name} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, func(a *depA) *depB { return &depB{name: a.name} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected circular dependency panic")
		}
		if !strings.Contains(fmt.Sprint(r), "circular dependency") {
			t.Fatalf("expected circular dependency error, got: %v", r)
		}
	}()
	_ = Resolve[*depA](c, ctx)
}

func TestContainer_Resolve_UnregisteredService(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when resolving unregistered service")
		}
	}()
	_ = Resolve[*depA](c, ctx)
}

func TestContainer_Resolve_UnregisteredDependency(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depC](c, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when dependency is not registered")
		}
	}()
	_ = Resolve[*depC](c, ctx)
}

func TestContainer_RegisterValidation(t *testing.T) {
	c := NewContainer()

	if err := c.Register(nil, func() *depA { return &depA{} }, Transient); err == nil {
		t.Fatal("expected error for nil serviceType")
	}

	if err := c.Register(libUtils.TypeOf[*depA](), nil, Transient); err == nil {
		t.Fatal("expected error for nil factoryFn")
	}

	if err := c.Register(libUtils.TypeOf[*depA](), 42, Transient); err == nil {
		t.Fatal("expected error for non-function factoryFn")
	}

	if err := c.Register(libUtils.TypeOf[*depA](), func() (*depA, error) { return &depA{}, nil }, Transient); err == nil {
		t.Fatal("expected error for invalid return count")
	}

	if err := c.Register(libUtils.TypeOf[*depA](), func() *depB { return &depB{} }, Transient); err == nil {
		t.Fatal("expected error for mismatched return type")
	}

	if err := Register[*depA](c, func() *depA { return &depA{} }, Transient); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depA](c, func() *depA { return &depA{} }, Transient); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestResolve_NilContainer(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when container is nil")
		}
	}()
	_ = Resolve[*depA](nil, nil)
}

func TestResolve_NilContextUsesBackground(t *testing.T) {
	c := NewContainer()
	created := 0

	if err := Register[*depA](c, func() *depA {
		created++
		return &depA{name: "bg"}
	}, Singleton); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first := Resolve[*depA](c, nil)
	second := Resolve[*depA](c, nil)
	if first != second {
		t.Fatal("expected background context singleton to be reused")
	}
	if created != 1 {
		t.Fatalf("expected factory to be called once, got %d", created)
	}
}

func TestContainer_CloseContext_ShutdownsScopedInstances(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()
	called := int32(0)

	if err := Register[*listenerDep](c, func() *listenerDep {
		return &listenerDep{called: &called}
	}, Scoped); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_ = Resolve[*listenerDep](c, ctx)

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

	if err := Register[*listenerErr](c, func() *listenerErr {
		return &listenerErr{}
	}, Scoped); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_ = Resolve[*listenerErr](c, ctx1)
	_ = Resolve[*listenerErr](c, ctx2)

	errs := c.Shutdown()
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}
