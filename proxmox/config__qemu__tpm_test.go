package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_TpmState_Validate(t *testing.T) {
	type testInput struct {
		config  TpmState
		current *TpmState
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Invalid Storage Create`, input: testInput{
			config: TpmState{Storage: ""}},
			output: errors.New("storage is required")},
		{name: `Invalid Storage Update`, input: testInput{
			config:  TpmState{Storage: ""},
			current: &TpmState{Storage: "local-lvm"}},
			output: errors.New("storage is required")},
		{name: `Invalid Version=nil Create`, input: testInput{
			config: TpmState{Storage: "local-lvm"}},
			output: errors.New(TmpState_Error_VersionRequired)},
		{name: `Invalid Version="" Create`, input: testInput{
			config: TpmState{Storage: "local-lvm", Version: util.Pointer(TpmVersion(""))}},
			output: errors.New(TpmVersion_Error_Invalid)},
		{name: `Invalid Version="" Update`, input: testInput{
			config:  TpmState{Storage: "local-lvm", Version: util.Pointer(TpmVersion(""))},
			current: &TpmState{Storage: "local-lvm", Version: util.Pointer(TpmVersion("v2.0"))}},
			output: errors.New(TpmVersion_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_TpmVersion_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  TpmVersion
		output error
	}{
		{name: "Valid v1.2", input: TpmVersion_1_2},
		{name: "Valid v2.0", input: TpmVersion_2_0},
		{name: "Valid 1.2", input: "1.2"},
		{name: "Valid 2", input: "2"},
		{name: "Valid 2.0", input: "2.0"},
		{name: "Valid v2", input: "v2"},
		{name: `Invalid ""`, output: errors.New(TpmVersion_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
