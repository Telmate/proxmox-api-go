package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_guest"
	"github.com/stretchr/testify/require"
)

func Test_GuestName_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid GuestName`,
			input:  test_data_guest.GuestNameLegal(),
			output: nil},
		{name: `Invalid GuestName Empty`,
			input:  []string{test_data_guest.GuestNameEmpty()},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid GuestName Invalid`,
			input:  test_data_guest.GuestNameCharacterIllegal(),
			output: errors.New(GuestNameErrorInvalid)},
		{name: `Invalid GuestName Max Length`,
			input:  []string{test_data_guest.GuestNameMaxIllegal()},
			output: errors.New(GuestNameErrorLength)},
		{name: `Invalid GuestName begin with illegal end character`,
			input:  test_data_guest.GuestNameEndIllegal(),
			output: errors.New(GuestNameErrorEnd)},
		{name: `Invalid GuestName begin with illegal start character`,
			input:  test_data_guest.GuestNameStartIllegal(),
			output: errors.New(GuestNameErrorStart)},
	}
	for _, test := range tests {
		for _, e := range test.input {
			t.Run(test.name+"/"+e, func(t *testing.T) {
				require.Equal(t, test.output, (GuestName(e)).Validate())
			})
		}
	}
}

func Test_GuestNetworkRate_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  GuestNetworkRate
		output error
	}{
		{name: `Valid maximum`,
			input: 10240000},
		{name: `Valid minimum`,
			input: 0},
		{name: `Invalid`,
			input:  10240001,
			output: errors.New(GuestNetworkRate_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_GuestFeature_mapToStruct(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		output bool
	}{
		{name: "false",
			input:  map[string]interface{}{"hasFeature": float64(0)},
			output: false,
		},
		{name: "not set",
			input:  map[string]interface{}{},
			output: false,
		},
		{name: "true",
			input:  map[string]interface{}{"hasFeature": float64(1)},
			output: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, GuestFeature("").mapToStruct(test.input), test.name)
		})
	}
}

func Test_GuestFeature_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input GuestFeature
		err   error
	}{
		// Invalid
		{name: "Invalid empty",
			input: "",
			err:   GuestFeature("").Error(),
		},
		{name: "Invalid not enum",
			input: "invalid",
			err:   GuestFeature("").Error(),
		},
		// Valid
		{name: "Valid GuestFeature_Clone",
			input: GuestFeature_Clone,
		},
		{name: "Valid GuestFeature_Copy",
			input: GuestFeature_Copy,
		},
		{name: "Valid GuestFeature_Snapshot",
			input: GuestFeature_Snapshot,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.err, test.input.Validate(), test.name)
		})
	}
}

func Test_GuestID_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input GuestID
		err   error
	}{
		// Invalid
		{name: "Invalid to big (Maximum)",
			input: GuestID(GuestIdMaximum + 1),
			err:   errors.New(GuestID_Error_Maximum)},
		{name: "Invalid to small (Minimum)",
			input: GuestID(GuestIdMinimum - 1),
			err:   errors.New(GuestID_Error_Minimum)},
		// Valid
		{name: "Valid Maximum",
			input: GuestID(GuestIdMaximum)},
		{name: "Valid Minimum",
			input: GuestID(GuestIdMinimum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.err, test.input.Validate(), test.name)
		})
	}
}

func Test_GuestType_Parse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output GuestType
		err    error
	}{
		{name: "Valid qemu",
			input:  "qemu",
			output: GuestQemu},
		{name: "Valid lxc",
			input:  "lxc",
			output: GuestLxc},
		{name: "Invalid empty",
			input:  "",
			output: guestUnknown,
			err:    errors.New(GuestType_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result GuestType
			err := result.Parse(test.input)
			require.Equal(t, test.output, result)
			require.Equal(t, test.err, err)
		})
	}
}

func Test_GuestType_parse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output GuestType
	}{
		{name: "Valid qemu",
			input:  "qemu",
			output: GuestQemu},
		{name: "Valid lxc",
			input:  "lxc",
			output: GuestLxc},
		{name: "Invalid empty",
			input:  "",
			output: guestUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result GuestType
			result.parse(test.input)
			require.Equal(t, test.output, result)
		})
	}
}

func Test_GuestType_String(t *testing.T) {
	tests := []struct {
		name   string
		input  GuestType
		output string
	}{
		{name: "Valid qemu",
			input:  GuestQemu,
			output: "qemu"},
		{name: "Valid lxc",
			input:  GuestLxc,
			output: "lxc"},
		{name: "Invalid unknown",
			input:  guestUnknown,
			output: "unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.String())
		})
	}
}
