package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SerialID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  SerialID
		output error
	}{
		{name: `Valid`,
			input: 2},
		{name: `Invalid`,
			input:  4,
			output: errors.New(SerialID_Errors_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_SerialPort_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  SerialInterface
		output error
	}{
		{name: `Valid Path`,
			input: SerialInterface{Path: "/dev/ttyS0"}},
		{name: `Valid Path + Delete`,
			input: SerialInterface{Path: "/dev/ttyS0", Delete: true}},
		{name: `Valid Socket`,
			input: SerialInterface{Socket: true}},
		{name: `Valid Socket + Delete`,
			input: SerialInterface{Socket: true, Delete: true}},
		{name: `Invalid errors.New(SerialInterface_Errors_MutualExclusive)`,
			input:  SerialInterface{Path: "/dev/ttyS0", Socket: true},
			output: errors.New(SerialInterface_Errors_MutualExclusive)},
		{name: `Invalid errors.New(SerialInterface_Errors_Empty)`,
			input:  SerialInterface{},
			output: errors.New(SerialInterface_Errors_Empty)},
		{name: `Invalid errors.New(SerialPath_Errors_Invalid)`,
			input:  SerialInterface{Path: "ttyS0"},
			output: errors.New(SerialPath_Errors_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_SerialInterfaces_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  SerialInterfaces
		output error
	}{
		{name: `Valid`,
			input: SerialInterfaces{
				SerialID0: SerialInterface{Path: "/dev/ttyS0"},
				SerialID1: SerialInterface{Path: "/dev/ttyS1", Delete: true},
				SerialID2: SerialInterface{Socket: true},
				SerialID3: SerialInterface{Socket: true, Delete: true}}},
		{name: `Valid delete`,
			input: SerialInterfaces{
				SerialID1: SerialInterface{Delete: true}}},
		{name: `Invalid errors.New(SerialID_Errors_Invalid)`,
			input:  SerialInterfaces{4: SerialInterface{Delete: true}},
			output: errors.New(SerialID_Errors_Invalid)},
		{name: `Invalid errors.New(SerialPath_Errors_MutualExclusive)`,
			input:  SerialInterfaces{SerialID0: SerialInterface{Path: "/dev/ttyS0", Socket: true}},
			output: errors.New(SerialInterface_Errors_MutualExclusive)},
		{name: `Invalid errors.New(SerialInterface_Errors_Empty)`,
			input:  SerialInterfaces{SerialID0: SerialInterface{}},
			output: errors.New(SerialInterface_Errors_Empty)},
		{name: `Invalid errors.New(SerialPath_Errors_Invalid)`,
			input:  SerialInterfaces{SerialID0: SerialInterface{Path: "ttyS0"}},
			output: errors.New(SerialPath_Errors_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_SerialPath_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  SerialPath
		output error
	}{
		{name: `Valid`,
			input: "/dev/ttyS0"},
		{name: `Invalid errors.New(SerialPath_Errors_Invalid)`,
			input:  "ttyS0",
			output: errors.New(SerialPath_Errors_Invalid)},
		{name: `Invalid /dev/ only`,
			input:  "/dev/",
			output: errors.New(SerialPath_Errors_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
