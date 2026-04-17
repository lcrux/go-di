package di

import (
	"context"
	"sync"
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

func (l *listenerDep) EndLifecycle(_ ...context.Context) error {
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

func TestContainer_RemoveContext_ShutsDownLifecycleContext(t *testing.T) {
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

	if err := c.RemoveContext(ctx); err != nil {
		t.Fatalf("unexpected remove context error: %v", err)
	}

	if called != 1 {
		t.Fatalf("expected EndLifecycle to be called after RemoveContext, got %d", called)
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

func TestContainer_Shutdown_CanceledContextSkipsLifecycleEnd(t *testing.T) {
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

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	errs := c.Shutdown(cancelCtx)
	if len(errs) == 0 {
		t.Fatalf("expected at most 1 error, got %d", len(errs))
	}
	if called != 0 {
		t.Fatalf("expected EndLifecycle not to be called, got %d", called)
	}
}

// TestContainer_Shutdown_ConcurrentResolveNoPanic reproduces the nil-background-context window
// described in issue #41: a concurrent Resolve during Shutdown could call BackgroundContext()
// in the brief gap after the old map was replaced but before the new background context was set,
// receiving nil and then panicking in loadInstance.
func TestContainer_Shutdown_ConcurrentResolveNoPanic(t *testing.T) {
	for i := 0; i < 100; i++ {
		c := NewContainer()

		if err := Register[*depA](c, Singleton, func() *depA { return &depA{name: "a"} }); err != nil {
			t.Fatalf("unexpected register error: %v", err)
		}

		// Pre-warm the dependency tree cache to avoid the pre-existing dependencyTreeCache data race.
		if _, err := Resolve[*depA](c, nil); err != nil {
			t.Fatalf("unexpected resolve error during warm-up: %v", err)
		}

		// Spin up many goroutines calling Resolve concurrently with Shutdown so that at least
		// one is likely to hit the window between the map replacement and the Set call.
		var wg sync.WaitGroup
		for j := 0; j < 20; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Ignore errors — the service may not resolve after shutdown; what matters is no panic.
				_, _ = Resolve[*depA](c, nil)
			}()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Shutdown()
		}()

		wg.Wait()
	}
}

func TestContainer_Validate_IgnoresContainerAndContextDependencies(t *testing.T) {
	c := NewContainer()

	if err := Register[*depWithContainerAndContext](c, Transient, func(c Container, ctx LifecycleContext) *depWithContainerAndContext {
		return &depWithContainerAndContext{c: c, ctx: ctx}
	}); err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	if err := c.Validate(); err != nil {
		t.Fatalf("expected validation to ignore container and context dependencies, got: %v", err)
	}
}

// TestContainer_getDependencyTree_NoDataRace verifies that concurrent calls to getDependencyTree
// for the same key do not cause a data race on dependencyTreeCache.
func TestContainer_getDependencyTree_NoDataRace(t *testing.T) {
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

	const goroutines = 50
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := Resolve[*depC](c, nil); err != nil {
				t.Errorf("unexpected resolve error: %v", err)
			}
		}()
	}

	close(start)
	wg.Wait()
}
