package proxmox

import (
	"iter"
	"time"
)

type NodeInfo struct {
	CertificateFingerprint string
	Name                   NodeName
	Cpu                    NodeCpuInfo
	Memory                 NodeMemoryInfo
	PowerState             NodePowerState
	Uptime                 time.Duration
}

// NodePowerState is an enum for the power state of the node.
//
//	const (
//		NodePowerStateOffline
//		NodePowerStateOnline
//	)
type NodePowerState int

const (
	NodePowerStateOffline NodePowerState = -1
	NodePowerStateOnline  NodePowerState = 1
	NodePowerStateUnknown NodePowerState = 0
)

type NodeCpuInfo struct {
	Threads uint
	Usage   float64
}

type NodeMemoryInfo struct {
	TotalBytes uint
	UsedBytes  uint
}

type RawNodesInfo interface {
	AsArray() []RawNodeInfo
	AsMap() map[NodeName]RawNodeInfo
	Iter() iter.Seq[RawNodeInfo]
	Len() int
}

var _ RawNodesInfo = (*rawNodesInfo)(nil)

type rawNodesInfo struct{ a []any }

func (r *rawNodesInfo) AsArray() []RawNodeInfo {
	raws := make([]RawNodeInfo, len(r.a))
	for i := range r.a {
		raws[i] = &rawNodeInfo{a: r.a[i].(map[string]any)}
	}
	return raws
}

func (r *rawNodesInfo) AsMap() map[NodeName]RawNodeInfo {
	nodes := make(map[NodeName]RawNodeInfo, len(r.a))
	for i := range r.a {
		raw := rawNodeInfo{a: r.a[i].(map[string]any)}
		node := NodeName(raw.a[nodesApiKeyName].(string))
		raw.node = &node
		nodes[node] = &raw
	}
	return nodes
}

func (raw *rawNodesInfo) Iter() iter.Seq[RawNodeInfo] {
	return func(yield func(RawNodeInfo) bool) {
		for i := range raw.a {
			if !yield(&rawNodeInfo{
				a: raw.a[i].(map[string]any),
			}) {
				return
			}
		}
	}
}

func (r *rawNodesInfo) Len() int { return len(r.a) }

type RawNodeInfo interface {
	Get() NodeInfo
	GetCertificateFingerprint() string
	GetCpu() NodeCpuInfo
	GetCpuThreads() uint
	GetCpuUsage() float64
	GetMemory() NodeMemoryInfo
	GetMemoryTotalInBytes() uint
	GetMemoryUsedInBytes() uint
	GetName() NodeName
	GetPowerState() NodePowerState
	GetUptime() time.Duration
}

var _ RawNodeInfo = (*rawNodeInfo)(nil)

type rawNodeInfo struct {
	a    map[string]any
	node *NodeName // Not always set.
}

func (r *rawNodeInfo) Get() NodeInfo {
	return NodeInfo{
		CertificateFingerprint: r.GetCertificateFingerprint(),
		Cpu:                    r.GetCpu(),
		Memory:                 r.GetMemory(),
		Name:                   r.GetName(),
		PowerState:             r.GetPowerState(),
		Uptime:                 r.GetUptime()}
}

func (r *rawNodeInfo) GetCertificateFingerprint() string {
	if v, ok := r.a["ssl_fingerprint"]; ok {
		return v.(string)
	}
	return ""
}

func (r *rawNodeInfo) GetCpu() NodeCpuInfo {
	return NodeCpuInfo{
		Threads: r.GetCpuThreads(),
		Usage:   r.GetCpuUsage()}
}

func (r *rawNodeInfo) GetCpuThreads() uint {
	if v, ok := r.a["maxcpu"]; ok {
		return uint(v.(float64))
	}
	return 0
}

func (r *rawNodeInfo) GetCpuUsage() float64 {
	if v, ok := r.a["cpu"]; ok {
		return v.(float64) * 100
	}
	return 0
}

func (r *rawNodeInfo) GetMemory() NodeMemoryInfo {
	return NodeMemoryInfo{
		TotalBytes: r.GetMemoryTotalInBytes(),
		UsedBytes:  r.GetMemoryUsedInBytes()}
}

func (r *rawNodeInfo) GetMemoryTotalInBytes() uint {
	if v, ok := r.a["maxmem"]; ok {
		return uint(v.(float64))
	}
	return 0
}

func (r *rawNodeInfo) GetMemoryUsedInBytes() uint {
	if v, ok := r.a["mem"]; ok {
		return uint(v.(float64))
	}
	return 0
}

func (r *rawNodeInfo) GetName() NodeName {
	if r.node != nil {
		return *r.node
	}
	if v, ok := r.a[nodesApiKeyName]; ok {
		return NodeName(v.(string))
	}
	return ""
}

func (r *rawNodeInfo) GetPowerState() NodePowerState {
	if v, ok := r.a["status"]; ok {
		switch v.(string) {
		case "online":
			return NodePowerStateOnline
		case "offline":
			return NodePowerStateOffline
		}
	}
	return NodePowerStateUnknown
}

func (r *rawNodeInfo) GetUptime() time.Duration {
	if v, ok := r.a["uptime"]; ok {
		return time.Duration(v.(float64)) * time.Second
	}
	return 0
}

const (
	nodesApiKeyName = "node"
)
