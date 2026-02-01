package proxmox

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_rawAgentNetworkInterfaces_SelectMacAddress(t *testing.T) {
	t.Parallel()
	parseMAC := func(mac string) net.HardwareAddr {
		parsedMac, _ := net.ParseMAC(mac)
		return parsedMac
	}
	testInput := func() map[string]any {
		return map[string]any{
			"result": []any{
				map[string]any{
					"name": string("lo")},
				map[string]any{
					"hardware-address": string("7a:b1:8f:2e:4d:6c"),
					"name":             string("eth0")},
				map[string]any{
					"hardware-address": string("1a:2b:3c:4d:5e:6f"),
					"name":             string("eth1")}}}
	}
	tests := []struct {
		name   string
		input  map[string]any
		mac    net.HardwareAddr
		output AgentNetworkInterface
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
			output: AgentNetworkInterface{
				MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
				Name:       "eth1"}},
		{name: `no interfaces`,
			input: map[string]any{
				"result": []any{}},
			mac: parseMAC("1a:2b:3c:4d:5e:6f"),
			set: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, outBool := RawAgentNetworkInterfaces(&rawAgentNetworkInterfaces{a: test.input}).SelectMacAddress(test.mac)
			require.Equal(t, test.set, outBool)
			if outBool {
				require.Equal(t, test.output, tmpOutput.Get())
			} else {
				require.Nil(t, tmpOutput)
			}
		})
	}
}

func Test_rawAgentNetworkInterfaces_SelectName(t *testing.T) {
	t.Parallel()
	parseMAC := func(mac string) net.HardwareAddr {
		parsedMac, _ := net.ParseMAC(mac)
		return parsedMac
	}
	testInput := func() map[string]any {
		return map[string]any{
			"result": []any{
				map[string]any{
					"name": string("lo")},
				map[string]any{
					"hardware-address": string("7a:b1:8f:2e:4d:6c"),
					"name":             string("eth0")},
				map[string]any{
					"hardware-address": string("1a:2b:3c:4d:5e:6f"),
					"name":             string("eth1")}}}
	}
	tests := []struct {
		name   string
		input  map[string]any
		iName  string
		output AgentNetworkInterface
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
			output: AgentNetworkInterface{
				MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
				Name:       "eth1"}},
		{name: `no interfaces`,
			input: map[string]any{
				"result": []any{}},
			iName: "eth1",
			set:   false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpOutput, outBool := RawAgentNetworkInterfaces(&rawAgentNetworkInterfaces{a: test.input}).SelectName(test.iName)
			require.Equal(t, test.set, outBool)
			if outBool {
				require.Equal(t, test.output, tmpOutput.Get())
			} else {
				require.Nil(t, tmpOutput)
			}
		})
	}
}

func Test_rawAgentNetworkInterfaces_Get(t *testing.T) {
	t.Parallel()
	parseMAC := func(mac string) net.HardwareAddr {
		parsedMac, _ := net.ParseMAC(mac)
		return parsedMac
	}
	parseCIDR := func(cidr string) (ip net.IP) {
		ip, _, _ = net.ParseCIDR(cidr)
		return
	}
	baseInput := func(params []any) map[string]any {
		return map[string]any{"result": params}
	}
	tests := []struct {
		name   string
		input  map[string]any
		output []AgentNetworkInterface
	}{
		{name: `Empty`,
			input:  baseInput([]any{}),
			output: nil},
		{name: `IpAddresses Empty`,
			input: baseInput([]any{map[string]any{
				"ip-addresses": []any{}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{}}}},
		{name: `IpAddresses Single`,
			input: baseInput([]any{map[string]any{
				"ip-addresses": []any{map[string]any{
					"ip-address": string("127.0.0.1"),
					"prefix":     float64(8)}}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8")}}}},
		{name: `IpAddresses multiple`,
			input: baseInput([]any{map[string]any{
				"ip-addresses": []any{
					map[string]any{
						"ip-address": string("127.0.0.1"),
						"prefix":     float64(8)},
					map[string]any{
						"ip-address": string("::1"),
						"prefix":     float64(128)}}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8"),
				parseCIDR("::1/128")}}}},
		{name: `MacAddress`,
			input: baseInput([]any{map[string]any{
				"hardware-address": string("54:1a:12:8f:7b:ed")}}),
			output: []AgentNetworkInterface{{MacAddress: parseMAC("54:1a:12:8f:7b:ed")}}},
		{name: `Name`,
			input: baseInput([]any{map[string]any{
				"name": "test"}}),
			output: []AgentNetworkInterface{{Name: string("test")}}},
		{name: `Statistics`,
			input: baseInput([]any{map[string]any{
				"statistics": map[string]any{
					"rx-bytes":   float64(1),
					"rx-packets": float64(2),
					"rx-errs":    float64(3),
					"rx-dropped": float64(4),
					"tx-bytes":   float64(5),
					"tx-packets": float64(6),
					"tx-errs":    float64(7),
					"tx-dropped": float64(8)}}}),
			output: []AgentNetworkInterface{{Statistics: &AgentInterfaceStatistics{
				RxBytes:   1,
				RxPackets: 2,
				RxErrors:  3,
				RxDropped: 4,
				TxBytes:   5,
				TxPackets: 6,
				TxErrors:  7,
				TxDropped: 8}}}},
		{name: `Full true`,
			input: baseInput(
				[]any{
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
						"name": string("eth1"),
						"statistics": map[string]any{
							"rx-bytes":   float64(8),
							"rx-packets": float64(7),
							"rx-errs":    float64(6),
							"rx-dropped": float64(5),
							"tx-bytes":   float64(4),
							"tx-packets": float64(3),
							"tx-errs":    float64(2),
							"tx-dropped": float64(1)}}}),
			output: []AgentNetworkInterface{
				{Name: string("lo"),
					IpAddresses: []net.IP{
						parseCIDR("127.0.0.1/8"),
						parseCIDR("::1/128")}},
				{Name: string("eth0"),
					MacAddress: parseMAC("7a:b1:8f:2e:4d:6c")},
				{Name: string("eth1"),
					MacAddress: parseMAC("1a:2b:3c:4d:5e:6f"),
					IpAddresses: []net.IP{
						parseCIDR("2001:0db8:85a3:0000:0000:8a2e:0370:7334/64"),
						parseCIDR("192.168.0.1/24"),
						parseCIDR("10.20.30.244/16")},
					Statistics: &AgentInterfaceStatistics{
						RxBytes:   8,
						RxPackets: 7,
						RxErrors:  6,
						RxDropped: 5,
						TxBytes:   4,
						TxPackets: 3,
						TxErrors:  2,
						TxDropped: 1}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawAgentNetworkInterfaces(&rawAgentNetworkInterfaces{a: test.input}).Get())
		})
	}
}
