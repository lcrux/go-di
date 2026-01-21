package di

import (
	"reflect"
	"testing"

	"github.com/lcrux/go-di/v0"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAndResolve(t *testing.T) {
	type TestService struct{}

	factoryFn := func() *TestService {
		return &TestService{}
	}

	err := di.Register[*TestService](factoryFn, di.Singleton)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	resolved, err := di.Resolve[*TestService]()
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

	_ = di.Register[*Dependency](func() *Dependency {
		return &Dependency{}
	}, di.Singleton)

	_ = di.Register[*ServiceWithDependency](func(dep *Dependency) *ServiceWithDependency {
		return &ServiceWithDependency{Dep: dep}
	}, di.Singleton)
	resolved, err := di.Resolve[*ServiceWithDependency]()

	assert.NotNil(t, resolved, "Expected non-nil resolved service with dependency")
	assert.NoError(t, err, "Expected no error when resolving service with dependency")
	assert.NotNil(t, resolved.Dep, "Expected dependency to be resolved")

	typOfDep := di.TypeOf[*Dependency]()
	typOfResolvedDep := reflect.TypeOf(resolved.Dep)

	assert.Equal(t, typOfResolvedDep, typOfDep, "Expected resolved dependency type to match")
}

func TestCircularDependencyDetection(t *testing.T) {
	type A struct{}
	type B struct{}

	_ = di.Register[*A](func(b *B) *A {
		return &A{}
	}, di.Singleton)

	_ = di.Register[*B](func(a *A) *B {
		return &B{}
	}, di.Singleton)

	_, err := di.Resolve[*A]()

	assert.Error(t, err, "Expected error due to circular dependency")
}
