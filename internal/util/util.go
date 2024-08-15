package util

// Gets inlined by the compiler, so it's not a performance hit
func Pointer[T any](item T) *T {
	return &item
}
