package proxmox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_RawNodeInfo_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawNodeInfo
		output NodeInfo
	}{
		{name: `Empty Node Info`,
			input: rawNodeInfo{a: map[string]any{}}},
		{name: `CertificateFingerprint`,
			input:  rawNodeInfo{a: map[string]any{"ssl_fingerprint": string("D0:0C:27:1A:0E:F3:14:16:6F:57:55:CD:4B:EB:7A:B4:C4:17:62:25:B4:C2:E9:5F:0D:9A:F8:E5:EB:EC:D9:13")}},
			output: NodeInfo{CertificateFingerprint: "D0:0C:27:1A:0E:F3:14:16:6F:57:55:CD:4B:EB:7A:B4:C4:17:62:25:B4:C2:E9:5F:0D:9A:F8:E5:EB:EC:D9:13"}},
		{name: `Cpu Threads`,
			input:  rawNodeInfo{a: map[string]any{"maxcpu": float64(12)}},
			output: NodeInfo{Cpu: NodeCpuInfo{Threads: 12}}},
		{name: `Cpu Usage`,
			input:  rawNodeInfo{a: map[string]any{"cpu": float64(0.0632704976844302)}},
			output: NodeInfo{Cpu: NodeCpuInfo{Usage: 6.32704976844302}}},
		{name: `Name`,
			input:  rawNodeInfo{a: map[string]any{"node": string("pve7")}},
			output: NodeInfo{Name: "pve7"}},
		{name: `Name cached`,
			input: rawNodeInfo{
				a:    map[string]any{},
				node: new(NodeName("pve7"))},
			output: NodeInfo{Name: "pve7"}},
		{name: `Memory TotalBytes`,
			input:  rawNodeInfo{a: map[string]any{"maxmem": float64(134974939136)}},
			output: NodeInfo{Memory: NodeMemoryInfo{TotalBytes: 134974939136}}},
		{name: `Memory UsedBytes`,
			input:  rawNodeInfo{a: map[string]any{"mem": float64(115165569024)}},
			output: NodeInfo{Memory: NodeMemoryInfo{UsedBytes: 115165569024}}},
		{name: `NodePowerState online`,
			input:  rawNodeInfo{a: map[string]any{"status": string("online")}},
			output: NodeInfo{PowerState: NodePowerStateOnline}},
		{name: `NodePowerState offline`,
			input:  rawNodeInfo{a: map[string]any{"status": string("offline")}},
			output: NodeInfo{PowerState: NodePowerStateOffline}},
		{name: `Uptime`,
			input:  rawNodeInfo{a: map[string]any{"uptime": float64(1628483)}},
			output: NodeInfo{Uptime: time.Duration(1628483) * time.Second}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, RawNodeInfo(&test.input).Get())
		})
	}
}

func testData_RawNodesInfo() []struct {
	Name  string
	Input rawNodesInfo
	Array []RawNodeInfo
	Map   map[NodeName]RawNodeInfo
	Len   int
} {
	return []struct {
		Name  string
		Input rawNodesInfo
		Array []RawNodeInfo
		Map   map[NodeName]RawNodeInfo
		Len   int
	}{
		{Name: `Empty members`,
			Input: rawNodesInfo{a: []any{}},
			Array: []RawNodeInfo{},
			Map:   map[NodeName]RawNodeInfo{}},
		{Name: `Single member`,
			Input: rawNodesInfo{a: []any{
				map[string]any{"node": "pve1"}}},
			Array: []RawNodeInfo{
				&rawNodeInfo{a: map[string]any{"node": "pve1"}}},
			Map: map[NodeName]RawNodeInfo{
				"pve1": &rawNodeInfo{
					a:    map[string]any{"node": "pve1"},
					node: new(NodeName("pve1"))}},
			Len: 1},
		{Name: `Multiple members`,
			Input: rawNodesInfo{a: []any{
				map[string]any{"node": "pve1"},
				map[string]any{"node": "pve2"},
				map[string]any{"node": "pve3"}}},
			Array: []RawNodeInfo{
				&rawNodeInfo{a: map[string]any{"node": "pve1"}},
				&rawNodeInfo{a: map[string]any{"node": "pve2"}},
				&rawNodeInfo{a: map[string]any{"node": "pve3"}}},
			Map: map[NodeName]RawNodeInfo{
				"pve1": &rawNodeInfo{
					a:    map[string]any{"node": "pve1"},
					node: new(NodeName("pve1"))},
				"pve2": &rawNodeInfo{
					a:    map[string]any{"node": "pve2"},
					node: new(NodeName("pve2"))},
				"pve3": &rawNodeInfo{
					a:    map[string]any{"node": "pve3"},
					node: new(NodeName("pve3"))}},
			Len: 3},
	}
}

func Test_RawNodesInfo_AsArray(t *testing.T) {
	t.Parallel()
	for _, test := range testData_RawNodesInfo() {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Array, RawNodesInfo(&test.Input).AsArray())
		})
	}
}

func Test_RawNodesInfo_AsMap(t *testing.T) {
	t.Parallel()
	for _, test := range testData_RawNodesInfo() {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Map, RawNodesInfo(&test.Input).AsMap())
		})
	}
}

func Test_RawNodesInfo_Iter(t *testing.T) {
	t.Parallel()
	for _, test := range testData_RawNodesInfo() {
		t.Run(test.Name, func(t *testing.T) {
			testIter(t, RawNodesInfo(&test.Input), test.Array)
		})
	}
}

func Test_RawNodesInfo_Len(t *testing.T) {
	t.Parallel()
	for _, test := range testData_RawNodesInfo() {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Len, RawNodesInfo(&test.Input).Len())
		})
	}
}
