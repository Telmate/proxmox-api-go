package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_HotPlug_Get() qemuTestGetFunc {
	base := func(h HotPlug) *HotPlug {
		if h.CPU == nil {
			h.CPU = new(false)
		}
		if h.Disk == nil {
			h.Disk = new(false)
		}
		if h.Memory == nil {
			h.Memory = new(false)
		}
		if h.Network == nil {
			h.Network = new(false)
		}
		if h.USB == nil {
			h.USB = new(false)
		}
		return &h
	}
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `all nil`,
				input:  map[string]any{},
				output: testQemuBaseConfig_get(ConfigQemu{})},
			{name: `all false`,
				input:  map[string]any{"hotplug": "0"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{})})},
			{name: `CPU`,
				input:  map[string]any{"hotplug": "cpu"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{CPU: new(true)})})},
			{name: `Disk`,
				input:  map[string]any{"hotplug": "disk"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{Disk: new(true)})})},
			{name: `Memory`,
				input:  map[string]any{"hotplug": "memory"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{Memory: new(true)})})},
			{name: `Network`,
				input:  map[string]any{"hotplug": "network"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{Network: new(true)})})},
			{name: `USB`,
				input:  map[string]any{"hotplug": "usb"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: base(HotPlug{USB: new(true)})})},
			{name: `all`,
				input: map[string]any{"hotplug": "disk,cpu,memory,usb,network"},
				output: testQemuBaseConfig_get(ConfigQemu{HotPlug: &HotPlug{
					CPU:     new(true),
					Disk:    new(true),
					Memory:  new(true),
					Network: new(true),
					USB:     new(true)}})}}
	}
}

func Test_ConfigQemu_HotPlug_MapToApi(t *testing.T) {
	t.Parallel()
	base := func(v bool) *HotPlug {
		return &HotPlug{
			CPU:     &v,
			Disk:    &v,
			Memory:  &v,
			Network: &v,
			USB:     &v}
	}

	tests := qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			create: []qemuTestCaseAPI{
				{name: `default`,
					config: &ConfigQemu{HotPlug: &HotPlug{
						CPU:     new(false),
						Disk:    new(true),
						Memory:  new(false),
						Network: new(true),
						USB:     new(true)}}},
				{name: `CPU false`,
					config: &ConfigQemu{HotPlug: &HotPlug{CPU: new(false)}},
					body:   map[string]string{"hotplug": "0"}},
				{name: `Disk false`,
					config: &ConfigQemu{HotPlug: &HotPlug{Disk: new(false)}},
					body:   map[string]string{"hotplug": "0"}},
				{name: `Memory false`,
					config: &ConfigQemu{HotPlug: &HotPlug{Memory: new(false)}},
					body:   map[string]string{"hotplug": "0"}},
				{name: `Network false`,
					config: &ConfigQemu{HotPlug: &HotPlug{Network: new(false)}},
					body:   map[string]string{"hotplug": "0"}},
				{name: `USB false`,
					config: &ConfigQemu{HotPlug: &HotPlug{USB: new(false)}},
					body:   map[string]string{"hotplug": "0"}}},
			createUpdate: []qemuTestCaseAPI{
				{name: `CPU true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{CPU: new(true)}},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "cpu"}},
				{name: `Disk true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Disk: new(true)}},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "disk"}},
				{name: `Memory true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Memory: new(true)}},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "memory"}},
				{name: `Network true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Network: new(true)}},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "network"}},
				{name: `USB true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{USB: new(true)}},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "usb"}},
				{name: `all false to true`,
					config:        &ConfigQemu{HotPlug: base(true)},
					currentLegacy: ConfigQemu{HotPlug: base(false)},
					body:          map[string]string{"hotplug": "cpu%2Cdisk%2Cmemory%2Cnetwork%2Cusb"}}}, // "cpu,disk,memory,network,usb"
			update: []qemuTestCaseAPI{
				{name: `CPU false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{CPU: new(false)}},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "disk%2Cmemory%2Cnetwork%2Cusb"}}, // "disk,memory,network,usb"
				{name: `Disk false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Disk: new(false)}},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "cpu%2Cmemory%2Cnetwork%2Cusb"}}, // "cpu,memory,network,usb"
				{name: `Memory false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Memory: new(false)}},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "cpu%2Cdisk%2Cnetwork%2Cusb"}}, // "cpu,disk,network,usb"
				{name: `Network false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{Network: new(false)}},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "cpu%2Cdisk%2Cmemory%2Cusb"}}, // "cpu,disk,memory,usb"
				{name: `USB false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{USB: new(false)}},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "cpu%2Cdisk%2Cmemory%2Cnetwork"}}, // "cpu,disk,memory,network"
				{name: `none true`,
					config:        &ConfigQemu{HotPlug: &HotPlug{}},
					currentLegacy: ConfigQemu{HotPlug: base(true)}},
				{name: `none false`,
					config:        &ConfigQemu{HotPlug: &HotPlug{}},
					currentLegacy: ConfigQemu{HotPlug: base(false)}},
				{name: `all false to false`,
					config:        &ConfigQemu{HotPlug: base(false)},
					currentLegacy: ConfigQemu{HotPlug: base(false)}},
				{name: `all true to false`,
					config:        &ConfigQemu{HotPlug: base(false)},
					currentLegacy: ConfigQemu{HotPlug: base(true)},
					body:          map[string]string{"hotplug": "0"}},
				{name: `all true to true`,
					config:        &ConfigQemu{HotPlug: base(true)},
					currentLegacy: ConfigQemu{HotPlug: base(true)}},
			}}
	})
	tests.Test(t)
}

func Test_RawConfigQemu_GetHotPlug(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_HotPlug_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, *output.HotPlug, raw.GetHotPlug())
		})
}

func Test_RawConfigQemu_HotPlug_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_HotPlug_Get().Test(t)
}
