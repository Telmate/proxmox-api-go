package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LxcMountSize_String(t *testing.T) {
	tests := []struct {
		name   string
		input  LxcMountSize
		output string
	}{
		{name: "Kibibyte",
			input:  kibiByte,
			output: "1K"},
		{name: "Mebibyte",
			input:  mebiByte,
			output: "1M"},
		{name: "Gibibyte",
			input:  gibiByte,
			output: "1G"},
		{name: "Tebibyte",
			input:  tebiByte,
			output: "1T"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}
