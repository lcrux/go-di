package di

import (
	"reflect"
	"testing"
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

	if resolved == nil {
		t.Fatalf("Expected non-nil resolved service with dependency")
	}
	if err != nil {
		t.Fatalf("Expected no error when resolving service with dependency, got: %v", err)
	}
	if resolved.Dep == nil {
		t.Fatal("Expected dependency to be resolved")
	}

	typOfDep := TypeOf[*Dependency]()
	typOfResolvedDep := reflect.TypeOf(resolved.Dep)

	if typOfResolvedDep != typOfDep {
		t.Errorf("Expected resolved dependency type to match %v, got %v", typOfDep, typOfResolvedDep)
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	type A struct{ b interface{} }
	type B struct{ a interface{} }

	_ = Register[*A](func(b *B) *A {
		return &A{b: b}
	}, Singleton)

	_ = Register[*B](func(a *A) *B {
		return &B{a: a}
	}, Singleton)

	_, err := Resolve[*A]()

	if err == nil {
		t.Fatal("Expected error due to circular dependency")
	}
}
