package mapUtil

// Difference returns:
//   - keys in a but not in b
//   - keys in b but not in a
//
// The map values are ignored; maps are treated as sets.
func Difference[K comparable, V1, V2 any](a map[K]V1, b map[K]V2) (onlyInA, onlyInB []K) {

	// A - B
	for k := range a {
		if _, ok := b[k]; !ok {
			onlyInA = append(onlyInA, k)
		}
	}

	// B - A
	for k := range b {
		if _, ok := a[k]; !ok {
			onlyInB = append(onlyInB, k)
		}
	}

	return
}

func SameKeys[K comparable, V1, V2 any](a map[K]V1, b map[K]V2) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func Array[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}
