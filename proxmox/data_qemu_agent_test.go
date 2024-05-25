package proxmox

import (
	"net"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_AgentNetworkInterface_mapToSDK(t *testing.T) {
	parseMAC := func(mac string) net.HardwareAddr {
		parsedMac, _ := net.ParseMAC(mac)
		return parsedMac
	}
	parseCIDR := func(cidr string) (ip net.IP) {
		ip, _, _ = net.ParseCIDR(cidr)
		return
	}
	type testInput struct {
		params     map[string]interface{}
		statistics *bool // nil is false and true at the same time
	}
	baseInput := func(statistics *bool, params []interface{}) testInput {
		return testInput{
			params:     map[string]interface{}{"result": params},
			statistics: statistics}
	}
	inputFullTest := func() []interface{} {
		return []interface{}{
			map[string]interface{}{
				"hardware-address": string("54:1a:12:8f:7b:ed"),
				"ip-addresses": []interface{}{
					map[string]interface{}{"ip-address": string("127.0.0.1"), "prefix": float64(8)},
					map[string]interface{}{"ip-address": string("::1"), "prefix": float64(128)}},
				"name": string("lo")},
			map[string]interface{}{
				"hardware-address": string("7a:b1:8f:2e:4d:6c"),
				"name":             string("eth0")},
			map[string]interface{}{
				"hardware-address": string("1a:2b:3c:4d:5e:6f"),
				"ip-addresses": []interface{}{
					map[string]interface{}{"ip-address": string("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), "prefix": float64(64)},
					map[string]interface{}{"ip-address": string("192.168.0.1"), "prefix": float64(24)},
					map[string]interface{}{"ip-address": string("10.20.30.244"), "prefix": float64(16)}},
				"name": string("eth1"),
				"statistics": map[string]interface{}{
					"rx-bytes":   float64(8),
					"rx-packets": float64(7),
					"rx-errs":    float64(6),
					"rx-dropped": float64(5),
					"tx-bytes":   float64(4),
					"tx-packets": float64(3),
					"tx-errs":    float64(2),
					"tx-dropped": float64(1)}}}
	}
	tests := []struct {
		name   string
		input  testInput
		output []AgentNetworkInterface
	}{
		{name: `IpAddresses Empty`,
			input: baseInput(nil, []interface{}{map[string]interface{}{
				"ip-addresses": []interface{}{}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{}}}},
		{name: `IpAddresses Single`,
			input: baseInput(nil, []interface{}{map[string]interface{}{
				"ip-addresses": []interface{}{map[string]interface{}{
					"ip-address": string("127.0.0.1"),
					"prefix":     float64(8)}}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8")}}}},
		{name: `IpAddresses multiple`,
			input: baseInput(nil, []interface{}{map[string]interface{}{
				"ip-addresses": []interface{}{
					map[string]interface{}{
						"ip-address": string("127.0.0.1"),
						"prefix":     float64(8)},
					map[string]interface{}{
						"ip-address": string("::1"),
						"prefix":     float64(128)}}}}),
			output: []AgentNetworkInterface{{IpAddresses: []net.IP{
				parseCIDR("127.0.0.1/8"),
				parseCIDR("::1/128")}}}},
		{name: `MacAddress`,
			input: baseInput(nil, []interface{}{map[string]interface{}{
				"hardware-address": string("54:1a:12:8f:7b:ed")}}),
			output: []AgentNetworkInterface{{MacAddress: parseMAC("54:1a:12:8f:7b:ed")}}},
		{name: `Name`,
			input: baseInput(nil, []interface{}{map[string]interface{}{
				"name": "test"}}),
			output: []AgentNetworkInterface{{Name: string("test")}}},
		{name: `Statistics false full`,
			input: baseInput(util.Pointer(false), []interface{}{map[string]interface{}{
				"statistics": map[string]interface{}{
					"rx-bytes": float64(1)}}}),
			output: []AgentNetworkInterface{{Statistics: nil}}},
		{name: `Statistics true full`,
			input: baseInput(util.Pointer(true), []interface{}{map[string]interface{}{
				"statistics": map[string]interface{}{
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
		{name: `Statistics true&false empty`,
			input:  baseInput(nil, []interface{}{map[string]interface{}{}}),
			output: []AgentNetworkInterface{{Statistics: nil}}},
		{name: `Full true`,
			input: baseInput(util.Pointer(true), inputFullTest()),
			output: []AgentNetworkInterface{
				{Name: string("lo"),
					MacAddress: parseMAC("54:1a:12:8f:7b:ed"),
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
		{name: `Full false`,
			input: baseInput(util.Pointer(false), inputFullTest()),
			output: []AgentNetworkInterface{
				{Name: string("lo"),
					MacAddress: parseMAC("54:1a:12:8f:7b:ed"),
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
						parseCIDR("10.20.30.244/16")}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.input.statistics != nil {
				require.Equal(t, test.output, AgentNetworkInterface{}.mapToSDK(test.input.params, *test.input.statistics))
			} else {
				require.Equal(t, test.output, AgentNetworkInterface{}.mapToSDK(test.input.params, false))
				require.Equal(t, test.output, AgentNetworkInterface{}.mapToSDK(test.input.params, true))
			}
		})
	}
}
