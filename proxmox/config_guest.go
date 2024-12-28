package proxmox

import (
	"context"
	"errors"
	"net/netip"
	"strconv"
)

// All code LXC and Qemu have in common should be placed here.

type GuestDNS struct {
	NameServers  *[]netip.Addr `json:"nameservers,omitempty"`
	SearchDomain *string       `json:"searchdomain,omitempty"` // we are not validating this field, as validating domain names is a complex topic.
}

type GuestResource struct {
	CpuCores           uint      `json:"cpu_cores"`
	CpuUsage           float64   `json:"cpu_usage"`
	DiskReadTotal      uint      `json:"disk_read"`
	DiskSizeInBytes    uint      `json:"disk_size"`
	DiskUsedInBytes    uint      `json:"disk_used"`
	DiskWriteTotal     uint      `json:"disk_write"`
	HaState            string    `json:"hastate"` // TODO custom type?
	Id                 uint      `json:"id"`
	MemoryTotalInBytes uint      `json:"memory_total"`
	MemoryUsedInBytes  uint      `json:"memory_used"`
	Name               string    `json:"name"` // TODO custom type
	NetworkIn          uint      `json:"network_in"`
	NetworkOut         uint      `json:"network_out"`
	Node               string    `json:"node"` // TODO custom type
	Pool               PoolName  `json:"pool"`
	Status             string    `json:"status"` // TODO custom type?
	Tags               []Tag     `json:"tags"`
	Template           bool      `json:"template"`
	Type               GuestType `json:"type"`
	UptimeInSeconds    uint      `json:"uptime"`
}

// https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/resources
func (GuestResource) mapToStruct(params []interface{}) []GuestResource {
	if len(params) == 0 {
		return nil
	}
	resources := make([]GuestResource, len(params))
	for i := range params {
		tmpParams := params[i].(map[string]interface{})
		if _, isSet := tmpParams["maxcpu"]; isSet {
			resources[i].CpuCores = uint(tmpParams["maxcpu"].(float64))
		}
		if _, isSet := tmpParams["cpu"]; isSet {
			resources[i].CpuUsage = tmpParams["cpu"].(float64)
		}
		if _, isSet := tmpParams["diskread"]; isSet {
			resources[i].DiskReadTotal = uint(tmpParams["diskread"].(float64))
		}
		if _, isSet := tmpParams["maxdisk"]; isSet {
			resources[i].DiskSizeInBytes = uint(tmpParams["maxdisk"].(float64))
		}
		if _, isSet := tmpParams["disk"]; isSet {
			resources[i].DiskUsedInBytes = uint(tmpParams["disk"].(float64))
		}
		if _, isSet := tmpParams["diskwrite"]; isSet {
			resources[i].DiskWriteTotal = uint(tmpParams["diskwrite"].(float64))
		}
		if _, isSet := tmpParams["hastate"]; isSet {
			resources[i].HaState = tmpParams["hastate"].(string)
		}
		if _, isSet := tmpParams["vmid"]; isSet {
			resources[i].Id = uint(tmpParams["vmid"].(float64))
		}
		if _, isSet := tmpParams["maxmem"]; isSet {
			resources[i].MemoryTotalInBytes = uint(tmpParams["maxmem"].(float64))
		}
		if _, isSet := tmpParams["mem"]; isSet {
			resources[i].MemoryUsedInBytes = uint(tmpParams["mem"].(float64))
		}
		if _, isSet := tmpParams["name"]; isSet {
			resources[i].Name = tmpParams["name"].(string)
		}
		if _, isSet := tmpParams["netin"]; isSet {
			resources[i].NetworkIn = uint(tmpParams["netin"].(float64))
		}
		if _, isSet := tmpParams["netout"]; isSet {
			resources[i].NetworkOut = uint(tmpParams["netout"].(float64))
		}
		if _, isSet := tmpParams["node"]; isSet {
			resources[i].Node = tmpParams["node"].(string)
		}
		if _, isSet := tmpParams["status"]; isSet {
			resources[i].Status = tmpParams["status"].(string)
		}
		if _, isSet := tmpParams["tags"]; isSet {
			resources[i].Tags = Tag("").mapToSDK(tmpParams["tags"].(string))
		}
		if _, isSet := tmpParams["template"]; isSet {
			resources[i].Template = Itob(int(tmpParams["template"].(float64)))
		}
		if _, isSet := tmpParams["type"]; isSet {
			resources[i].Type = GuestType(tmpParams["type"].(string))
		}
		if _, isSet := tmpParams["uptime"]; isSet {
			resources[i].UptimeInSeconds = uint(tmpParams["uptime"].(float64))
		}
	}
	return resources
}

// Enum
type GuestFeature string

const (
	GuestFeature_Clone    GuestFeature = "clone"
	GuestFeature_Copy     GuestFeature = "copy"
	GuestFeature_Snapshot GuestFeature = "snapshot"
)

