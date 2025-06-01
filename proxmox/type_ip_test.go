package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IPv4Address_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv4Address
		output error
	}{
		{name: `Valid`,
			input: "192.168.45.1"},
		{name: "Valid empty"},
		{name: `Invalid is CIDR`,
			input:  "192.168.45.1/24",
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid is IPv6`,
			input:  "3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc",
			output: errors.New(IPv4Address_Error_Invalid)},
		{name: `Invalid is gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv4Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv4CIDR_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv4CIDR
		output error
	}{
		{name: `Valid`,
			input: "192.168.45.0/24"},
		{name: `Valid empty`},
		{name: `Invalid only IP no CIDR`,
			input:  "192.168.45.0",
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid is IPv6`,
			input:  "2001:0db8:85a3::/64",
			output: errors.New(IPv4CIDR_Error_Invalid)},
		{name: `Invalid gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv4CIDR_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv6Address_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv6Address
		output error
	}{
		{name: `Valid`,
			input: "3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"},
		{name: `Valid empty`},
		{name: `Invalid is CIDR`,
			input:  "2001:0db8:85a3::/64",
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid is IPv4`,
			input:  "192.168.45.0",
			output: errors.New(IPv6Address_Error_Invalid)},
		{name: `Invalid is gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv6Address_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_IPv6CIDR_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  IPv6CIDR
		output error
	}{
		{name: `Valid`,
			input: "2001:0db8:85a3::/64"},
		{name: `Valid empty`},
		{name: `Invalid only IP no CIDR`,
			input:  "2001:0db8:85a3::",
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid is IPv4`,
			input:  "192.168.45.0/24",
			output: errors.New(IPv6CIDR_Error_Invalid)},
		{name: `Invalid gibberish`,
			input:  "ABCDEFG123",
			output: errors.New(IPv6CIDR_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
