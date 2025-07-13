package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_resourcemapping"
	"github.com/stretchr/testify/require"
)

func Test_ResourceMappingUsbID_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		err   error
	}{
		// Valid
		{name: "Valid", input: test_data_resourcemapping.ResourceMappingUsbID_Legal()},
		// Invalid
		{name: "Invalid ResourceMappingUsbID_Error_MinLength",
			input: test_data_resourcemapping.ResourceMappingUsbID_Min_Illegal(),
			err:   errors.New(ResourceMappingUsbID_Error_MinLength)},
		{name: "Invalid ResourceMappingUsbID_Error_MaxLength",
			input: []string{test_data_resourcemapping.ResourceMappingUsbID_Max_Illegal()},
			err:   errors.New(ResourceMappingUsbID_Error_MaxLength)},
		{name: "Invalid ResourceMappingUsbID_Error_Start",
			input: test_data_resourcemapping.ResourceMappingUsbID_Start_Illegal(),
			err:   errors.New(ResourceMappingUsbID_Error_Start)},
		{name: "Invalid ResourceMappingUsbID_Error_Invalid",
			input: test_data_resourcemapping.ResourceMappingUsbID_Character_Illegal(),
			err:   errors.New(ResourceMappingUsbID_Error_Invalid)},
	}
	for _, test := range tests {
		for _, snapshot := range test.input {
			t.Run(test.name+" :"+snapshot, func(*testing.T) {
				require.Equal(t, ResourceMappingUsbID(snapshot).Validate(), test.err, test.name+" :"+snapshot)
			})
		}
	}
}
