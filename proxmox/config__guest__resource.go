package proxmox

import (
	"context"
	"time"
)

// List all guest the user has viewing rights for in the cluster
func ListGuests(ctx context.Context, c *Client) (RawGuestResources, error) {
	return c.new().guestListResources(ctx)
}

func (c *clientNew) guestListResources(ctx context.Context) (RawGuestResources, error) {
	return listGuests_Unsafe(ctx, c.api)

}

func listGuests_Unsafe(ctx context.Context, c clientApiInterface) (RawGuestResources, error) {
	raw, err := c.listGuestResources(ctx)
	if err != nil {
		return nil, err
	}
	resources := make(RawGuestResources, len(raw))
	for i := range raw {
		resources[i] = &rawGuestResource{a: raw[i].(map[string]any)}
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

func (r RawGuestResources) SelectID(id GuestID) (RawGuestResource, error) {
	for i := range r {
		if r[i].GetID() == id {
			return r[i], nil
		}
	}
	return nil, errorMsg{}.guestDoesNotExist(id)
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
	if v, isSet := raw.a["tags"]; isSet {
		return Tags{}.mapToSDK(v.(string))
	}
	return nil
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
