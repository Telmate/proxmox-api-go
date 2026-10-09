package proxmox

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_Networks_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `all e1000`,
				input: map[string]interface{}{"net0": "e1000=BC:24:11:E1:BB:5d,bridge=vmbr0,mtu=1395,firewall=1,link_down=1,queues=23,rate=1.53,tag=12,trunks=34;18;25"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
					Bridge:    new("vmbr0"),
					Connected: new(false),
					Firewall:  new(true),
					MAC:       new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:       "BC:24:11:E1:BB:5d",
					// MTU is only supported for virtio
					Model:         new(QemuNetworkModelE1000),
					MultiQueue:    new(QemuNetworkQueue(23)),
					RateLimitKBps: new(GuestNetworkRate(1530)),
					NativeVlan:    new(Vlan(12)),
					TaggedVlans:   new(Vlans{34, 18, 25})}}})},
			{name: `all virtio`,
				input: map[string]interface{}{"net31": "virtio=BC:24:11:E1:BB:5D,bridge=vmbr0,mtu=1395,firewall=1,link_down=1,queues=23,rate=1.53,tag=12,trunks=34;18;25"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: QemuNetworkInterface{
					Bridge:        new("vmbr0"),
					Connected:     new(false),
					Firewall:      new(true),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					MTU:           new(QemuMTU{Value: 1395}),
					Model:         new(QemuNetworkModelVirtIO),
					MultiQueue:    new(QemuNetworkQueue(23)),
					RateLimitKBps: new(GuestNetworkRate(1530)),
					NativeVlan:    new(Vlan(12)),
					TaggedVlans:   new(Vlans{34, 18, 25})}}})},
			{name: `Bridge`,
				input: map[string]interface{}{"net2": "virtio=BC:24:11:E1:BB:5D,bridge=vmbr0"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{
					Bridge:      new("vmbr0"),
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Model and Mac`,
				input: map[string]interface{}{"net3": "virtio=BC:24:11:E1:BB:5D"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Connected false`,
				input: map[string]interface{}{"net4": "virtio=BC:24:11:E1:BB:5D,link_down=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{
					Connected:   new(false),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Connected true`,
				input: map[string]interface{}{"net5": "virtio=BC:24:11:E1:BB:5D,link_down=0"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Connected unset`,
				input: map[string]interface{}{"net6": "virtio=BC:24:11:E1:BB:5D"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Firwall true`,
				input: map[string]interface{}{"net7": "virtio=BC:24:11:E1:BB:5D,firewall=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(true),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Firwall false`,
				input: map[string]interface{}{"net8": "virtio=BC:24:11:E1:BB:5D,firewall=0"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `Firwall unset`,
				input: map[string]interface{}{"net9": "virtio=BC:24:11:E1:BB:5D"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `MTU value`,
				input: map[string]interface{}{"net10": "virtio=BC:24:11:E1:BB:5D,mtu=1500"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					MTU:         &QemuMTU{Value: 1500},
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `MTU inherit`,
				input: map[string]interface{}{"net11": "virtio=BC:24:11:E1:BB:5D,mtu=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					MTU:         &QemuMTU{Inherit: true},
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `MultiQueue disable`,
				input: map[string]interface{}{"net12": "virtio=BC:24:11:E1:BB:5D"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `MultiQueue enable`,
				input: map[string]interface{}{"net0": "virtio=BC:24:11:E1:BB:5D,queues=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MultiQueue:  new(QemuNetworkQueue(1)),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `RateLimitKBps disable`,
				input: map[string]interface{}{"net13": "virtio=BC:24:11:E1:BB:5D"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{})}}})},
			{name: `RateLimitKBps 0.001`,
				input: map[string]interface{}{"net14": "virtio=BC:24:11:E1:BB:5D,rate=0.001"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(1)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `RateLimitKBps 0.01`,
				input: map[string]interface{}{"net15": "virtio=BC:24:11:E1:BB:5D,rate=0.010"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(10)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `RateLimitKBps 0.1`,
				input: map[string]interface{}{"net16": "virtio=BC:24:11:E1:BB:5D,rate=0.1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(100)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `RateLimitKBps 1`,
				input: map[string]interface{}{"net17": "virtio=BC:24:11:E1:BB:5D,rate=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(1000)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `RateLimitKBps 1.264`,
				input: map[string]interface{}{"net18": "virtio=BC:24:11:E1:BB:5D,rate=1.264"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(1264)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `RateLimitKBps 15.264`,
				input: map[string]interface{}{"net19": "virtio=BC:24:11:E1:BB:5D,rate=15.264"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{
					Connected:     new(true),
					Firewall:      new(false),
					MAC:           new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:           "BC:24:11:E1:BB:5D",
					Model:         new(QemuNetworkModelVirtIO),
					RateLimitKBps: new(GuestNetworkRate(15264)),
					TaggedVlans:   new(Vlans{})}}})},
			{name: `NaitiveVlan`,
				input: map[string]interface{}{"net20": "virtio=BC:24:11:E1:BB:5D,tag=1"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					NativeVlan:  new(Vlan(1)),
					TaggedVlans: new(Vlans{})}}})},
			{name: `TaggedVlans`,
				input: map[string]interface{}{"net21": "virtio=BC:24:11:E1:BB:5D,trunks=1;63;21"},
				output: testQemuBaseConfig_get(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{
					Connected:   new(true),
					Firewall:    new(false),
					MAC:         new(parseMAC("BC:24:11:E1:BB:5D")),
					mac:         "BC:24:11:E1:BB:5D",
					Model:       new(QemuNetworkModelVirtIO),
					TaggedVlans: new(Vlans{1, 63, 21})}}})},
		}
	}
}

func Test_ConfigQemu_QemuNetworkInterfaces_MapToApi(t *testing.T) {
	t.Parallel()
	networkInterface := func() QemuNetworkInterface {
		return QemuNetworkInterface{
			Bridge:        new("vmbr0"),
			Connected:     new(false),
			Firewall:      new(true),
			MAC:           new(parseMAC("52:54:00:12:34:56")),
			MTU:           new(QemuMTU{Value: 1500}),
			Model:         new(QemuNetworkModel("virtio")),
			MultiQueue:    new(QemuNetworkQueue(5)),
			RateLimitKBps: new(GuestNetworkRate(45)),
			NativeVlan:    new(Vlan(23)),
			TaggedVlans:   new(Vlans{12, 23, 45})}
	}
	tests := qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			create: []qemuTestCaseAPI{
				{name: `MTU.Value model=none`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1400)})}}},
					body:   map[string]string{"net7": ""}}},
			createUpdate: []qemuTestCaseAPI{
				{name: `Delete no effect`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
						Bridge:        new("vmbr0"),
						Connected:     new(true),
						Delete:        true,
						Firewall:      new(true),
						MAC:           new(net.HardwareAddr("00:11:22:33:44:55")),
						Model:         new(QemuNetworkModelVirtIO),
						MultiQueue:    new(QemuNetworkQueue(4)),
						RateLimitKBps: new(GuestNetworkRate(45)),
						NativeVlan:    new(Vlan(23))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{}}}},
				{name: `Bridge`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Bridge: new("vmbr0")}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Bridge: new("vmbr1")}}},
					body:          map[string]string{"net0": "%2Cbridge%3Dvmbr0"}}, // ",bridge=vmbr0"
				{name: `Connected true`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{Connected: new(true)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{Connected: new(false)}}},
					body:          map[string]string{"net1": ""}},
				{name: `Connected false`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{Connected: new(false)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{Connected: new(true)}}},
					body:          map[string]string{"net2": "%2Clink_down%3D1"}}, // ",link_down=1"
				{name: `Firewall true`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{Firewall: new(true)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{Firewall: new(false)}}},
					body:          map[string]string{"net3": "%2Cfirewall%3D1"}}, // ",firewall=1"
				{name: `Firewall false`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{Firewall: new(false)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{Firewall: new(true)}}},
					body:          map[string]string{"net4": ""}},
				{name: `MAC`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
						Model: new(QemuNetworkModelE1000),
						MAC:   new(net.HardwareAddr(parseMAC("BC:11:22:33:44:55")))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
						Model: new(QemuNetworkModelVirtIO),
						MAC:   new(net.HardwareAddr(parseMAC("bc:11:22:33:44:56")))}}},
					body: map[string]string{"net5": "e1000%3DBC%3A11%3A22%3A33%3A44%3A55"}}, // "e1000=BC:11:22:33:44:55"
				{name: `MTU.Inherit model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
						Model: new(QemuNetworkModelVirtIO),
						MTU:   new(QemuMTU{Inherit: true})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1500)})}}},
					body:          map[string]string{"net6": "virtio%2Cmtu%3D1"}}, // "virtio,mtu=1"
				{name: `MTU.Value model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
						Model: new(QemuNetworkModelVirtIO),
						MTU:   new(QemuMTU{Value: MTU(1400)})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1500)})}}},
					body:          map[string]string{"net7": "virtio%2Cmtu%3D1400"}}, // "virtio,mtu=1400"
				{name: `MTU.Value=0 model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
						Model: new(QemuNetworkModelVirtIO),
						MTU:   new(QemuMTU{Value: MTU(0)})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{MTU: new(QemuMTU{})}}},
					body:          map[string]string{"net8": "virtio"}},
				{name: `Model`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{Model: new(qemuNetworkModelE100082544gc_Lower)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{Model: new(QemuNetworkModelVirtIO)}}},
					body:          map[string]string{"net9": "e1000-82544gc"}},
				{name: `Model invalid`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{Model: new(QemuNetworkModel("gibberish"))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{Model: new(QemuNetworkModelVirtIO)}}},
					body:          map[string]string{"net10": ""}},
				{name: `MultiQueue set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(4))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(2))}}},
					body:          map[string]string{"net11": "%2Cqueues%3D4"}}, // ",queues=4"
				{name: `MultiQueue unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(0))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(2))}}},
					body:          map[string]string{"net12": ""}},
				{name: `RateLimitKBps 0`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(0))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net13": ""}},
				{name: `RateLimitKBps 0.007`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(7))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net13": "%2Crate%3D0.007"}}, // ",rate=0.007"
				{name: `RateLimitKBps 0.07`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(70))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net14": "%2Crate%3D0.07"}}, // ",rate=0.07"
				{name: `RateLimitKBps 0.7`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(700))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net15": "%2Crate%3D0.7"}}, // ",rate=0.7"
				{name: `RateLimitKBps 7`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(7000))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net16": "%2Crate%3D7"}}, // ",rate=7"
				{name: `RateLimitKBps 7.546`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(7546))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net17": "%2Crate%3D7.546"}}, // ",rate=7.546"
				{name: `RateLimitKBps 734.546`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(734546))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(5))}}},
					body:          map[string]string{"net18": "%2Crate%3D734.546"}}, // ",rate=734.546"
				{name: `NativeVlan unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{NativeVlan: new(Vlan(0))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{NativeVlan: new(Vlan(2))}}},
					body:          map[string]string{"net19": ""}},
				{name: `NativeVlan set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{NativeVlan: new(Vlan(83))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{NativeVlan: new(Vlan(2))}}},
					body:          map[string]string{"net20": "%2Ctag%3D83"}}, // ",tag=83"
				{name: `TaggedVlans set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{TaggedVlans: new(Vlans{10, 43, 23})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{TaggedVlans: new(Vlans{12, 56})}}},
					body:          map[string]string{"net21": "%2Ctrunks%3D10%3B23%3B43"}}, // ",trunks=10;23;43"
				{name: `TaggedVlans unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: new(Vlans{})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: new(Vlans{12, 56})}}},
					body:          map[string]string{"net22": ""}}},
			update: []qemuTestCaseAPI{
				{name: `create`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: networkInterface()}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{}}},
					body:          map[string]string{"net12": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `Bridge replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: QemuNetworkInterface{Bridge: new("vmbr45")}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: networkInterface()}},
					body:          map[string]string{"net31": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr45%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr45,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `Connected replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID30: QemuNetworkInterface{Connected: new(true)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID30: networkInterface()}},
					body:          map[string]string{"net30": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `Firewall replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID29: QemuNetworkInterface{Firewall: new(false)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID29: networkInterface()}},
					body:          map[string]string{"net29": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `MAC clear`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{MAC: new(net.HardwareAddr(""))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: networkInterface()}},
					body:          map[string]string{"net28": "virtio%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `MAC replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{MAC: new(net.HardwareAddr(parseMAC("BC:24:11:C2:75:20")))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: networkInterface()}},
					body:          map[string]string{"net28": "virtio%3DBC%3A24%3A11%3AC2%3A75%3A20%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=BC:24:11:C2:75:20,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `MAC binary match, do nothing`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{MAC: new(net.HardwareAddr(parseMAC("BC:24:11:C2:75:20")))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{
						MAC: new(net.HardwareAddr(parseMAC("bc:24:11:C2:75:20"))),
						mac: "bc:24:11:C2:75:20"}}}},
				{name: `MAC no update mixed case, do nothing`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{
						MAC: new(net.HardwareAddr(parseMAC("bc:24:11:C2:75:20"))),
						mac: "bc:24:11:C2:75:20"}}}},
				{name: `MTU.Value model=none`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1400)})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1500)})}}}},
				{name: `MTU replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID27: QemuNetworkInterface{MTU: new(QemuMTU{Value: MTU(1400)})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID27: networkInterface()}},
					body:          map[string]string{"net27": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1400%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1400,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `Model replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID26: QemuNetworkInterface{Model: new(qemuNetworkModelE100082544gc_Lower)}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID26: networkInterface()}},
					body:          map[string]string{"net26": "e1000-82544gc%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "e1000-82544gc=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,queues=5,rate=0.045,tag=23,trunks=12;23;45"
				{name: `MultiQueue replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID25: QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(4))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID25: networkInterface()}},
					body:          map[string]string{"net25": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D4%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=4,rate=0.045,tag=23,trunks=12;23;45"
				{name: `RateLimitKBps replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID24: QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(539))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID24: networkInterface()}},
					body:          map[string]string{"net24": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.539%2Ctag%3D23%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.539,tag=23,trunks=12;23;45"
				{name: `NaitiveVlan replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID23: QemuNetworkInterface{NativeVlan: new(Vlan(0))}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID23: networkInterface()}},
					body:          map[string]string{"net23": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctrunks%3D12%3B23%3B45"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,trunks=12;23;45"
				{name: `TaggedVlans replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: new(Vlans{10, 70, 18})}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: networkInterface()}},
					body:          map[string]string{"net22": "virtio%3D52%3A54%3A00%3A12%3A34%3A56%2Cbridge%3Dvmbr0%2Cfirewall%3D1%2Clink_down%3D1%2Cmtu%3D1500%2Cqueues%3D5%2Crate%3D0.045%2Ctag%3D23%2Ctrunks%3D10%3B18%3B70"}}, // "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=10;18;70"
				{name: `Delete`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{Delete: true}}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{}}},
					body:          map[string]string{"delete": "net21"}},
				{name: `no change`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: networkInterface()}},
					currentLegacy: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: networkInterface()}}},
			}}
	})
	tests.Test(t)
}

