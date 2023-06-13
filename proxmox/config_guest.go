package proxmox

import (
	"strconv"
	"strings"
)

// All code LXC and Qemu have in common should be placed here.

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
	Node               string    `json:"node"`   // TODO custom type
	Pool               string    `json:"pool"`   // TODO custom type
	Status             string    `json:"status"` // TODO custom type?
	Tags               []string  `json:"tags"`   // TODO custom type
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
		if _, isSet := tmpParams["pool"]; isSet {
			resources[i].Pool = tmpParams["pool"].(string)
		}
		if _, isSet := tmpParams["status"]; isSet {
			resources[i].Status = tmpParams["status"].(string)
		}
		if _, isSet := tmpParams["tags"]; isSet {
			resources[i].Tags = strings.Split(tmpParams["tags"].(string), ";")
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

type GuestType string

const (
	GuestLXC  GuestType = "lxc"
	GuestQemu GuestType = "qemu"
)

// Check if there are any pending changes that require a reboot to be applied.
func GuestHasPendingChanges(vmr *VmRef, client *Client) (bool, error) {
	params, err := pendingGuestConfigFromApi(vmr, client)
	if err != nil {
		return false, err
	}
	return keyExists(params, "pending"), nil
}

// Reboot the specified guest
func GuestReboot(vmr *VmRef, client *Client) (err error) {
	_, err = client.ShutdownVm(vmr)
	if err != nil {
		return
	}
	_, err = client.StartVm(vmr)
	return
}

// List all guest the user has viewing rights for in the cluster
func ListGuests(client *Client) ([]GuestResource, error) {
	list, err := client.GetResourceList("vm")
	if err != nil {
		return nil, err
	}
	return GuestResource{}.mapToStruct(list), nil
}

func pendingGuestConfigFromApi(vmr *VmRef, client *Client) ([]interface{}, error) {
	err := vmr.nilCheck()
	if err != nil {
		return nil, err
	}
	if err = client.CheckVmRef(vmr); err != nil {
		return nil, err
	}
	return client.GetItemConfigInterfaceArray("/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/pending", "Guest", "PENDING CONFIG")
}
