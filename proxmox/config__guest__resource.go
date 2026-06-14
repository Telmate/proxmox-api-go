package proxmox

import (
	"context"
	"iter"
	"time"
)

func (c *guestClient) List(ctx context.Context) (RawGuestResources, error) { return c.ListNoCheck(ctx) }

func (c *guestClient) ListNoCheck(ctx context.Context) (RawGuestResources, error) {
	return c.api.listGuestResources(ctx)
}

// Deprecated: use GuestInterface.List() instead.
// List all guest the user has viewing rights for in the cluster
func ListGuests(ctx context.Context, c *Client) (RawGuestResources, error) {
	return c.New().Guest.ListNoCheck(ctx)
}

type RawGuestResources interface {
	AsArray() []RawGuestResource
	AsMap() map[GuestID]RawGuestResource
	Iter() iter.Seq[RawGuestResource]
	Len() int
}

var _ RawGuestResources = (*rawGuestResources)(nil)

type rawGuestResources struct{ a []any }

func (r *rawGuestResources) AsArray() []RawGuestResource {
	new := make([]RawGuestResource, len(r.a))
	for i := range r.a {
		new[i] = &rawGuestResource{a: r.a[i].(map[string]any)}
	}
	return new
}

func (r *rawGuestResources) AsMap() map[GuestID]RawGuestResource {
	new := make(map[GuestID]RawGuestResource, len(r.a))
	for i := range r.a {
		raw := r.a[i].(map[string]any)
		new[GuestID(raw["vmid"].(float64))] = &rawGuestResource{a: raw}
	}
	return new
}

func (r *rawGuestResources) asMap() map[GuestID]*rawGuestResource {
	new := make(map[GuestID]*rawGuestResource, len(r.a))
	for i := range r.a {
		raw := r.a[i].(map[string]any)
		new[GuestID(raw["vmid"].(float64))] = &rawGuestResource{a: raw}
	}
	return new
}

func (raw *rawGuestResources) Iter() iter.Seq[RawGuestResource] {
	return func(yield func(RawGuestResource) bool) {
		for i := range raw.a {
			if !yield(&rawGuestResource{
				a: raw.a[i].(map[string]any),
			}) {
				return
			}
		}
	}
}

func (r *rawGuestResources) Len() int { return len(r.a) }

func (r *rawGuestResources) selectID(id GuestID) (*rawGuestResource, bool) {
	for i := range r.a {
		raw := r.a[i].(map[string]any)
		if v := raw["vmid"]; GuestID(v.(float64)) == id {
			return &rawGuestResource{a: raw}, true
		}
	}
	return nil, false
}

type GuestResource struct {
	CpuCores           uint          `json:"cpu_cores"`
	CpuUsage           float64       `json:"cpu_usage"`
	DiskReadTotal      uint          `json:"disk_read"`
	DiskSizeInBytes    uint          `json:"disk_size"`
	DiskUsedInBytes    uint          `json:"disk_used"`
	DiskWriteTotal     uint          `json:"disk_write"`
	HaState            string        `json:"hastate"` // TODO custom type?
	ID                 GuestID       `json:"id"`
	MemoryTotalInBytes uint          `json:"memory_total"`
	MemoryUsedInBytes  uint          `json:"memory_used"`
	Name               GuestName     `json:"name"`
	NetworkIn          uint          `json:"network_in"`
	NetworkOut         uint          `json:"network_out"`
	Node               NodeName      `json:"node"`
	Pool               PoolName      `json:"pool"`
	Status             PowerState    `json:"status"`
	Tags               Tags          `json:"tags"`
	Locked             bool          `json:"locked"`
	Template           bool          `json:"template"`
	Type               GuestType     `json:"type"`
	Uptime             time.Duration `json:"uptime"`
}

type RawGuestResource interface {
	Get() GuestResource
	GetCPUcores() uint
	GetCPUusage() float64
	GetDiskReadTotal() uint
	GetDiskSizeInBytes() uint
	GetDiskUsedInBytes() uint
	GetDiskWriteTotal() uint
	GetHaState() string
	GetID() GuestID
	GetLocked() bool
	GetMemoryTotalInBytes() uint
	GetMemoryUsedInBytes() uint
	GetName() GuestName
	GetNetworkIn() uint
	GetNetworkOut() uint
	GetNode() NodeName
	GetPool() PoolName
	GetStatus() PowerState
	GetTags() Tags
	GetTemplate() bool
	GetType() GuestType
	GetUptime() time.Duration
}

