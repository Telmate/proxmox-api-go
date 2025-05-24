package proxmox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type CpuArchitecture string

type OperatingSystem string

type ConfigLXC struct {
	Architecture    CpuArchitecture `json:"architecture"` // only returned
	Description     *string         `json:"description,omitempty"`
	ID              *GuestID        `json:"id"` // only used during creation
	Memory          *LxcMemory      `json:"memory,omitempty"`
	Name            *GuestName      `json:"name,omitempty"`
	Node            *NodeName       `json:"node,omitempty"` // only used during creation
	OperatingSystem OperatingSystem `json:"os"`             // only returned
	Pool            *PoolName       `json:"pool,omitempty"`
	Privileged      *bool           `json:"privileged,omitempty"` // only used during creation
	Tags            *Tags           `json:"tags,omitempty"`
}

const (
	ConfigLXC_Error_NoSettingsSpecified = "no settings specified"
)

func (config ConfigLXC) Create(ctx context.Context, c *Client) (*VmRef, error) {
	if err := config.Validate(nil); err != nil {
		return nil, err
	}
	return config.CreateNoCheck(ctx, c)
}

func (config ConfigLXC) CreateNoCheck(ctx context.Context, c *Client) (*VmRef, error) {
	params, pool := config.mapToApiCreate()

	var err error
	var id GuestID
	var node NodeName
	if config.Node != nil {
		node = *config.Node
	}
	url := "/nodes/" + node.String() + "/qemu"
	if config.ID == nil {
		id, err = guestCreateLoop(ctx, "vmid", url, params, c)
		if err != nil {
			return nil, err
		}
	} else {
		var exitStatus string
		exitStatus, err = c.PostWithTask(ctx, params, url)
		if err != nil {
			return nil, fmt.Errorf("error creating LXC: %v, error status: %s (params: %v)", err, exitStatus, params)
		}
	}

	return &VmRef{
		node:   node,
		vmId:   id,
		pool:   pool,
		vmType: vmRefQemu,
	}, nil
}

func (config ConfigLXC) mapToApiCreate() (map[string]any, PoolName) {
	params := config.mapToApiShared()
	var pool PoolName
	if config.Description != nil && *config.Description != "" {
		params[lxcApiKeyDescription] = *config.Description
	}
	if config.Memory != nil {
		params[lxcApiKeyMemory] = *config.Memory
	}
	if config.Name != nil {
		params[lxcApiKeyName] = (*config.Name).String()
	}
	if config.Pool != nil {
		pool = *config.Pool
		params[lxcApiKeyPool] = string(pool)
	}
	if config.Privileged != nil && !*config.Privileged {
		params[lxcApiKeyUnprivileged] = 1
	}
	if config.ID != nil {
		params[lxcApiKeyGuestID] = *config.ID
	}
	if config.Tags != nil {
		params[lxcApiKeyTags] = (*config.Tags).mapToApiCreate()
	}
	return params, pool
}

func (config ConfigLXC) mapToApiShared() map[string]any {
	params := make(map[string]any)
	return params
}

func (config ConfigLXC) mapToApiUpdate(current ConfigLXC) map[string]any {
	params := config.mapToApiShared()
	var delete string
	if config.Description != nil {
		if current.Description == nil || *config.Description != *current.Description {
			if *config.Description == "" {
				delete += "," + lxcApiKeyDescription
			} else {
				params[lxcApiKeyDescription] = *config.Description
			}
		}
	}
	if config.Memory != nil && (current.Memory == nil || *config.Memory != *current.Memory) {
		params[lxcApiKeyMemory] = *config.Memory
	}
	if config.Name != nil {
		if current.Name == nil || *config.Name != *current.Name {
			params[lxcApiKeyName] = (*config.Name).String()
		}
	}
	if config.Tags != nil {
		if v, ok := (*config.Tags).mapToApiUpdate(current.Tags); ok {
			params[lxcApiKeyTags] = v
		}
	}
	if delete != "" {
		params[lxcApiKeyDelete] = strings.TrimPrefix(delete, ",")
	}
	return params
}

