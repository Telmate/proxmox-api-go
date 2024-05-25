package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_tag"
	"github.com/stretchr/testify/require"
)

func Test_Tag_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid Tag`,
			input:  test_data_tag.Tag_Legal(),
			output: nil,
		},
		{name: `Invalid Tag`,
			input:  test_data_tag.Tag_Character_Illegal(),
			output: errors.New(Tag_Error_Invalid),
		},
		{name: `Invalid Tag Empty`,
			input:  []string{test_data_tag.Tag_Empty()},
			output: errors.New(Tag_Error_Empty),
		},
		{name: `Invalid Tag Max Length`,
			input:  []string{test_data_tag.Tag_Max_Illegal()},
			output: errors.New(Tag_Error_MaxLength),
		},
	}
	for _, test := range tests {
		for _, e := range test.input {
			t.Run(test.name+": "+e, func(t *testing.T) {
				require.Equal(t, test.output, (Tag(e)).Validate())
			})
		}
	}
}
