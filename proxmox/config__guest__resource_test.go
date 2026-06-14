package proxmox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/stretchr/testify/require"
)

func Test_guestClient_List(t *testing.T) {
	t.Parallel()
	const path = "/cluster/resources?type=vm"
	tests := []struct {
		name     string
		output   map[GuestID]GuestResource
		requests []mockServer.Request
		err      error
	}{
		{name: `List`,
			output: map[GuestID]GuestResource{
				100: {
					CpuCores:           10,
					CpuUsage:           3.141592653589793,
					DiskReadTotal:      1637428,
					DiskSizeInBytes:    8589934592,
					DiskUsedInBytes:    0,
					DiskWriteTotal:     1690811,
					HaState:            new("started"),
					ID:                 100,
					MemoryTotalInBytes: 2147483648,
					MemoryUsedInBytes:  1048576,
					Name:               "test-vm1",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "pve1",
					Status:             PowerStateRunning,
					Tags:               []Tag{"tag1", "tag2", "tag3"},
					Template:           false,
					Type:               GuestQemu,
					Uptime:             72169 * time.Second},
				999: {
					CpuCores:           1,
					CpuUsage:           0,
					DiskReadTotal:      846348234,
					DiskUsedInBytes:    0,
					DiskSizeInBytes:    56742482484,
					DiskWriteTotal:     3432,
					ID:                 999,
					MemoryTotalInBytes: 727345728374,
					MemoryUsedInBytes:  68467234324,
					Name:               "template-linux",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "node3",
					Status:             PowerStateStopped,
					Tags:               []Tag{"template"},
					Template:           true,
					Type:               GuestQemu,
					Uptime:             0},
				100000: {
					CpuCores:           50,
					CpuUsage:           0.141592653589793,
					DiskReadTotal:      857324,
					DiskUsedInBytes:    23234,
					DiskSizeInBytes:    9743424,
					DiskWriteTotal:     78347843754,
					ID:                 100000,
					Locked:             true,
					MemoryTotalInBytes: 946856732535,
					MemoryUsedInBytes:  1342,
					Name:               "dev-vm1",
					NetworkIn:          2331323424,
					NetworkOut:         88775378423476,
					Node:               "pve2",
					Status:             PowerStateRunning,
					Tags:               []Tag{"dev"},
					Template:           false,
					Type:               GuestLxc,
					Uptime:             88678345 * time.Second}},
			requests: mockServer.RequestsGetJson(path, map[string]any{
				"data": []map[string]any{
					{
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
						"status":    "running",
						"tags":      "tag1;tag2;tag3",
						"template":  float64(0),
						"type":      "qemu",
						"uptime":    float64(72169)},
					{
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
						"lock":      string("clone"),
						"netin":     float64(2331323424),
						"netout":    float64(88775378423476),
						"node":      "pve2",
						"status":    "running",
						"tags":      "dev",
						"template":  float64(0),
						"type":      "lxc",
						"uptime":    float64(88678345)},
					{
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
						"status":    "stopped",
						"tags":      "template",
						"template":  float64(1),
						"type":      "qemu",
						"uptime":    float64(0),
					}}})},
		{name: `500 internal server error`,
			requests: mockServer.RequestsError(path, mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Guest.List(context.Background())
			require.Equal(t, test.err, err)
			if test.err == nil {
				guestMap := raw.AsMap()
				require.Len(t, guestMap, len(test.output))
				for k := range guestMap {
					v, ok := guestMap[k]
					if !ok {
						t.Fatalf("expected (%v) not found", k)
					}
					guest := v.Get()
					require.Equal(t, test.output[k], guest)
				}
			}
			server.Clear(t)
		})
	}
}

