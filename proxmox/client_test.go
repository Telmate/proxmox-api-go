package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Version_Greater(t *testing.T) {
	type input struct {
		a Version
		b Version
	}
	tests := []struct {
		name   string
		input  input
		output bool
	}{
		{"a > b 0", input{Version{1, 0, 0}, Version{0, 0, 0}}, true},
		{"a > b 1", input{Version{0, 1, 0}, Version{0, 0, 255}}, true},
		{"a > b 2", input{Version{1, 0, 0}, Version{0, 255, 255}}, true},
		{"a < b 0", input{Version{7, 4, 1}, Version{7, 4, 2}}, false},
		{"a < b 1", input{Version{0, 0, 255}, Version{0, 1, 0}}, false},
		{"a < b 2", input{Version{0, 255, 255}, Version{1, 0, 0}}, false},
		{"a = b", input{Version{0, 0, 0}, Version{0, 0, 0}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.a.Greater(test.input.b))
		})
	}
}

func Test_Version_mapToSDK(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		output Version
	}{
		{"empty", map[string]interface{}{}, Version{}},
		{"full", map[string]interface{}{"version": "1.2.3"}, Version{1, 2, 3}},
		{"invalid", map[string]interface{}{"version": ""}, Version{}},
		{"major", map[string]interface{}{"version": "1"}, Version{1, 0, 0}},
		{"partial", map[string]interface{}{"version": "1.2"}, Version{1, 2, 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, Version{}.mapToSDK(test.input))
		})
	}
}

func Test_Version_Smaller(t *testing.T) {
	type input struct {
		a Version
		b Version
	}
	tests := []struct {
		name   string
		input  input
		output bool
	}{
		{"a > b 0", input{Version{1, 0, 0}, Version{0, 0, 0}}, false},
		{"a > b 1", input{Version{0, 1, 0}, Version{0, 0, 255}}, false},
		{"a > b 2", input{Version{1, 0, 0}, Version{0, 255, 255}}, false},
		{"a < b 0", input{Version{7, 4, 1}, Version{7, 4, 2}}, true},
		{"a < b 1", input{Version{0, 0, 255}, Version{0, 1, 0}}, true},
		{"a < b 2", input{Version{0, 255, 255}, Version{1, 0, 0}}, true},
		{"a = b", input{Version{0, 0, 0}, Version{0, 0, 0}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.a.Smaller(test.input.b))
		})
	}
}
