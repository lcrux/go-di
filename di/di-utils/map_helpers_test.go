package diutils

import "testing"

func TestGetMapKeys_NilMapReturnsEmpty(t *testing.T) {
	keys := GetMapKeys[string, int](nil)
	if keys == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}
}

func TestGetMapValues_NilMapReturnsEmpty(t *testing.T) {
	values := GetMapValues[string, int](nil)
	if values == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(values) != 0 {
		t.Fatalf("expected 0 values, got %d", len(values))
	}
}

func TestGetMapKeysAndValues_ReturnsAllEntries(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}

	keys := GetMapKeys(m)
	values := GetMapValues(m)

	if len(keys) != len(m) {
		t.Fatalf("expected %d keys, got %d", len(m), len(keys))
	}
	if len(values) != len(m) {
		t.Fatalf("expected %d values, got %d", len(m), len(values))
	}

	keySet := map[string]struct{}{}
	for _, k := range keys {
		keySet[k] = struct{}{}
	}
	for k := range m {
		if _, ok := keySet[k]; !ok {
			t.Fatalf("missing key %s", k)
		}
	}

	valueSet := map[int]int{}
	for _, v := range values {
		valueSet[v]++
	}
	for _, v := range m {
		if valueSet[v] == 0 {
			t.Fatalf("missing value %d", v)
		}
	}
}