func Test_RawGuestResources_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGuestResources
		output []GuestResource
	}{
		{name: "CpuCores",
			input:  rawGuestResources{a: []any{map[string]any{"maxcpu": float64(10)}}},
			output: []GuestResource{{CpuCores: 10}}},
		{name: "CpuUsage",
			input:  rawGuestResources{a: []any{map[string]any{"cpu": float64(3.141592653589793)}}},
			output: []GuestResource{{CpuUsage: 3.141592653589793}}},
		{name: "DiskReadTotal",
			input:  rawGuestResources{a: []any{map[string]any{"diskread": float64(1637428)}}},
			output: []GuestResource{{DiskReadTotal: 1637428}}},
		{name: "DiskSizeInBytes",
			input:  rawGuestResources{a: []any{map[string]any{"maxdisk": float64(8589934592)}}},
			output: []GuestResource{{DiskSizeInBytes: 8589934592}}},
		{name: "DiskUsedInBytes",
			input:  rawGuestResources{a: []any{map[string]any{"disk": float64(1073741824)}}},
			output: []GuestResource{{DiskUsedInBytes: 1073741824}}},
		{name: "DiskWriteTotal",
			input:  rawGuestResources{a: []any{map[string]any{"diskwrite": float64(1690811)}}},
			output: []GuestResource{{DiskWriteTotal: 1690811}}},
		{name: "HaState",
			input:  rawGuestResources{a: []any{map[string]any{"hastate": "started"}}},
			output: []GuestResource{{HaState: new("started")}}},
		{name: "Id",
			input:  rawGuestResources{a: []any{map[string]any{"vmid": float64(100)}}},
			output: []GuestResource{{ID: 100}}},
		{name: "Locked",
			input:  rawGuestResources{a: []any{map[string]any{"lock": "clone"}}},
			output: []GuestResource{{Locked: true}}},
		{name: "MemoryTotalInBytes",
			input:  rawGuestResources{a: []any{map[string]any{"maxmem": float64(2147483648)}}},
			output: []GuestResource{{MemoryTotalInBytes: 2147483648}}},
		{name: "MemoryUsedInBytes",
			input:  rawGuestResources{a: []any{map[string]any{"mem": float64(1048576)}}},
			output: []GuestResource{{MemoryUsedInBytes: 1048576}}},
		{name: "Name",
			input:  rawGuestResources{a: []any{map[string]any{"name": "test-vm1"}}},
			output: []GuestResource{{Name: "test-vm1"}}},
		{name: "NetworkIn",
			input:  rawGuestResources{a: []any{map[string]any{"netin": float64(23884639)}}},
			output: []GuestResource{{NetworkIn: 23884639}}},
		{name: "NetworkOut",
			input:  rawGuestResources{a: []any{map[string]any{"netout": float64(1000123465987)}}},
			output: []GuestResource{{NetworkOut: 1000123465987}}},
		{name: "Node",
			input:  rawGuestResources{a: []any{map[string]any{"node": "pve1"}}},
			output: []GuestResource{{Node: "pve1"}}},
		{name: "Status",
			input:  rawGuestResources{a: []any{map[string]any{"status": "running"}}},
			output: []GuestResource{{Status: PowerStateRunning}}},
		{name: "Tags",
			input:  rawGuestResources{a: []any{map[string]any{"tags": "tag1;tag2;tag3"}}},
			output: []GuestResource{{Tags: []Tag{"tag1", "tag2", "tag3"}}}},
		{name: "Template",
			input:  rawGuestResources{a: []any{map[string]any{"template": float64(1)}}},
			output: []GuestResource{{Template: true}}},
		{name: "Type",
			input:  rawGuestResources{a: []any{map[string]any{"type": "qemu"}}},
			output: []GuestResource{{Type: GuestQemu}}},
		{name: "Uptime",
			input:  rawGuestResources{a: []any{map[string]any{"uptime": float64(72169)}}},
			output: []GuestResource{{Uptime: 72169 * time.Second}}},
		{name: "[]GuestResource",
			input: rawGuestResources{a: []any{
				map[string]any{
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
					"status":    "running",
					"tags":      "tag1;tag2;tag3",
					"template":  float64(0),
					"type":      "qemu",
					"uptime":    float64(72169)},
				map[string]any{
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
					"lock":      string("clone"),
					"netin":     float64(2331323424),
					"netout":    float64(88775378423476),
					"node":      "pve2",
					"status":    "running",
					"tags":      "dev",
					"template":  float64(0),
					"type":      "lxc",
					"uptime":    float64(88678345)},
				map[string]any{
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
					"status":    "stopped",
					"tags":      "template",
					"template":  float64(1),
					"type":      "qemu",
					"uptime":    float64(0)}}},
			output: []GuestResource{
				{
					CpuCores:           10,
					CpuUsage:           3.141592653589793,
					DiskReadTotal:      1637428,
					DiskUsedInBytes:    0,
					DiskSizeInBytes:    8589934592,
					DiskWriteTotal:     1690811,
					HaState:            new("started"),
					ID:                 100,
					MemoryTotalInBytes: 2147483648,
					MemoryUsedInBytes:  1048576,
					Name:               "test-vm1",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "pve1",
					Status:             PowerStateRunning,
					Tags:               []Tag{"tag1", "tag2", "tag3"},
					Template:           false,
					Type:               GuestQemu,
					Uptime:             72169 * time.Second,
				},
				{
					CpuCores:           50,
					CpuUsage:           0.141592653589793,
					DiskReadTotal:      857324,
					DiskUsedInBytes:    23234,
					DiskSizeInBytes:    9743424,
					DiskWriteTotal:     78347843754,
					ID:                 100000,
					Locked:             true,
					MemoryTotalInBytes: 946856732535,
					MemoryUsedInBytes:  1342,
					Name:               "dev-vm1",
					NetworkIn:          2331323424,
					NetworkOut:         88775378423476,
					Node:               "pve2",
					Status:             PowerStateRunning,
					Tags:               []Tag{"dev"},
					Template:           false,
					Type:               GuestLxc,
					Uptime:             88678345 * time.Second,
				},
				{
					CpuCores:           1,
					CpuUsage:           0,
					DiskReadTotal:      846348234,
					DiskUsedInBytes:    0,
					DiskSizeInBytes:    56742482484,
					DiskWriteTotal:     3432,
					ID:                 999,
					MemoryTotalInBytes: 727345728374,
					MemoryUsedInBytes:  68467234324,
					Name:               "template-linux",
					NetworkIn:          23884639,
					NetworkOut:         1000123465987,
					Node:               "node3",
					Status:             PowerStateStopped,
					Tags:               []Tag{"template"},
					Template:           true,
					Type:               GuestQemu,
					Uptime:             0,
				}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			rawGuests := test.input.AsArray()
			guests := make([]GuestResource, len(rawGuests))
			for i := range rawGuests {
				guests[i] = rawGuests[i].Get()
			}
			require.Equal(t, test.output, guests, test.name)
		})
	}
}

