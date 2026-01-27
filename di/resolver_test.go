package di

import (
	"strings"
	"testing"
)

func TestResolve_TransientDifferentInstances(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, Transient, func() *depA { return &depA{name: "a"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first, err := Resolve[*depA](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	second, err := Resolve[*depA](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if first == second {
		t.Fatal("expected transient instances to differ")
	}
}

func TestResolve_SingletonAcrossContexts(t *testing.T) {
	c := NewContainer()
	ctx1 := c.NewContext()
	ctx2 := c.NewContext()

	created := 0
	if err := Register[*depA](c, Singleton, func() *depA {
		created++
		return &depA{name: "singleton"}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first, err := Resolve[*depA](c, ctx1)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	second, err := Resolve[*depA](c, ctx2)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if first != second {
		t.Fatal("expected singleton instance to be shared across contexts")
	}
	if created != 1 {
		t.Fatalf("expected factory to be called once, got %d", created)
	}
}

func TestResolve_ScopedPerContext(t *testing.T) {
	c := NewContainer()
	ctx1 := c.NewContext()
	ctx2 := c.NewContext()

	created := 0
	if err := Register[*depA](c, Scoped, func() *depA {
		created++
		return &depA{name: "scoped"}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	first, err := Resolve[*depA](c, ctx1)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	second, err := Resolve[*depA](c, ctx1)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	third, err := Resolve[*depA](c, ctx2)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
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

func TestResolve_MultipleDependencies(t *testing.T) {
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

	service, err := Resolve[*depD](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if service == nil || service.c == nil || service.c.a == nil || service.c.b == nil {
		t.Fatal("expected all dependencies to be resolved")
	}
}

func TestResolve_FactoryReceivesContainer(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContainer](c, Transient, func(c Container) *depWithContainer {
		return &depWithContainer{c: c}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance, err := Resolve[*depWithContainer](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if instance == nil || instance.c == nil {
		t.Fatal("expected container to be injected")
	}
	if instance.c != c {
		t.Fatal("expected injected container to be the same instance")
	}
}

func TestResolve_FactoryReceivesLifecycleContext(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContext](c, Transient, func(ctx LifecycleContext) *depWithContext {
		return &depWithContext{ctx: ctx}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance, err := Resolve[*depWithContext](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if instance == nil || instance.ctx == nil {
		t.Fatal("expected lifecycle context to be injected")
	}
	if instance.ctx.ID() != ctx.ID() {
		t.Fatal("expected injected context to match the provided context")
	}
}

func TestResolve_FactoryReceivesContainerAndLifecycleContext(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depWithContainerAndContext](c, Transient, func(c Container, ctx LifecycleContext) *depWithContainerAndContext {
		return &depWithContainerAndContext{c: c, ctx: ctx}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance, err := Resolve[*depWithContainerAndContext](c, ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
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

func TestResolve_CircularDependenciesReturnsError(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, Transient, func(b *depB) *depA { return &depA{name: b.name} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}
	if err := Register[*depB](c, Transient, func(a *depA) *depB { return &depB{name: a.name} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_, err := Resolve[*depA](c, ctx)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Fatalf("expected circular dependency error, got: %v", err)
	}
}

func TestResolve_UnregisteredServiceReturnsError(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	_, err := Resolve[*depA](c, ctx)
	if err == nil {
		t.Fatal("expected error when resolving unregistered service")
	}
}

func TestResolve_UnregisteredDependencyReturnsError(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depC](c, Transient, func(a *depA, b *depB) *depC { return &depC{a: a, b: b} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_, err := Resolve[*depC](c, ctx)
	if err == nil {
		t.Fatal("expected error when dependency is not registered")
	}
}

func TestResolveWithKey_CustomKey(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := RegisterWithKey[*depA](c, "custom.key", Transient, func() *depA { return &depA{name: "custom"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance, err := ResolveWithKey[*depA](c, "custom.key", ctx)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if instance == nil || instance.name != "custom" {
		t.Fatal("expected to resolve instance by custom key")
	}
}

func TestResolve_NilContainerReturnsError(t *testing.T) {
	_, err := Resolve[*depA](nil, nil)
	if err == nil {
		t.Fatal("expected error when container is nil")
	}
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

	first, err := Resolve[*depA](c, nil)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	second, err := Resolve[*depA](c, nil)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if first != second {
		t.Fatal("expected background context singleton to be reused")
	}
	if created != 1 {
		t.Fatalf("expected factory to be called once, got %d", created)
	}
}

func TestResolve_ContainerSelf(t *testing.T) {
	c := NewContainer()

	got, err := Resolve[Container](c, nil)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if got == nil {
		t.Fatal("expected to resolve container instance")
	}
	if got != c {
		t.Fatal("expected resolved container to be the same instance")
	}
}

func TestResolveWithKey_EmptyKeyReturnsError(t *testing.T) {
	c := NewContainer()

	_, err := ResolveWithKey[*depA](c, " ", nil)
	if err == nil {
		t.Fatal("expected error when key is empty")
	}
}

func TestResolveWithKey_TypeMismatchReturnsError(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := RegisterWithKey[*depA](c, "mismatch.key", Transient, func() *depA { return &depA{name: "a"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_, err := ResolveWithKey[*depB](c, "mismatch.key", ctx)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
	if !strings.Contains(err.Error(), "not of type") {
		t.Fatalf("expected type mismatch error, got: %v", err)
	}
}

func TestMustResolve_PanicsOnNilContainer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when container is nil")
		}
	}()
	_ = MustResolve[*depA](nil, nil)
}

func TestMustResolveWithKey_PanicsOnEmptyKey(t *testing.T) {
	c := NewContainer()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when key is empty")
		}
	}()
	_ = MustResolveWithKey[*depA](c, " ", nil)
}

func TestMustResolveWithKey_PanicsOnResolveError(t *testing.T) {
	c := NewContainer()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when resolve fails")
		}
	}()
	_ = MustResolveWithKey[*depA](c, "missing.key", nil)
}

func TestMustResolve_Succeeds(t *testing.T) {
	c := NewContainer()
	ctx := c.NewContext()

	if err := Register[*depA](c, Transient, func() *depA { return &depA{name: "ok"} }); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	instance := MustResolve[*depA](c, ctx)
	if instance == nil || instance.name != "ok" {
		t.Fatal("expected to resolve instance")
	}
}
