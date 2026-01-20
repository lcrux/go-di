package tests

import (
	"reflect"
	"testing"

	godi "github.com/lcrux/go-di"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAndResolve(t *testing.T) {
	type TestService struct{}

	factoryFn := func() *TestService {
		return &TestService{}
	}

	err := godi.Register[*TestService](factoryFn, godi.Singleton)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	resolved, err := godi.Resolve[*TestService]()
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	if resolved == nil {
		t.Fatal("Expected non-nil resolved service")
	}
}

func TestResolveWithDependencies(t *testing.T) {
	type Dependency struct{}
	type ServiceWithDependency struct {
		Dep *Dependency
	}

	_ = godi.Register[*Dependency](func() *Dependency {
		return &Dependency{}
	}, godi.Singleton)

	_ = godi.Register[*ServiceWithDependency](func(dep *Dependency) *ServiceWithDependency {
		return &ServiceWithDependency{Dep: dep}
	}, godi.Singleton)
	resolved, err := godi.Resolve[*ServiceWithDependency]()

	assert.NotNil(t, resolved, "Expected non-nil resolved service with dependency")
	assert.NoError(t, err, "Expected no error when resolving service with dependency")
	assert.NotNil(t, resolved.Dep, "Expected dependency to be resolved")

	typOfDep := godi.TypeOf[*Dependency]()
	typOfResolvedDep := reflect.TypeOf(resolved.Dep)

	assert.Equal(t, typOfResolvedDep, typOfDep, "Expected resolved dependency type to match")
}

func TestCircularDependencyDetection(t *testing.T) {
	type A struct{}
	type B struct{}

	_ = godi.Register[*A](func(b *B) *A {
		return &A{}
	}, godi.Singleton)

	_ = godi.Register[*B](func(a *A) *B {
		return &B{}
	}, godi.Singleton)

	_, err := godi.Resolve[*A]()

	assert.Error(t, err, "Expected error due to circular dependency")
}
