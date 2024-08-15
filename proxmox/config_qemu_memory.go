package proxmox

import (
	"errors"

	"github.com/Telmate/proxmox-api-go/internal/parse"
)

type QemuMemory struct {
	CapacityMiB        *QemuMemoryCapacity        `json:"capacity,omitempty"` // min 1, max 4178944
	MinimumCapacityMiB *QemuMemoryBalloonCapacity `json:"balloon,omitempty"`  // 0 to clear (balloon), max 4178944
	Shares             *QemuMemoryShares          `json:"shares,omitempty"`   // 0 to clear, max 50000
}

const (
	QemuMemory_Error_MinimumCapacityMiB_GreaterThan_CapacityMiB string = "minimum capacity MiB cannot be greater than capacity MiB"
	QemuMemory_Error_NoMemoryCapacity                           string = "no memory capacity specified"
	QemuMemory_Error_SharesHasNoEffectWithoutBallooning         string = "shares has no effect when capacity equals minimum capacity"
)

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

func (config QemuMemory) Validate(current *QemuMemory) error {
	var eventualCapacityMiB QemuMemoryCapacity
	var eventualMinimumCapacityMiB QemuMemoryBalloonCapacity
	if config.MinimumCapacityMiB != nil {
		if err := config.MinimumCapacityMiB.Validate(); err != nil {
			return err
		}
		if config.CapacityMiB != nil && QemuMemoryCapacity(*config.MinimumCapacityMiB) > *config.CapacityMiB {
			return errors.New(QemuMemory_Error_MinimumCapacityMiB_GreaterThan_CapacityMiB)
		}
		eventualMinimumCapacityMiB = *config.MinimumCapacityMiB
		eventualCapacityMiB = QemuMemoryCapacity(eventualMinimumCapacityMiB)
	} else if current != nil && current.MinimumCapacityMiB != nil {
		eventualMinimumCapacityMiB = *current.MinimumCapacityMiB
	}
	if config.CapacityMiB != nil {
		if err := config.CapacityMiB.Validate(); err != nil {
			return err
		}
		eventualCapacityMiB = *config.CapacityMiB
	} else if current != nil && current.CapacityMiB != nil {
		eventualCapacityMiB = *current.CapacityMiB
	}
	if eventualCapacityMiB == 0 {
		return errors.New(QemuMemory_Error_NoMemoryCapacity)
	}
	if config.Shares != nil {
		if err := config.Shares.Validate(); err != nil {
			return err
		}
		if *config.Shares > 0 {
			if eventualCapacityMiB == QemuMemoryCapacity(eventualMinimumCapacityMiB) {
				return errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)
			}
		}
	}
	return nil
}

type QemuMemoryBalloonCapacity uint32 // max 4178944

const (
	QemuMemoryBalloonCapacity_Error_Invalid string                    = "memory balloon capacity has a maximum of 4178944"
	qemuMemoryBalloonCapacity_Max           QemuMemoryBalloonCapacity = 4178944
)

func (m QemuMemoryBalloonCapacity) Validate() error {
	if m > qemuMemoryBalloonCapacity_Max {
		return errors.New(QemuMemoryBalloonCapacity_Error_Invalid)
	}
	return nil
}

type QemuMemoryCapacity uint32 // max 4178944

const (
	QemuMemoryCapacity_Error_Maximum string             = "memory capacity has a maximum of 4178944"
	QemuMemoryCapacity_Error_Minimum string             = "memory capacity has a minimum of 1"
	qemuMemoryCapacity_Max           QemuMemoryCapacity = 4178944
)

func (m QemuMemoryCapacity) Validate() error {
	if m == 0 {
		return errors.New(QemuMemoryCapacity_Error_Minimum)
	}
	if m > qemuMemoryCapacity_Max {
		return errors.New(QemuMemoryCapacity_Error_Maximum)
	}
	return nil
}

type QemuMemoryShares uint16 // max 50000

const (
	QemuMemoryShares_Error_Invalid string           = "memory shares has a maximum of 50000"
	qemuMemoryShares_Max           QemuMemoryShares = 50000
)

func (m QemuMemoryShares) Validate() error {
	if m > qemuMemoryShares_Max {
		return errors.New(QemuMemoryShares_Error_Invalid)
	}
	return nil
}
