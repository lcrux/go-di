package diutils

import "sync"

type Map[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: make(map[K]V),
	}
}

func (m *Map[K, V]) Set(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = value
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

func (m *Map[K, V]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, key)
}

func (m *Map[K, V]) Keys() []K {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return getMapKeys(m.data)
}

func (m *Map[K, V]) Values() []V {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return getMapValues(m.data)
}

func (m *Map[K, V]) Cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data = make(map[K]V)
}

func getMapKeys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return make([]K, 0)
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getMapValues[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return make([]V, 0)
	}

	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
