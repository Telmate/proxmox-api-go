package mapUtil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Difference(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		a       map[string]struct{}
		b       map[string]struct{}
		onlyInA []string
		onlyInB []string
	}{
		{name: "both empty",
			a:       map[string]struct{}{},
			b:       map[string]struct{}{},
			onlyInA: []string{},
			onlyInB: []string{}},
		{name: "a empty",
			a:       map[string]struct{}{},
			b:       map[string]struct{}{"x": {}, "y": {}},
			onlyInA: []string{},
			onlyInB: []string{"x", "y"}},
		{name: "b empty",
			a:       map[string]struct{}{"x": {}, "y": {}},
			b:       map[string]struct{}{},
			onlyInA: []string{"x", "y"},
			onlyInB: []string{}},
		{name: "no difference",
			a:       map[string]struct{}{"x": {}, "y": {}, "z": {}},
			b:       map[string]struct{}{"x": {}, "y": {}, "z": {}},
			onlyInA: []string{},
			onlyInB: []string{}},
		{name: "partial overlap",
			a:       map[string]struct{}{"a": {}, "b": {}, "c": {}},
			b:       map[string]struct{}{"b": {}, "c": {}, "d": {}},
			onlyInA: []string{"a"},
			onlyInB: []string{"d"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			onlyInA, onlyInB := Difference(test.a, test.b)
			require.ElementsMatch(t, test.onlyInA, onlyInA)
			require.ElementsMatch(t, test.onlyInB, onlyInB)
		})
	}
}

func Test_SameKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		a      map[string]struct{}
		b      map[string]struct{}
		output bool
	}{
		{name: "both empty",
			a:      map[string]struct{}{},
			b:      map[string]struct{}{},
			output: true},
		{name: "a empty, b not",
			a:      map[string]struct{}{},
			b:      map[string]struct{}{"x": {}},
			output: false},
		{name: "a not empty, b empty",
			a:      map[string]struct{}{"x": {}},
			b:      map[string]struct{}{},
			output: false},
		{name: "same keys",
			a:      map[string]struct{}{"x": {}, "y": {}, "z": {}},
			b:      map[string]struct{}{"x": {}, "y": {}, "z": {}},
			output: true},
		{name: "different number of keys",
			a:      map[string]struct{}{"x": {}, "y": {}},
			b:      map[string]struct{}{"x": {}, "y": {}, "z": {}},
			output: false},
		{name: "same number, different keys",
			a:      map[string]struct{}{"a": {}, "b": {}, "c": {}},
			b:      map[string]struct{}{"x": {}, "y": {}, "z": {}},
			output: false},
		{name: "partial overlap",
			a:      map[string]struct{}{"a": {}, "b": {}, "c": {}},
			b:      map[string]struct{}{"b": {}, "c": {}, "d": {}},
			output: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, SameKeys(test.a, test.b))
		})
	}
}

func Test_Array(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  map[string]struct{}
		output []string
	}{
		{name: "empty map",
			input:  map[string]struct{}{},
			output: []string{}},
		{name: "single element",
			input:  map[string]struct{}{"a": {}},
			output: []string{"a"}},
		{name: "multiple elements",
			input:  map[string]struct{}{"a": {}, "b": {}, "c": {}},
			output: []string{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ElementsMatch(t, test.output, Array(test.input))
		})
	}
}