func Test_RawConfigQemu_GetNetworks(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Networks_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, config *ConfigQemu, err error) {
			require.Equal(t, config.Networks, raw.GetNetworks())
		})
}

func Test_RawConfigQemu_Networks_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Networks_Get().Test(t)
}

func Test_QemuMTU_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuMTU
		output error
	}{
		{name: `Valid inherit`,
			input: QemuMTU{Inherit: true}},
		{name: `Valid value`,
			input: QemuMTU{Value: 1500}},
		{name: `Valid empty`},
		{name: `Invalid mutually exclusive`,
			input:  QemuMTU{Inherit: true, Value: 1500},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid too small`,
			input:  QemuMTU{Value: 575},
			output: errors.New(MTU_Error_Invalid)},
		{name: `Invalid too large`,
			input:  QemuMTU{Value: 65521},
			output: errors.New(MTU_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkInterface_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		config  QemuNetworkInterface
		current *QemuNetworkInterface
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{
				config: QemuNetworkInterface{Delete: true}}},
		{name: `Valid MTU inherit`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MTU value`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 1500}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MTU empty`,
			input: testInput{
				config:  QemuNetworkInterface{MTU: &QemuMTU{}},
				current: &QemuNetworkInterface{}}},
		{name: `Valid Model`,
			input: testInput{
				config:  QemuNetworkInterface{Model: new(QemuNetworkModel("virtio"))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid MultiQueue`,
			input: testInput{
				config:  QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(64))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid RateLimitKBps`,
			input: testInput{
				config:  QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(10240000))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid NativeVlan`,
			input: testInput{
				config:  QemuNetworkInterface{NativeVlan: new(Vlan(5))},
				current: &QemuNetworkInterface{}}},
		{name: `Valid TaggedVlans`,
			input: testInput{
				config:  QemuNetworkInterface{TaggedVlans: new(Vlans{0, 45, 12, 4095, 12, 45})},
				current: &QemuNetworkInterface{}},
		},
		// Invalid
		{name: `Invalid errors.New(QemuNetworkInterface_Error_BridgeRequired)`,
			input:  testInput{config: QemuNetworkInterface{}},
			output: errors.New(QemuNetworkInterface_Error_BridgeRequired)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_ModelRequired)`,
			input:  testInput{config: QemuNetworkInterface{Bridge: new("vmbr0")}},
			output: errors.New(QemuNetworkInterface_Error_ModelRequired)},
		{name: `Invalid errors.New(QemuMTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true, Value: 1500}},
				current: &QemuNetworkInterface{}},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid errors.New(MTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 575}},
				current: &QemuNetworkInterface{}},
			output: errors.New(MTU_Error_Invalid)},

		{name: `Invalid Model`,
			input: testInput{
				config:  QemuNetworkInterface{Model: new(QemuNetworkModel("invalid"))},
				current: &QemuNetworkInterface{}},
			output: QemuNetworkModel("").Error()},
		{name: `Invalid errors.New(QemuNetworkQueue_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{MultiQueue: new(QemuNetworkQueue(65))},
				current: &QemuNetworkInterface{}},
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
		{name: `Invalid errors.New(GuestNetworkRate_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{RateLimitKBps: new(GuestNetworkRate(10240001))},
				current: &QemuNetworkInterface{}},
			output: errors.New(GuestNetworkRate_Error_Invalid)},
		{name: `Invalid NativeVlan errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{NativeVlan: new(Vlan(4096))},
				current: &QemuNetworkInterface{}},
			output: errors.New(Vlan_Error_Invalid)},
		{name: `Invalid TaggedVlans errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config:  QemuNetworkInterface{TaggedVlans: new(Vlans{0, 45, 12, 4095, 12, 45, 4096})},
				current: &QemuNetworkInterface{}},
			output: errors.New(Vlan_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuNetworkInterfaceID_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuNetworkInterfaceID
		output error
	}{
		{name: "Valid",
			input: QemuNetworkInterfaceID0},
		{name: "Invalid",
			input:  32,
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkInterfaces_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		config  QemuNetworkInterfaces
		current QemuNetworkInterfaces
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		{name: `Valid Delete`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Delete: true}}}},
		{name: `Valid MTU inherit`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{}}}},
		{name: `Valid MTU value`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 1500}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{}}}},
		{name: `Valid MTU empty`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{
					MTU: &QemuMTU{}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{}}}},
		{name: `Valid Model`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{
					Model: new(QemuNetworkModel("virtio"))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{}}}},
		{name: `Valid MultiQueue`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{
					MultiQueue: new(QemuNetworkQueue(64))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{}}}},
		{name: `Valid RateLimitKBps`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
					RateLimitKBps: new(GuestNetworkRate(10240000))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{}}}},
		{name: `Valid NativeVlan`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
					NativeVlan: new(Vlan(5))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{}}}},
		{name: `Valid TaggedVlans`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
					TaggedVlans: new(Vlans{0, 45, 12, 4095, 12, 45})}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{}}}},
		// Invalid
		{name: `Invalid errors.New(QemuNetworkInterfaceID_Error_Invalid)`,
			input:  testInput{config: QemuNetworkInterfaces{32: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkInterfaceID_Error_Invalid)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_BridgeRequired)`,
			input:  testInput{config: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkInterface_Error_BridgeRequired)},
		{name: `Invalid errors.New(QemuNetworkInterface_Error_ModelRequired)`,
			input: testInput{config: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
				Bridge: new("vmbr0")}}},
			output: errors.New(QemuNetworkInterface_Error_ModelRequired)},
		{name: `Invalid errors.New(MTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Value: 575}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{}}},
			output: errors.New(MTU_Error_Invalid)},
		{name: `Invalid errors.New(QemuMTU_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{
					Model: new(QemuNetworkModelVirtIO),
					MTU:   &QemuMTU{Inherit: true, Value: 1500}}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{}}},
			output: errors.New(QemuMTU_Error_Invalid)},
		{name: `Invalid Model`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{
					Model: new(QemuNetworkModel("invalid"))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{}}},
			output: QemuNetworkModel("").Error()},
		{name: `Invalid errors.New(QemuNetworkQueue_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{
					MultiQueue: new(QemuNetworkQueueMaximum + 1)}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{}}},
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
		{name: `Invalid errors.New(QemuNetworkRate_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{
					RateLimitKBps: new(GuestNetworkRate(10240001))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{}}},
			output: errors.New(GuestNetworkRate_Error_Invalid)},
		{name: `Invalid NativeVlan errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{
					NativeVlan: new(Vlan(4096))}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{}}},
			output: errors.New(Vlan_Error_Invalid)},
		{name: `Invalid TaggedVlans errors.New(Vlan_Error_Invalid)`,
			input: testInput{
				config: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{
					TaggedVlans: new(Vlans{0, 45, 12, 4095, 12, 45, 4096})}},
				current: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{}}},
			output: errors.New(Vlan_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.Validate(test.input.current))
		})
	}
}

func Test_QemuNetworkModel_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuNetworkModel
		output error
	}{
		{name: `Valid weird`,
			input: "E__1--0__-__00-8__2--__--545_Em__"},
		{name: `Valid normal`,
			input: QemuNetworkModelE100082544gc},
		{name: `Invalid`,
			input:  "invalid",
			output: QemuNetworkModel("").Error()},
		{name: `Invalid empty`,
			input:  "",
			output: QemuNetworkModel("").Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuNetworkQueue_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuNetworkQueue
		output error
	}{
		{name: `Valid Minimum`,
			input: 0},
		{name: `Valid Maximum`,
			input: 64},
		{name: `Invalid`,
			input:  65,
			output: errors.New(QemuNetworkQueue_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}
