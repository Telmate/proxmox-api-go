package array

import (
	"fmt"
	"strings"
)

// Combine two arrays of comparable types into one, removing duplicates.
// The output order is not guaranteed.
func Combine[T comparable](a, b []T) []T {
	combinedMap := make(map[T]struct{})
	for i := range a {
		combinedMap[a[i]] = struct{}{}
	}
	for i := range b {
		combinedMap[b[i]] = struct{}{}
	}
	combined := make([]T, len(combinedMap))
	var index int
	for e := range combinedMap {
		combined[index] = e
		index++
	}
	return combined
}

// CSV converts an array of fmt.Stringer to a CSV string
func CSV[T fmt.Stringer](array []T) string {
	switch len(array) {
	case 0:
		return ""
	case 1:
		return array[0].String()
	}
	builder := strings.Builder{}
	for i := range array {
		builder.WriteString("," + array[i].String())
	}
	return builder.String()[1:]
}

func Empty[T any]() []T { return []T{} }

func Nil[T any]() []T { return nil }

// RemoveItem removes all occurrences of the specified item from the array.
func RemoveItem[T comparable](array []T, remove T) []T {
	var length int
	for i := range array {
		if array[i] != remove {
			length++
		}
	}
	var index int
	newArray := make([]T, length)
	for i := range array {
		if array[i] != remove {
			newArray[index] = array[i]
			index++
		}
	}
	return newArray
}

// RemoveItems removes all occurrences of the specified items from the array.
// Duplicate items are reduced to a single instance.
// The order of the remaining items is not guaranteed.
func RemoveItems[T comparable](array []T, remove []T) []T {
	removeMap := make(map[T]struct{})
	for i := range array {
		removeMap[array[i]] = struct{}{}
	}
	for i := range remove {
		delete(removeMap, remove[i])
	}
	result := make([]T, len(removeMap))
	var index int
	for e := range removeMap {
		result[index] = e
		index++
	}
	return result
}
