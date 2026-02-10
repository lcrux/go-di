package diutils

import (
	"testing"
)

func TestNewMap(t *testing.T) {
	m := NewMap[string, int]()
	if m == nil {
		t.Fatal("NewMap returned nil")
	}
	if len(m.Keys()) != 0 {
		t.Fatal("NewMap should initialize an empty map")
	}
}

func TestMapSetAndGet(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("key1", 100)

	value, exists := m.Get("key1")
	if !exists {
		t.Fatal("Expected key1 to exist")
	}
	if value != 100 {
		t.Fatalf("Expected value 100, got %d", value)
	}

	_, exists = m.Get("key2")
	if exists {
		t.Fatal("Expected key2 to not exist")
	}
}

func TestMapDelete(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("key1", 100)
	m.Delete("key1")

	_, exists := m.Get("key1")
	if exists {
		t.Fatal("Expected key1 to be deleted")
	}
}

func TestMapKeys(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)

	keys := m.Keys()
	if len(keys) != 2 {
		t.Fatalf("Expected 2 keys, got %d", len(keys))
	}
}

func TestMapValues(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)

	values := m.Values()
	if len(values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(values))
	}
}

func TestMapCleanup(t *testing.T) {
	m := NewMap[string, int]()
	m.Set("key1", 100)
	m.Set("key2", 200)

	m.Cleanup()
	if len(m.Keys()) != 0 {
		t.Fatal("Expected map to be empty after Cleanup")
	}
}
