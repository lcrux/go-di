package di

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	diutils "github.com/lcrux/go-di/di/di-utils"
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

func TestContainer_RegisterAndResolve_Transient(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	err := Register[*depA](c, Transient, func() *depA { return &depA{name: "a"} })
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
	err := Register[*depA](c, Singleton, func() *depA {
		created++
		return &depA{name: "singleton"}
	})
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
	err := Register[*depA](c, Scoped, func() *depA {
		created++
		return &depA{name: "scoped"}
	})
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

	if err := Register[*depA](c, Transient, func() *depA { return &depA{name: "a"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, Transient, func() *depB { return &depB{name: "b"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depC](c, Transient, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depD](c, Transient, func(ca *depC) *depD { return &depD{c: ca} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	service := Resolve[*depD](c, ctx)
	if service == nil || service.c == nil || service.c.a == nil || service.c.b == nil {
		t.Fatal("expected all dependencies to be resolved")
	}
}

func TestContainer_FactoryReceivesContainer(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContainer](c, Transient, func(c Container) *depWithContainer {
		return &depWithContainer{c: c}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance := Resolve[*depWithContainer](c, ctx)
	if instance == nil || instance.c == nil {
		t.Fatal("expected container to be injected")
	}
	if instance.c != c {
		t.Fatal("expected injected container to be the same instance")
	}
}

func TestContainer_FactoryReceivesLifecycleContext(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContext](c, Transient, func(ctx LifecycleContext) *depWithContext {
		return &depWithContext{ctx: ctx}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance := Resolve[*depWithContext](c, ctx)
	if instance == nil || instance.ctx == nil {
		t.Fatal("expected lifecycle context to be injected")
	}
	if instance.ctx.ID() != ctx.ID() {
		t.Fatal("expected injected context to match the provided context")
	}
}

func TestContainer_FactoryReceivesContainerAndLifecycleContext(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContainerAndContext](c, Transient, func(c Container, ctx LifecycleContext) *depWithContainerAndContext {
		return &depWithContainerAndContext{c: c, ctx: ctx}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance := Resolve[*depWithContainerAndContext](c, ctx)
	if instance == nil || instance.c == nil || instance.ctx == nil {
		t.Fatal("expected container and lifecycle context to be injected")
	}
	if instance.c != c {
		t.Fatal("expected injected container to be the same instance")
	}
	if instance.ctx.ID() != ctx.ID() {
		t.Fatal("expected injected context to match the provided context")
	}
}

func TestContainer_Resolve_CircularDependencies(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, Transient, func(b *depB) *depA { return &depA{name: b.name} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, Transient, func(a *depA) *depB { return &depB{name: a.name} }); err != nil {
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

	if err := Register[*depC](c, Transient, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }); err != nil {
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

	depAType := diutils.TypeOf[*depA]()
	depAKey := diutils.NameOfType(depAType)

	if err := c.Register(nil, depAKey, Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error for nil serviceType")
	}

	if err := c.Register(depAType, "", Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error for empty key")
	}

	if err := c.Register(depAType, depAKey, Transient, nil); err == nil {
		t.Fatal("expected error for nil factoryFn")
	}

	if err := c.Register(depAType, depAKey, Transient, 42); err == nil {
		t.Fatal("expected error for non-function factoryFn")
	}

	if err := c.Register(depAType, depAKey, Transient, func() (*depA, error) { return &depA{}, nil }); err == nil {
		t.Fatal("expected error for invalid return count")
	}

	if err := c.Register(depAType, depAKey, Transient, func() *depB { return &depB{} }); err == nil {
		t.Fatal("expected error for mismatched return type")
	}

	if err := Register[*depA](c, Transient, func() *depA { return &depA{} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depA](c, Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error for duplicate registration")
	}

	if err := RegisterWithKey[*depA](c, "depAKey", Transient, func() *depA { return &depA{} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := RegisterWithKey[*depA](c, "depAKey", Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestContainer_ResolveWithKey_CustomKey(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := RegisterWithKey[*depA](c, "custom.key", Transient, func() *depA { return &depA{name: "custom"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance := ResolveWithKey[*depA](c, "custom.key", ctx)
	if instance == nil || instance.name != "custom" {
		t.Fatal("expected to resolve instance by custom key")
	}
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

	if err := Register[*depA](c, Singleton, func() *depA {
		created++
		return &depA{name: "bg"}
	}); err != nil {
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

	if err := Register[*listenerDep](c, Scoped, func() *listenerDep {
		return &listenerDep{called: &called}
	}); err != nil {
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

	if err := Register[*listenerErr](c, Scoped, func() *listenerErr {
		return &listenerErr{}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_ = Resolve[*listenerErr](c, ctx1)
	_ = Resolve[*listenerErr](c, ctx2)

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

func TestNewContainer_ResolvesContainerSelf(t *testing.T) {
	c := NewContainer()

	got := Resolve[Container](c, nil)
	if got == nil {
		t.Fatal("expected to resolve container instance")
	}
	if got != c {
		t.Fatal("expected resolved container to be the same instance")
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
