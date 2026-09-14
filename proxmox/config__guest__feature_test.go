package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GuestFeature_mapToStruct(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  map[string]any
		output bool
	}{
		{name: "false",
			input:  map[string]any{"hasFeature": float64(0)},
			output: false},
		{name: "not set",
			input:  map[string]any{},
			output: false},
		{name: "true",
			input:  map[string]any{"hasFeature": float64(1)},
			output: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, GuestFeature("").mapToStruct(test.input), test.name)
		})
	}
}

func Test_GuestFeature_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input GuestFeature
		err   error
	}{
		// Invalid
		{name: "Invalid empty",
			input: "",
			err:   GuestFeature("").Error()},
		{name: "Invalid not enum",
			input: "invalid",
			err:   GuestFeature("").Error()},
		// Valid
		{name: "Valid GuestFeature_Clone",
			input: GuestFeatureClone},
		{name: "Valid GuestFeature_Copy",
			input: GuestFeatureCopy},
		{name: "Valid GuestFeature_Snapshot",
			input: GuestFeatureSnapshot},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.err, test.input.Validate(), test.name)
		})
	}
}
