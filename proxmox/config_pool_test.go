package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_pool"
	"github.com/stretchr/testify/require"
)

func Test_ConfigPool_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  ConfigPool
		output error
	}{
		{name: "Valid PoolName",
			input: ConfigPool{Name: PoolName(test_data_pool.PoolName_Legal())}},
		{name: "Invalid PoolName Empty",
			input:  ConfigPool{Name: ""},
			output: errors.New(PoolName_Error_Empty)},
		{name: "Invalid PoolName Length",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Max_Illegal())},
			output: errors.New(PoolName_Error_Length)},
		{name: "Invalid PoolName Characters",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Error_Characters()[0])},
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PoolName_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid PoolName`,
			input: test_data_pool.PoolName_Legals()},
		{name: `Invalid Empty`,
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Length`,
			input:  []string{test_data_pool.PoolName_Max_Illegal()},
			output: errors.New(PoolName_Error_Length)},
		{name: `Invalid Characters`,
			input:  test_data_pool.PoolName_Error_Characters(),
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, PoolName(input).Validate())
			}
		})
	}
}
