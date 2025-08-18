package proxmox

import (
	"context"
	"time"
)

// List all guest the user has viewing rights for in the cluster
func ListGuests(ctx context.Context, c *Client) (RawGuestResources, error) {
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	return listGuests_Unsafe(ctx, c)
}

func listGuests_Unsafe(ctx context.Context, c *Client) (RawGuestResources, error) {
	raw, err := c.getResourceList_Unsafe(ctx, resourceListGuest)
	if err != nil {
		return nil, err
	}
	resources := make(RawGuestResources, len(raw))
	for i := range raw {
		resources[i] = raw[i].(map[string]any)
	}
	return resources, nil
}

type RawGuestResources []RawGuestResource

func (r RawGuestResources) Get() []GuestResource {
	resources := make([]GuestResource, len(r))
	for i := range r {
		resources[i] = r[i].Get()
	}
	return resources
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
	Template           bool          `json:"template"`
	Type               GuestType     `json:"type"`
	Uptime             time.Duration `json:"uptime"`
}

type RawGuestResource map[string]any

// https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/resources
func (r RawGuestResource) Get() GuestResource {
	return GuestResource{
		CpuCores:           r.GetCPUcores(),
		CpuUsage:           r.GetCPUusage(),
		DiskReadTotal:      r.GetDiskReadTotal(),
		DiskSizeInBytes:    r.GetDiskSizeInBytes(),
		DiskUsedInBytes:    r.GetDiskUsedInBytes(),
		DiskWriteTotal:     r.GetDiskWriteTotal(),
		HaState:            r.GetHaState(),
		ID:                 r.GetID(),
		MemoryTotalInBytes: r.GetMemoryTotalInBytes(),
		MemoryUsedInBytes:  r.GetMemoryUsedInBytes(),
		Name:               r.GetName(),
		NetworkIn:          r.GetNetworkIn(),
		NetworkOut:         r.GetNetworkOut(),
		Node:               r.GetNode(),
		Pool:               r.GetPool(),
		Status:             r.GetStatus(),
		Tags:               r.GetTags(),
		Template:           r.GetTemplate(),
		Type:               r.GetType(),
		Uptime:             r.GetUptime()}
}

func (r RawGuestResource) GetCPUcores() uint {
	if v, isSet := r["maxcpu"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetCPUusage() float64 {
	if v, isSet := r["cpu"]; isSet {
		return v.(float64)
	}
	return 0
}

func (r RawGuestResource) GetDiskReadTotal() uint {
	if v, isSet := r["diskread"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetDiskSizeInBytes() uint {
	if v, isSet := r["maxdisk"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetDiskUsedInBytes() uint {
	if v, isSet := r["disk"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetDiskWriteTotal() uint {
	if v, isSet := r["diskwrite"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetHaState() string {
	if v, isSet := r["hastate"]; isSet {
		return v.(string)
	}
	return ""
}

func (r RawGuestResource) GetID() GuestID {
	if v, isSet := r["vmid"]; isSet {
		return GuestID(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetMemoryTotalInBytes() uint {
	if v, isSet := r["maxmem"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetMemoryUsedInBytes() uint {
	if v, isSet := r["mem"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetName() GuestName {
	if v, isSet := r["name"]; isSet {
		return GuestName(v.(string))
	}
	return ""
}

func (r RawGuestResource) GetNetworkIn() uint {
	if v, isSet := r["netin"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetNetworkOut() uint {
	if v, isSet := r["netout"]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (r RawGuestResource) GetNode() NodeName {
	if v, isSet := r["node"]; isSet {
		return NodeName(v.(string))
	}
	return ""
}

func (r RawGuestResource) GetPool() PoolName {
	if v, isSet := r["pool"]; isSet {
		return PoolName(v.(string))
	}
	return ""
}

func (r RawGuestResource) GetStatus() PowerState {
	if v, isSet := r["status"]; isSet {
		return PowerState(0).parse(v.(string))
	}
	return PowerStateUnknown
}

func (r RawGuestResource) GetTags() Tags {
	if v, isSet := r["tags"]; isSet {
		return Tags{}.mapToSDK(v.(string))
	}
	return nil
}

func (r RawGuestResource) GetTemplate() bool {
	if v, isSet := r["template"]; isSet {
		return int(v.(float64)) == 1
	}
	return false
}

func (r RawGuestResource) GetType() GuestType {
	if v, isSet := r["type"]; isSet {
		return GuestType(v.(string))
	}
	return ""
}

func (r RawGuestResource) GetUptime() time.Duration {
	if v, isSet := r["uptime"]; isSet {
		return time.Duration(v.(float64)) * time.Second
	}
	return 0
}
