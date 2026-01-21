package di

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterAndResolve(t *testing.T) {
	type TestService struct{}

	factoryFn := func() *TestService {
		return &TestService{}
	}

	err := Register[*TestService](factoryFn, Singleton)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	resolved, err := Resolve[*TestService]()
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

	_ = Register[*Dependency](func() *Dependency {
		return &Dependency{}
	}, Singleton)

	_ = Register[*ServiceWithDependency](func(dep *Dependency) *ServiceWithDependency {
		return &ServiceWithDependency{Dep: dep}
	}, Singleton)
	resolved, err := Resolve[*ServiceWithDependency]()

	assert.NotNil(t, resolved, "Expected non-nil resolved service with dependency")
	assert.NoError(t, err, "Expected no error when resolving service with dependency")
	assert.NotNil(t, resolved.Dep, "Expected dependency to be resolved")

	typOfDep := TypeOf[*Dependency]()
	typOfResolvedDep := reflect.TypeOf(resolved.Dep)

	assert.Equal(t, typOfResolvedDep, typOfDep, "Expected resolved dependency type to match")
}

func TestCircularDependencyDetection(t *testing.T) {
	type A struct{}
	type B struct{}

	_ = Register[*A](func(b *B) *A {
		return &A{}
	}, Singleton)

	_ = Register[*B](func(a *A) *B {
		return &B{}
	}, Singleton)

	_ = Register[*B](func(a *A) *B {
		return &B{}
	}, Singleton)

	_ = Register[*B](func(a *A) *B {
		return &B{}
	}, Singleton)

	_, err := Resolve[*A]()

	assert.Error(t, err, "Expected error due to circular dependency")
}