func (ConfigLXC) mapToSDK(params map[string]any, vmr VmRef) *ConfigLXC {
	var privileged bool
	config := ConfigLXC{
		ID:         util.Pointer(vmr.vmId),
		Node:       util.Pointer(vmr.node),
		Privileged: &privileged}
	if vmr.pool != "" {
		config.Pool = util.Pointer(PoolName(vmr.pool))
	}
	if v, isSet := params[lxcApiKeyArchitecture]; isSet {
		config.Architecture = CpuArchitecture(v.(string))
	}
	if v, isSet := params[lxcApiKeyDescription]; isSet {
		config.Description = util.Pointer(v.(string))
	}
	if v, isSet := params[lxcApiKeyMemory]; isSet {
		config.Memory = util.Pointer(LxcMemory(v.(float64)))
	}
	if v, isSet := params[lxcApiKeyName]; isSet {
		config.Name = util.Pointer(GuestName(v.(string)))
	}
	if v, isSet := params[lxcApiKeyOperatingSystem]; isSet {
		config.OperatingSystem = OperatingSystem(v.(string))
	}
	if v, isSet := params[lxcApiKeyUnprivileged]; isSet {
		privileged = v.(float64) == 0
	}
	if v, isSet := params[lxcApiKeyTags]; isSet {
		config.Tags = util.Pointer(Tags{}.mapToSDK(v.(string)))
	}
	return &config
}

func (config ConfigLXC) Update(ctx context.Context, vmr *VmRef, c *Client) error {
	current, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	if err := config.Validate(current); err != nil {
		return err
	}
	return config.updateNoCheck(ctx, vmr, current, c)
}

func (config ConfigLXC) UpdateNoCheck(ctx context.Context, vmr *VmRef, c *Client) error {
	current, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	return config.updateNoCheck(ctx, vmr, current, c)
}

func (config ConfigLXC) updateNoCheck(ctx context.Context, vmr *VmRef, current *ConfigLXC, c *Client) error {
	params := config.mapToApiUpdate(*current)
	if len(params) == 0 {
		return errors.New(ConfigLXC_Error_NoSettingsSpecified)
	}
	return c.Put(ctx, params, "/nodes/"+vmr.node.String()+"/lxc/"+vmr.vmId.String()+"/config")
}

func (config ConfigLXC) Validate(current *ConfigLXC) (err error) {
	if current == nil { // Create
		err = config.validateCreate()
	} else { // Update
		err = config.validateUpdate()
	}
	if config.ID != nil {
		if err = config.ID.Validate(); err != nil {
			return
		}
	}
	if config.Memory != nil {
		if err = config.Memory.Validate(); err != nil {
			return
		}
	}
	if config.Name != nil {
		if err = config.Name.Validate(); err != nil {
			return
		}
	}
	if config.Node != nil {
		if err = config.Node.Validate(); err != nil {
			return
		}
	}
	if config.Pool != nil {
		if err = config.Pool.Validate(); err != nil {
			return
		}
	}
	if config.Tags != nil {
		if err = config.Tags.Validate(); err != nil {
			return
		}
	}
	return
}

func (config ConfigLXC) validateCreate() (err error) {
	return
}

func (config ConfigLXC) validateUpdate() (err error) {
	return
}

const (
	lxcApiKeyArchitecture    string = "arch"
	lxcApiKeyDelete          string = "delete"
	lxcApiKeyDescription     string = "description"
	lxcApiKeyGuestID         string = "vmid"
	lxcApiKeyMemory          string = "memory"
	lxcApiKeyName            string = "name"
	lxcApiKeyOperatingSystem string = "ostype"
	lxcApiKeyPool            string = "pool"
	lxcApiKeySwap            string = "swap"
	lxcApiKeyTags            string = "tags"
	lxcApiKeyUnprivileged    string = "unprivileged"
)

type LxcMemory uint

const (
	LxcMemoryMinimum        = 16
	LxcMemory_Error_Minimum = "memory has a minimum of 16"
)

func (memory LxcMemory) Validate() error {
	if memory < LxcMemoryMinimum {
		return errors.New(LxcMemory_Error_Minimum)
	}
	return nil
}

func (memory LxcMemory) String() string { return strconv.Itoa(int(memory)) } // String is for fmt.Stringer.

func NewConfigLXCFromApi(ctx context.Context, vmr *VmRef, c *Client) (*ConfigLXC, error) {
	rawConfig, err := c.GetVmConfig(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return ConfigLXC{}.mapToSDK(rawConfig, *vmr), nil
}
