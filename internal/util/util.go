package util

func Pointer[T any](item T) *T {
	return &item
}
