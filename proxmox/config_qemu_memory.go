package proxmox

import (
	"github.com/Telmate/proxmox-api-go/internal/parse"
)

type QemuMemory struct {
	CapacityMiB        *QemuMemoryCapacity        `json:"capacity,omitempty"` // min 1, max 4178944
	MinimumCapacityMiB *QemuMemoryBalloonCapacity `json:"balloon,omitempty"`  // 0 to clear (balloon), max 4178944
	Shares             *QemuMemoryShares          `json:"shares,omitempty"`   // 0 to clear, max 50000
}

func (config QemuMemory) mapToAPI(current *QemuMemory, params map[string]interface{}) string {
	if current == nil { // create
		if config.CapacityMiB != nil {
			params["memory"] = *config.CapacityMiB
		}
		if config.MinimumCapacityMiB != nil {
			params["balloon"] = *config.MinimumCapacityMiB
			if config.CapacityMiB == nil {
				params["memory"] = *config.MinimumCapacityMiB
			}
		}
		if config.Shares != nil {
			if *config.Shares > 0 {
				params["shares"] = *config.Shares
			}
		}
		return ""
	}
	// update
	if config.CapacityMiB != nil {
		params["memory"] = *config.CapacityMiB
		if config.MinimumCapacityMiB == nil && current.MinimumCapacityMiB != nil && uint32(*current.MinimumCapacityMiB) > uint32(*config.CapacityMiB) {
			params["balloon"] = *config.CapacityMiB
			return ",shares"
		}
	}
	if config.MinimumCapacityMiB != nil {
		params["balloon"] = *config.MinimumCapacityMiB
		if *config.MinimumCapacityMiB == 0 {
			return ",shares"
		}
	}
	if config.Shares != nil {
		if *config.Shares == 0 {
			return ",shares"
		}
		params["shares"] = *config.Shares
	}
	return ""
}

func (QemuMemory) mapToSDK(params map[string]interface{}) *QemuMemory {
	config := QemuMemory{}
	if v, isSet := params["memory"]; isSet {
		tmp, _ := parse.Uint(v)
		tmpIntermediate := QemuMemoryCapacity(tmp)
		config.CapacityMiB = &tmpIntermediate
	}
	if v, isSet := params["balloon"]; isSet {
		tmp, _ := parse.Uint(v)
		tmpIntermediate := QemuMemoryBalloonCapacity(tmp)
		config.MinimumCapacityMiB = &tmpIntermediate
	}
	if v, isSet := params["shares"]; isSet {
		tmp, _ := parse.Uint(v)
		tmpIntermediate := QemuMemoryShares(tmp)
		config.Shares = &tmpIntermediate
	}
	return &config
}

type QemuMemoryBalloonCapacity uint32 // max 4178944

type QemuMemoryCapacity uint32 // max 4178944

type QemuMemoryShares uint16 // max 50000