type rawGuestResource struct{ a map[string]any }

// https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/resources
func (raw *rawGuestResource) Get() GuestResource {
	return GuestResource{
		CpuCores:           raw.GetCPUcores(),
		CpuUsage:           raw.GetCPUusage(),
		DiskReadTotal:      raw.GetDiskReadTotal(),
		DiskSizeInBytes:    raw.GetDiskSizeInBytes(),
		DiskUsedInBytes:    raw.GetDiskUsedInBytes(),
		DiskWriteTotal:     raw.GetDiskWriteTotal(),
		HaState:            raw.GetHaState(),
		ID:                 raw.GetID(),
		Locked:             raw.GetLocked(),
		MemoryTotalInBytes: raw.GetMemoryTotalInBytes(),
		MemoryUsedInBytes:  raw.GetMemoryUsedInBytes(),
		Name:               raw.GetName(),
		NetworkIn:          raw.GetNetworkIn(),
		NetworkOut:         raw.GetNetworkOut(),
		Node:               raw.GetNode(),
		Pool:               raw.GetPool(),
		Status:             raw.GetStatus(),
		Tags:               raw.GetTags(),
		Template:           raw.GetTemplate(),
		Type:               raw.GetType(),
		Uptime:             raw.GetUptime()}
}

func (raw *rawGuestResource) GetCPUcores() uint {
	if v, isSet := raw.a["maxcpu"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetCPUusage() float64 {
	if v, isSet := raw.a["cpu"]; isSet {
		return v.(float64)
	}
	return 0
}

func (raw *rawGuestResource) GetDiskReadTotal() uint {
	if v, isSet := raw.a["diskread"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetDiskSizeInBytes() uint {
	if v, isSet := raw.a["maxdisk"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetDiskUsedInBytes() uint {
	if v, isSet := raw.a["disk"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetDiskWriteTotal() uint {
	if v, isSet := raw.a["diskwrite"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetHaState() string {
	if v, isSet := raw.a["hastate"]; isSet {
		return v.(string)
	}
	return ""
}

func (raw *rawGuestResource) GetID() GuestID {
	if v, isSet := raw.a["vmid"]; isSet {
		return GuestID(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetLocked() bool {
	_, isSet := raw.a["lock"]
	return isSet
}

func (raw *rawGuestResource) GetMemoryTotalInBytes() uint {
	if v, isSet := raw.a["maxmem"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetMemoryUsedInBytes() uint {
	if v, isSet := raw.a["mem"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetName() GuestName {
	if v, isSet := raw.a["name"]; isSet {
		return GuestName(v.(string))
	}
	return ""
}

func (raw *rawGuestResource) GetNetworkIn() uint {
	if v, isSet := raw.a["netin"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetNetworkOut() uint {
	if v, isSet := raw.a["netout"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawGuestResource) GetNode() NodeName {
	if v, isSet := raw.a["node"]; isSet {
		return NodeName(v.(string))
	}
	return ""
}

func (raw *rawGuestResource) GetPool() PoolName {
	if v, isSet := raw.a["pool"]; isSet {
		return PoolName(v.(string))
	}
	return ""
}

func (raw *rawGuestResource) GetStatus() PowerState {
	if v, isSet := raw.a["status"]; isSet {
		return PowerState(0).parse(v.(string))
	}
	return PowerStateUnknown
}

func (raw *rawGuestResource) GetTags() Tags {
	var t Tags
	if v, isSet := raw.a["tags"]; isSet {
		t.mapToSDK(v.(string))
	}
	return t
}

func (raw *rawGuestResource) GetTemplate() bool {
	if v, isSet := raw.a["template"]; isSet {
		return int(v.(float64)) == 1
	}
	return false
}

func (raw *rawGuestResource) GetType() GuestType {
	if v, isSet := raw.a["type"]; isSet {
		switch v.(string) {
		case "lxc":
			return GuestLxc
		case "qemu":
			return GuestQemu
		}
	}
	return guestUnknown
}

func (raw *rawGuestResource) GetUptime() time.Duration {
	if v, isSet := raw.a["uptime"]; isSet {
		return time.Duration(v.(float64)) * time.Second
	}
	return 0
}
