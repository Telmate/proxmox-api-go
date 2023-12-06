package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GuestResource_mapToStruct(t *testing.T) {
	tests := []struct {
		name   string
		input  []interface{}
		output []GuestResource
	}{
		{name: "CpuCores",
			input:  []interface{}{map[string]interface{}{"maxcpu": float64(10)}},
			output: []GuestResource{{CpuCores: 10}},
		},
		{name: "CpuUsage",
			input:  []interface{}{map[string]interface{}{"cpu": float64(3.141592653589793)}},
			output: []GuestResource{{CpuUsage: 3.141592653589793}},
		},
		{name: "DiskReadTotal",
			input:  []interface{}{map[string]interface{}{"diskread": float64(1637428)}},
			output: []GuestResource{{DiskReadTotal: 1637428}},
		},
		{name: "DiskSizeInBytes",
			input:  []interface{}{map[string]interface{}{"maxdisk": float64(8589934592)}},
			output: []GuestResource{{DiskSizeInBytes: 8589934592}},
		},
		{name: "DiskUsedInBytes",
			input:  []interface{}{map[string]interface{}{"disk": float64(1073741824)}},
			output: []GuestResource{{DiskUsedInBytes: 1073741824}},
		},
		{name: "DiskWriteTotal",
			input:  []interface{}{map[string]interface{}{"diskwrite": float64(1690811)}},
			output: []GuestResource{{DiskWriteTotal: 1690811}},
		},
		{name: "HaState",
			input:  []interface{}{map[string]interface{}{"hastate": "started"}},
			output: []GuestResource{{HaState: "started"}},
		},
		{name: "Id",
			input:  []interface{}{map[string]interface{}{"vmid": float64(100)}},
			output: []GuestResource{{Id: 100}},
		},
		{name: "MemoryTotalInBytes",
			input:  []interface{}{map[string]interface{}{"maxmem": float64(2147483648)}},
			output: []GuestResource{{MemoryTotalInBytes: 2147483648}},
		},
		{name: "MemoryUsedInBytes",
			input:  []interface{}{map[string]interface{}{"mem": float64(1048576)}},
			output: []GuestResource{{MemoryUsedInBytes: 1048576}},
		},
		{name: "Name",
			input:  []interface{}{map[string]interface{}{"name": "test-vm1"}},
			output: []GuestResource{{Name: "test-vm1"}},
		},
		{name: "NetworkIn",
			input:  []interface{}{map[string]interface{}{"netin": float64(23884639)}},
			output: []GuestResource{{NetworkIn: 23884639}},
		},
		{name: "NetworkOut",
			input:  []interface{}{map[string]interface{}{"netout": float64(1000123465987)}},
			output: []GuestResource{{NetworkOut: 1000123465987}},
		},
		{name: "Node",
			input:  []interface{}{map[string]interface{}{"node": "pve1"}},
			output: []GuestResource{{Node: "pve1"}},
		},
		{name: "Pool",
			input:  []interface{}{map[string]interface{}{"pool": "Production"}},
			output: []GuestResource{{Pool: "Production"}},
		},
		{name: "Status",
			input:  []interface{}{map[string]interface{}{"status": "running"}},
			output: []GuestResource{{Status: "running"}},
		},
		{name: "Tags",
			input:  []interface{}{map[string]interface{}{"tags": "tag1;tag2;tag3"}},
			output: []GuestResource{{Tags: []string{"tag1", "tag2", "tag3"}}},
		},
		{name: "Template",
			input:  []interface{}{map[string]interface{}{"template": float64(1)}},
			output: []GuestResource{{Template: true}},
		},
		{name: "Type",
			input:  []interface{}{map[string]interface{}{"type": "qemu"}},
			output: []GuestResource{{Type: GuestQemu}},
		},
		{name: "UptimeInSeconds",
			input:  []interface{}{map[string]interface{}{"uptime": float64(72169)}},
			output: []GuestResource{{UptimeInSeconds: 72169}},
		},
		{name: "[]GuestResource",
			input: []interface{}{
				map[string]interface{}{
					"maxcpu":    float64(10),
					"cpu":       float64(3.141592653589793),
					"diskread":  float64(1637428),
					"maxdisk":   float64(8589934592),
					"disk":      float64(0),
					"diskwrite": float64(1690811),
					"hastate":   "started",
					"vmid":      float64(100),
					"maxmem":    float64(2147483648),
					"mem":       float64(1048576),
					"name":      "test-vm1",
					"netin":     float64(23884639),
					"netout":    float64(1000123465987),
					"node":      "pve1",
					"pool":      "Production",
					"status":    "running",
					"tags":      "tag1;tag2;tag3",
					"template":  float64(0),
					"type":      "qemu",
					"uptime":    float64(72169),
				},
				map[string]interface{}{
					"maxcpu":    float64(50),
					"cpu":       float64(0.141592653589793),
					"diskread":  float64(857324),
					"maxdisk":   float64(9743424),
					"disk":      float64(23234),
					"diskwrite": float64(78347843754),
					"hastate":   "",
					"vmid":      float64(100000),
					"maxmem":    float64(946856732535),
					"mem":       float64(1342),
					"name":      "dev-vm1",
					"netin":     float64(2331323424),
					"netout":    float64(88775378423476),
					"node":      "pve2",
					"pool":      "Development",
					"status":    "running",
					"tags":      "dev",
					"template":  float64(0),
					"type":      "lxc",
					"uptime":    float64(88678345),
				},
				map[string]interface{}{
					"maxcpu":    float64(1),
					"cpu":       float64(0),
					"diskread":  float64(846348234),
					"disk":      float64(0),
					"maxdisk":   float64(56742482484),
					"diskwrite": float64(3432),
					"hastate":   "",
					"vmid":      float64(999),
					"maxmem":    float64(727345728374),
					"mem":       float64(68467234324),
					"name":      "template-linux",
					"netin":     float64(23884639),
					"netout":    float64(1000123465987),
					"node":      "node3",
					"pool":      "Templates",
					"status":    "stopped",
					"tags":      "template",
					"template":  float64(1),
					"type":      "qemu",
					"uptime":    float64(0),
				},
			},
			output: []GuestResource{
				{
					CpuCores:           10,
					CpuUsage:           3.141592653589793,
					DiskReadTotal:      1637428,
					DiskUsedInBytes:    0,
					DiskSizeInBytes:    8589934592,
					DiskWriteTotal:     1690811,
					HaState:            "started",
					Id:                 100,
					MemoryTotalInBytes: 2147483648,
					MemoryUsedInBytes:  1048576,
					Name:               "test-vm1",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "pve1",
					Pool:               "Production",
					Status:             "running",
					Tags:               []string{"tag1", "tag2", "tag3"},
					Template:           false,
					Type:               GuestQemu,
					UptimeInSeconds:    72169,
				},
				{
					CpuCores:           50,
					CpuUsage:           0.141592653589793,
					DiskReadTotal:      857324,
					DiskUsedInBytes:    23234,
					DiskSizeInBytes:    9743424,
					DiskWriteTotal:     78347843754,
					HaState:            "",
					Id:                 100000,
					MemoryTotalInBytes: 946856732535,
					MemoryUsedInBytes:  1342,
					Name:               "dev-vm1",
					NetworkIn:          2331323424,
					NetworkOut:         88775378423476,
					Node:               "pve2",
					Pool:               "Development",
					Status:             "running",
					Tags:               []string{"dev"},
					Template:           false,
					Type:               GuestLXC,
					UptimeInSeconds:    88678345,
				},
				{
					CpuCores:           1,
					CpuUsage:           0,
					DiskReadTotal:      846348234,
					DiskUsedInBytes:    0,
					DiskSizeInBytes:    56742482484,
					DiskWriteTotal:     3432,
					HaState:            "",
					Id:                 999,
					MemoryTotalInBytes: 727345728374,
					MemoryUsedInBytes:  68467234324,
					Name:               "template-linux",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "node3",
					Pool:               "Templates",
					Status:             "stopped",
					Tags:               []string{"template"},
					Template:           true,
					Type:               GuestQemu,
					UptimeInSeconds:    0,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, GuestResource{}.mapToStruct(test.input), test.name)
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