func (GuestFeature) Error() error {
	return errors.New("value should be one of (" + string(GuestFeature_Clone) + " ," + string(GuestFeature_Copy) + " ," + string(GuestFeature_Snapshot) + ")")
}

func (GuestFeature) mapToStruct(params map[string]interface{}) bool {
	if value, isSet := params["hasFeature"]; isSet {
		return Itob(int(value.(float64)))
	}
	return false
}

func (feature GuestFeature) Validate() error {
	switch feature {
	case GuestFeature_Copy, GuestFeature_Clone, GuestFeature_Snapshot:
		return nil
	}
	return GuestFeature("").Error()
}

type GuestFeatures struct {
	Clone    bool `json:"clone"`
	Copy     bool `json:"copy"`
	Snapshot bool `json:"snapshot"`
}

type GuestType string

const (
	GuestLXC  GuestType = "lxc"
	GuestQemu GuestType = "qemu"
)

// check if the guest has the specified feature.
func GuestHasFeature(ctx context.Context, vmr *VmRef, client *Client, feature GuestFeature) (bool, error) {
	err := feature.Validate()
	if err != nil {
		return false, err
	}
	err = client.CheckVmRef(ctx, vmr)
	if err != nil {
		return false, err
	}
	return guestHasFeature(ctx, vmr, client, feature)
}

func guestHasFeature(ctx context.Context, vmr *VmRef, client *Client, feature GuestFeature) (bool, error) {
	var params map[string]interface{}
	params, err := client.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/feature?feature=snapshot", "guest", "FEATURES")
	if err != nil {
		return false, err
	}
	return GuestFeature("").mapToStruct(params), nil
}

// Check if there are any pending changes that require a reboot to be applied.
func GuestHasPendingChanges(ctx context.Context, vmr *VmRef, client *Client) (bool, error) {
	params, err := pendingGuestConfigFromApi(ctx, vmr, client)
	if err != nil {
		return false, err
	}
	return keyExists(params, "pending") || keyExists(params, "delete"), nil
}

// Reboot the specified guest
func GuestReboot(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	_, err = client.RebootVm(ctx, vmr)
	return
}

func guestSetPool_Unsafe(ctx context.Context, c *Client, guestID uint, newPool PoolName, currentPool *PoolName, version Version) (err error) {
	if newPool == "" {
		if *currentPool != "" { // leave pool
			if err = (*currentPool).removeGuests_Unsafe(ctx, c, []uint{guestID}, version); err != nil {
				return
			}
		}
	} else {
		if *currentPool == "" { // join pool
			if version.Smaller(Version{8, 0, 0}) {
				if err = newPool.addGuests_UnsafeV7(ctx, c, []uint{guestID}); err != nil {
					return
				}
			} else {
				newPool.addGuests_UnsafeV8(ctx, c, []uint{guestID})
			}
		} else if newPool != *currentPool { // change pool
			if version.Smaller(Version{8, 0, 0}) {
				if err = (*currentPool).removeGuests_Unsafe(ctx, c, []uint{guestID}, version); err != nil {
					return
				}
				if err = newPool.addGuests_UnsafeV7(ctx, c, []uint{guestID}); err != nil {
					return
				}
			} else {
				if err = newPool.addGuests_UnsafeV8(ctx, c, []uint{guestID}); err != nil {
					return
				}
			}
		}
	}
	return
}

func GuestShutdown(ctx context.Context, vmr *VmRef, client *Client, force bool) (err error) {
	if err = client.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	var params map[string]interface{}
	if force {
		params = map[string]interface{}{"forceStop": force}
	}
	_, err = client.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/status/shutdown")
	return
}

func GuestStart(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	_, err = client.StartVm(ctx, vmr)
	return
}

// List all features the guest has.
func ListGuestFeatures(ctx context.Context, vmr *VmRef, client *Client) (features GuestFeatures, err error) {
	err = client.CheckVmRef(ctx, vmr)
	if err != nil {
		return
	}
	features.Clone, err = guestHasFeature(ctx, vmr, client, GuestFeature_Clone)
	if err != nil {
		return
	}
	features.Copy, err = guestHasFeature(ctx, vmr, client, GuestFeature_Copy)
	if err != nil {
		return
	}
	features.Snapshot, err = guestHasFeature(ctx, vmr, client, GuestFeature_Snapshot)
	return
}

// List all guest the user has viewing rights for in the cluster
func ListGuests(ctx context.Context, client *Client) ([]GuestResource, error) {
	list, err := client.GetResourceList(ctx, "vm")
	if err != nil {
		return nil, err
	}
	return GuestResource{}.mapToStruct(list), nil
}

func pendingGuestConfigFromApi(ctx context.Context, vmr *VmRef, client *Client) ([]interface{}, error) {
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return nil, err
	}
	return client.GetItemConfigInterfaceArray(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/pending", "Guest", "PENDING CONFIG")
}
