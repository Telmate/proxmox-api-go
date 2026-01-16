package array

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Combine(t *testing.T) {
	tests := []struct {
		name   string
		inputA []string
		inputB []string
		output map[string]struct{}
	}{
		{name: `both empty`,
			inputA: []string{},
			inputB: []string{},
			output: map[string]struct{}{}},
		{name: `first empty`,
			inputA: []string{},
			inputB: []string{"a", "b", "c"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{name: `second empty`,
			inputA: []string{"a", "b", "c"},
			inputB: []string{},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{name: `no overlap`,
			inputA: []string{"a", "b", "c"},
			inputB: []string{"d", "e", "f"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}, "e": {}, "f": {}}},
		{name: `with overlap`,
			inputA: []string{"a", "b", "c"},
			inputB: []string{"b", "c", "d"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}}},
		{name: `with duplicates`,
			inputA: []string{"a", "b", "b", "c", "c", "c"},
			inputB: []string{"c", "d", "d", "e", "e", "e"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}, "e": {}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			result := Combine(test.inputA, test.inputB)
			resultMap := make(map[string]struct{}, len(result))
			for i := range result {
				resultMap[result[i]] = struct{}{}
			}
			require.Equal(t, test.output, resultMap)
		})
	}
}

type stringer string

func (s stringer) String() string {
	return string(s)
}

func Test_CSV(t *testing.T) {
	tests := []struct {
		name   string
		input  []stringer
		output string
	}{
		{name: `single items`,
			input:  []stringer{"a"},
			output: "a"},
		{name: `multiple items`,
			input: []stringer{
				"a",
				"b",
				"c",
			},
			output: "a,b,c"},
		{name: `empty`,
			input:  []stringer{},
			output: ""},
		{name: `big test`,
			input: func() []stringer {
				var arr []stringer
				for i := range 1000 {
					arr = append(arr, stringer("item"+strconv.Itoa(i)))
				}
				return arr
			}(),
			output: func() string {
				var result string
				for i := range 1000 {
					result += ",item" + strconv.Itoa(i)
				}
				return result[1:]
			}()},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, CSV(test.input))
		})
	}
}

func Test_Map(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output map[string]struct{}
	}{
		{name: "empty input",
			input:  []string{},
			output: map[string]struct{}{}},
		{name: "single element",
			input:  []string{"a"},
			output: map[string]struct{}{"a": {}}},
		{name: "multiple elements",
			input:  []string{"a", "b", "c"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{name: "multiple elements with duplicates",
			input:  []string{"a", "b", "c", "a", "b"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			result := Map(test.input)
			require.Len(t, result, len(test.output))
			require.Equal(t, test.output, result)
		})
	}
}

func Test_RemoveItem(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		remove string
		output []string
	}{
		{name: `single items`,
			input:  []string{"a"},
			remove: "a",
			output: []string{}},
		{name: `multiple items`,
			input:  []string{"a", "b", "c", "b", "d"},
			remove: "b",
			output: []string{"a", "c", "d"}},
		{name: `no items removed`,
			input:  []string{"a", "b", "c"},
			remove: "d",
			output: []string{"a", "b", "c"}},
		{name: `all items removed`,
			input:  []string{"a", "a", "a"},
			remove: "a",
			output: []string{}},
		{name: `empty input`,
			input:  []string{},
			remove: "a",
			output: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, RemoveItem(test.input, test.remove))
		})
	}
}

func Test_RemoveItems(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		remove []string
		output map[string]struct{}
	}{
		{name: `single items`,
			input:  []string{"a"},
			remove: []string{"a"},
			output: map[string]struct{}{}},
		{name: `multiple items`,
			input:  []string{"a", "b", "c", "a", "d"},
			remove: []string{"b", "d"},
			output: map[string]struct{}{"a": {}, "c": {}}},
		{name: `no items removed`,
			input:  []string{"a", "b", "c"},
			remove: []string{"d"},
			output: map[string]struct{}{"a": {}, "b": {}, "c": {}}},
		{name: `all items removed`,
			input:  []string{"a", "a", "a"},
			remove: []string{"a"},
			output: map[string]struct{}{}},
		{name: `empty input`,
			input:  []string{},
			remove: []string{"a"},
			output: map[string]struct{}{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			result := RemoveItems(test.input, test.remove)
			resultMap := make(map[string]struct{}, len(result))
			for i := range result {
				resultMap[result[i]] = struct{}{}
			}
			require.Equal(t, test.output, resultMap)
		})
	}
}
