package diutils

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	if m == nil {
		return make([]K, 0)
	}

	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	if m == nil {
		return make([]V, 0)
	}

	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