func Test_RawGuestResources_Iter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGuestResources
		output []RawGuestResource
	}{
		{name: `Empty`,
			input:  rawGuestResources{a: []any{}},
			output: []RawGuestResource{}},
		{name: `Single Guest`,
			input: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)}}},
			output: []RawGuestResource{
				&rawGuestResource{a: map[string]any{"vmid": float64(100)}}}},
		{name: `Multiple Guests`,
			input: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)},
				map[string]any{"vmid": float64(200)},
				map[string]any{"vmid": float64(300)}}},
			output: []RawGuestResource{
				&rawGuestResource{a: map[string]any{"vmid": float64(100)}},
				&rawGuestResource{a: map[string]any{"vmid": float64(200)}},
				&rawGuestResource{a: map[string]any{"vmid": float64(300)}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test iterating over all items
			var result []RawGuestResource
			for guest := range RawGuestResources(&test.input).Iter() {
				result = append(result, guest)
			}
			require.Equal(t, len(test.output), len(result))
			for i := range result {
				guest := result[i].Get()
				expected := test.output[i].Get()
				require.Equal(t, expected, guest)
			}
			// Test early termination (break after first item)
			if len(test.output) > 0 {
				count := 0
				for range RawGuestResources(&test.input).Iter() {
					count++
					break
				}
				require.Equal(t, 1, count)
			}
		})
	}
}

func Test_RawGuestResources_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawGuestResources
		output int
	}{
		{name: `Empty`,
			input:  rawGuestResources{a: []any{}},
			output: 0},
		{name: `Single Pool`,
			input: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)}}},
			output: 1},
		{name: `Multiple Pools`,
			input: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)},
				map[string]any{"vmid": float64(200)},
				map[string]any{"vmid": float64(300)}}},
			output: 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawGuestResources(&test.input).Len())
		})
	}
}

func Test_RawGuestResources_selectID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  GuestID
		data   rawGuestResources
		isSet  bool
		output *rawGuestResource
	}{
		{name: `Exists`,
			input: 200,
			isSet: true,
			data: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)},
				map[string]any{"vmid": float64(200)},
				map[string]any{"vmid": float64(300)}}},
			output: &rawGuestResource{a: map[string]any{"vmid": float64(200)}}},
		{name: `Doesn't Exists`,
			input: 500,
			data: rawGuestResources{a: []any{
				map[string]any{"vmid": float64(100)},
				map[string]any{"vmid": float64(200)},
				map[string]any{"vmid": float64(300)}}}},
		{name: `No guests`,
			input: 500,
			data:  rawGuestResources{a: []any{}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, isSet := test.data.selectID(test.input)
			require.Equal(t, test.isSet, isSet)
			require.Equal(t, test.output, raw)
		})
	}
}
