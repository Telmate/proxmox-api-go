package proxmox

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_rawLxcInfoNetworkInterfaces_SelectMacAddress(t *testing.T) {
	t.Parallel()
	testInput := func() []any {
		return []any{
			map[string]any{
				"name": string("lo")},
			map[string]any{
				"hardware-address": string("7a:b1:8f:2e:4d:6c"),
				"name":             string("eth0")},
			map[string]any{
				"hardware-address": string("1a:2b:3c:4d:5e:6f"),
				"name":             string("eth1")}}
	}
	tests := []struct {
		name   string
		input  []any
		mac    net.HardwareAddr
		output LxcInfoNetworkInterface
		set    bool
	}{
		{name: `missing`,
			input: testInput(),
			mac:   parseMAC("00:11:22:33:44:55"),
			set:   false},
		{name: `contains`,
			input: testInput(),
			mac:   parseMAC("1a:2b:3c:4d:5e:6f"),
			set:   true,
			output: LxcInfoNetworkInterface{
				MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
				Name:       "eth1"}},
		{name: `no interfaces`,
			input: []any{},
			mac:   parseMAC("1a:2b:3c:4d:5e:6f"),
			set:   false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, outBool := RawLxcInfoNetworkInterfaces(&rawLxcInfoNetworkInterfaces{a: test.input}).SelectMacAddress(test.mac)
			require.Equal(t, test.set, outBool)
			if outBool {
				require.Equal(t, test.output, tmpOutput.Get())
			} else {
				require.Nil(t, tmpOutput)
			}
		})
	}
}

func Test_rawLxcInfoNetworkInterfaces_SelectName(t *testing.T) {
	t.Parallel()
	testInput := func() []any {
		return []any{
			map[string]any{
				"name": string("lo")},
			map[string]any{
				"hardware-address": string("7a:b1:8f:2e:4d:6c"),
				"name":             string("eth0")},
			map[string]any{
				"hardware-address": string("1a:2b:3c:4d:5e:6f"),
				"name":             string("eth1")}}
	}
	tests := []struct {
		name   string
		input  []any
		iName  LxcNetworkName
		output LxcInfoNetworkInterface
		set    bool
	}{
		{name: `missing`,
			input: testInput(),
			iName: "eth2",
			set:   false},
		{name: `contains`,
			input: testInput(),
			iName: "eth1",
			set:   true,
			output: LxcInfoNetworkInterface{
				MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
				Name:       "eth1"}},
		{name: `no interfaces`,
			input: []any{},
			iName: "eth1",
			set:   false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, outBool := RawLxcInfoNetworkInterfaces(&rawLxcInfoNetworkInterfaces{a: test.input}).SelectName(test.iName)
			require.Equal(t, test.set, outBool)
			if outBool {
				require.Equal(t, test.output, tmpOutput.Get())
			} else {
				require.Nil(t, tmpOutput)
			}
		})
	}
}

func Test_rawLxcInfoNetworkInterfaces_Get(t *testing.T) {
	t.Parallel()
	parseCIDR := func(cidr string) (ip net.IP) {
		ip, _ = parseCIDR(cidr)
		return
	}
	tests := []struct {
		name   string
		input  []any
		output []LxcInfoNetworkInterface
	}{
		{name: `Empty`,
			input:  []any{},
			output: nil},
		{name: `IpAddresses Empty`,
			input: []any{map[string]any{
				"ip-addresses": []any{}}},
			output: []LxcInfoNetworkInterface{{IpAddresses: []net.IP{}}}},
		{name: `IpAddresses Single`,
			input: []any{map[string]any{
				"ip-addresses": []any{map[string]any{
					"ip-address": string("127.0.0.1"),
					"prefix":     float64(8)}}}},
			output: []LxcInfoNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8")}}}},
		{name: `IpAddresses multiple`,
			input: []any{map[string]any{
				"ip-addresses": []any{
					map[string]any{
						"ip-address": string("127.0.0.1"),
						"prefix":     float64(8)},
					map[string]any{
						"ip-address": string("::1"),
						"prefix":     float64(128)}}}},
			output: []LxcInfoNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8"),
				parseCIDR("::1/128")}}}},
		{name: `MacAddress`,
			input: []any{map[string]any{
				"hardware-address": string("54:1a:12:8f:7b:ed")}},
			output: []LxcInfoNetworkInterface{{MacAddress: parseMAC("54:1a:12:8f:7b:ed")}}},
		{name: `Name`,
			input: []any{map[string]any{
				"name": "test"}},
			output: []LxcInfoNetworkInterface{{Name: "test"}}},
		{name: `Full true`,
			input: []any{
				map[string]any{
					"ip-addresses": []any{
						map[string]any{"ip-address": string("127.0.0.1"), "prefix": float64(8)},
						map[string]any{"ip-address": string("::1"), "prefix": float64(128)}},
					"name": string("lo")},
				map[string]any{
					"hardware-address": string("7a:b1:8f:2e:4d:6c"),
					"name":             string("eth0")},
				map[string]any{
					"hardware-address": string("1a:2b:3c:4d:5e:6f"),
					"ip-addresses": []any{
						map[string]any{"ip-address": string("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), "prefix": float64(64)},
						map[string]any{"ip-address": string("192.168.0.1"), "prefix": float64(24)},
						map[string]any{"ip-address": string("10.20.30.244"), "prefix": float64(16)}},
					"name": string("eth1")}},
			output: []LxcInfoNetworkInterface{
				{Name: "lo",
					IpAddresses: []net.IP{
						parseCIDR("127.0.0.1/8"),
						parseCIDR("::1/128")}},
				{Name: "eth0",
					MacAddress: parseMAC("7a:b1:8f:2e:4d:6c")},
				{Name: "eth1",
					MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
					IpAddresses: []net.IP{
						parseCIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"),
						parseCIDR("192.168.0.1/24"),
						parseCIDR("10.20.30.244/16")}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawLxcInfoNetworkInterfaces(&rawLxcInfoNetworkInterfaces{a: test.input}).Get())
		})
	}
}
