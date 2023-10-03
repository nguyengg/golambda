package mutil

// First returns the first key-value entry from looping over the map.
//
// If the map is empty, the bool value will be false.
//
// Useful if the map contains only one entry. If it has more than one, this method may return different entries upon
// subsequent calls. If the map is empty, empty string and zero-value V are returned.
func First[V any](m map[string]V) (string, V, bool) {
	for k, v := range m {
		return k, v, true
	}

	var v V
	return "", v, false
}
