package libutils

import (
	"reflect"
	"testing"
)

type sample struct{}

type sampleAlias sample

func TestTypeOf(t *testing.T) {
	expected := reflect.TypeOf((*sample)(nil)).Elem()

	if got := TypeOf[sample](); got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestTypeOf_Alias(t *testing.T) {
	expected := reflect.TypeOf((*sampleAlias)(nil)).Elem()
	if got := TypeOf[sampleAlias](); got != expected {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestNameOf_CustomType(t *testing.T) {
	got := NameOf[sample]()
	if got != "github.com/lcrux/go-di/lib_utils/sample" {
		t.Fatalf("expected github.com/lcrux/go-di/lib_utils/sample, got %s", got)
	}
}

func TestNameOf_Builtin(t *testing.T) {
	got := NameOf[int]()
	if got != "/int" {
		t.Fatalf("expected /int, got %s", got)
	}
}
