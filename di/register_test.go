package di

import (
	"testing"

	diutils "github.com/lcrux/go-di/di/di-utils"
)

func TestRegister_NilContainerReturnsError(t *testing.T) {
	if err := Register[*depA](nil, Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error when container is nil")
	}
}

func TestRegisterWithKey_EmptyKeyReturnsError(t *testing.T) {
	c := NewContainer()

	if err := RegisterWithKey[*depA](c, " ", Transient, func() *depA { return &depA{} }); err == nil {
		t.Fatal("expected error when key is empty")
	}
}

func TestRegisterValidation(t *testing.T) {
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
