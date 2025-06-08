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

func (arch CpuArchitecture) String() string { // String is for fmt.Stringer.
	return string(arch)
}

type OperatingSystem string

type ConfigLXC struct {
	Architecture    CpuArchitecture   `json:"architecture"` // only returned
	BootMount       *LxcBootMount     `json:"boot_mount,omitempty"`
	CPU             *LxcCPU           `json:"cpu,omitempty"`
	CreateOptions   *LxcCreateOptions `json:"create,omitempty"` // only used during creation, never returned
	DNS             *GuestDNS         `json:"dns,omitempty"`
	Description     *string           `json:"description,omitempty"`
	Features        *LxcFeatures      `json:"features,omitempty"`
	ID              *GuestID          `json:"id"`               // only used during creation
	Memory          *LxcMemory        `json:"memory,omitempty"` // Never nil when returned
	Name            *GuestName        `json:"name,omitempty"`
	Networks        LxcNetworks       `json:"networks,omitempty"`
	Node            *NodeName         `json:"node,omitempty"` // only used during creation
	OperatingSystem OperatingSystem   `json:"os"`             // only returned
	Pool            *PoolName         `json:"pool,omitempty"`
	Privileged      *bool             `json:"privileged,omitempty"` // only used during creation
	Swap            *LxcSwap          `json:"swap,omitempty"`
	Tags            *Tags             `json:"tags,omitempty"`
}

const (
	ConfigLXC_Error_BootMountMissing     = "boot mount is required during creation"
	ConfigLXC_Error_CreateOptionsMissing = "create options are required during creation"
	ConfigLXC_Error_NoSettingsSpecified  = "no settings specified"
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
	url := "/nodes/" + node.String() + "/lxc"
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
		vmType: vmRefLXC,
	}, nil
}

func (config ConfigLXC) mapToApiCreate() (map[string]any, PoolName) {
	params := config.mapToApiShared()
	var pool PoolName
	if config.BootMount != nil {
		params[lxcApiKeyRootFS] = config.BootMount.mapToApiCreate()
	}
	if config.CPU != nil {
		config.CPU.mapToApiCreate(params)
	}
	if config.CreateOptions != nil {
		config.CreateOptions.mapToAPI(params)
	}
	if config.Description != nil && *config.Description != "" {
		params[lxcApiKeyDescription] = *config.Description
	}
	if config.DNS != nil {
		config.DNS.mapToApiCreate(params)
	}
	if config.Features != nil {
		config.Features.mapToApiCreate(params)
	}
	if config.ID != nil {
		params[lxcApiKeyGuestID] = *config.ID
	}
	if config.Memory != nil {
		params[lxcApiKeyMemory] = *config.Memory
	}
	if config.Name != nil {
		params[lxcApiKeyName] = (*config.Name).String()
	}
	if config.Networks != nil {
		config.Networks.mapToApiCreate(params)
	}
	if config.Pool != nil {
		pool = *config.Pool
		params[lxcApiKeyPool] = string(pool)
	}
	if config.Privileged != nil && !*config.Privileged {
		params[lxcApiKeyUnprivileged] = 1
	}
	if config.Swap != nil {
		params[lxcApiKeySwap] = int(*config.Swap)
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
	if config.BootMount != nil && current.BootMount != nil {
		config.BootMount.mapToApiUpdate(*current.BootMount, params)
	}
	if config.CPU != nil {
		if current.CPU != nil {
			delete += config.CPU.mapToApiUpdate(*current.CPU, params)
		} else {
			config.CPU.mapToApiCreate(params)
		}
	}
	if config.Description != nil && (current.Description == nil || *config.Description != *current.Description) {
		if *config.Description == "" {
			delete += "," + lxcApiKeyDescription
		} else {
			params[lxcApiKeyDescription] = *config.Description
		}
	}
	if config.DNS != nil {
		if current.DNS != nil {
			delete += config.DNS.mapToApiUpdate(*current.DNS, params)
		} else {
			config.DNS.mapToApiCreate(params)
		}
	}
	if config.Features != nil {
		if current.Features != nil {
			delete += config.Features.mapToApiUpdate(*current.Features, params)
		} else {
			config.Features.mapToApiCreate(params)
		}
	}
	if config.Memory != nil && (current.Memory == nil || *config.Memory != *current.Memory) {
		params[lxcApiKeyMemory] = *config.Memory
	}
	if config.Name != nil && (current.Name == nil || *config.Name != *current.Name) {
		params[lxcApiKeyName] = (*config.Name).String()
	}
	if config.Networks != nil {
		if len(current.Networks) > 0 {
			delete += config.Networks.mapToApiUpdate(current.Networks, params)
		} else {
			config.Networks.mapToApiCreate(params)
		}
	}
	if config.Swap != nil && (current.Swap == nil || *config.Swap != *current.Swap) {
		params[lxcApiKeySwap] = int(*config.Swap)
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

func (config ConfigLXC) Update(ctx context.Context, vmr *VmRef, c *Client) error {
	raw, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	current := raw.ALL(*vmr)
	if err := config.Validate(current); err != nil {
		return err
	}
	return config.updateNoCheck(ctx, vmr, current, c)
}

func (config ConfigLXC) UpdateNoCheck(ctx context.Context, vmr *VmRef, c *Client) error {
	raw, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	return config.updateNoCheck(ctx, vmr, raw.ALL(*vmr), c)
}

func (config ConfigLXC) updateNoCheck(ctx context.Context, vmr *VmRef, current *ConfigLXC, c *Client) error {
	params := config.mapToApiUpdate(*current)
	if len(params) == 0 {
		return errors.New(ConfigLXC_Error_NoSettingsSpecified)
	}
	// TODO add disk migration code here
	return c.Put(ctx, params, "/nodes/"+vmr.node.String()+"/lxc/"+vmr.vmId.String()+"/config")
}

func (config ConfigLXC) Validate(current *ConfigLXC) (err error) {
	if current != nil { // Update
		err = config.validateUpdate(*current)
	} else { // Create
		err = config.validateCreate()
	}
	if err != nil {
		return
	}
	if config.CPU != nil {
		if err = config.CPU.Validate(); err != nil {
			return
		}
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
	if config.BootMount == nil {
		return errors.New(ConfigLXC_Error_BootMountMissing)
	}
	if err = config.BootMount.Validate(nil); err != nil {
		return
	}
	if config.CreateOptions == nil {
		return errors.New(ConfigLXC_Error_CreateOptionsMissing)
	}
	if err = config.CreateOptions.Validate(); err != nil {
		return
	}
	return config.Networks.Validate(nil)
}

func (config ConfigLXC) validateUpdate(current ConfigLXC) (err error) {
	if err = config.BootMount.Validate(current.BootMount); err != nil {
		return
	}
	return config.Networks.Validate(current.Networks)
}

type RawConfigLXC map[string]any

func (raw RawConfigLXC) ALL(vmr VmRef) *ConfigLXC {
	config := ConfigLXC{
		Architecture:    raw.Architecture(),
		BootMount:       raw.BootMount(),
		CPU:             raw.CPU(),
		DNS:             raw.DNS(),
		Description:     raw.Description(),
		Features:        raw.Features(),
		ID:              util.Pointer(vmr.vmId),
		Memory:          raw.Memory(),
		Name:            raw.Name(),
		Networks:        raw.Networks(),
		Node:            util.Pointer(vmr.node),
		OperatingSystem: raw.OperatingSystem(),
		Privileged:      raw.Privileged(),
		Swap:            raw.Swap(),
		Tags:            raw.Tags()}
	if vmr.pool != "" {
		config.Pool = util.Pointer(PoolName(vmr.pool))
	}
	return &config
}

func (raw RawConfigLXC) Architecture() CpuArchitecture {
	if v, isSet := raw[lxcApiKeyArchitecture]; isSet {
		return CpuArchitecture(v.(string))
	}
	return ""
}

func (raw RawConfigLXC) Description() *string {
	if v, isSet := raw[lxcApiKeyDescription]; isSet {
		return util.Pointer(v.(string))
	}
	return nil
}

func (raw RawConfigLXC) DNS() *GuestDNS {
	return GuestDNS{}.mapToSDK(raw)
}

func (raw RawConfigLXC) Memory() *LxcMemory {
	var memory LxcMemory
	if v, isSet := raw[lxcApiKeyMemory]; isSet {
		memory = LxcMemory(v.(float64))
	}
	return &memory
}

func (raw RawConfigLXC) Name() *GuestName {
	if v, isSet := raw[lxcApiKeyName]; isSet {
		return util.Pointer(GuestName(v.(string)))
	}
	return nil
}

func (raw RawConfigLXC) OperatingSystem() OperatingSystem {
	if v, isSet := raw[lxcApiKeyOperatingSystem]; isSet {
		return OperatingSystem(v.(string))
	}
	return ""
}

// Privileged returns true if the container is privileged, false if it is unprivileged.
// Pointer is never nil.
func (raw RawConfigLXC) Privileged() *bool {
	if v, isSet := raw[lxcApiKeyUnprivileged]; isSet {
		return util.Pointer(v.(float64) == 0)
	}
	return util.Pointer(false)
}

func (raw RawConfigLXC) Swap() *LxcSwap {
	if v, isSet := raw[lxcApiKeySwap]; isSet {
		return util.Pointer(LxcSwap(v.(float64)))
	}
	return nil
}

func (raw RawConfigLXC) Tags() *Tags {
	if v, isSet := raw[lxcApiKeyTags]; isSet {
		return util.Pointer(Tags{}.mapToSDK(v.(string)))
	}
	return nil
}

const (
	lxcApiKeyArchitecture    string = "arch"
	lxcApiKeyCores           string = "cores"
	lxcApiKeyCpuLimit        string = "cpulimit"
	lxcApiKeyCpuUnits        string = "cpuunits"
	lxcApiKeyDelete          string = "delete"
	lxcApiKeyDescription     string = "description"
	lxcApiKeyFeatures        string = "features"
	lxcApiKeyGuestID         string = "vmid"
	lxcApiKeyMemory          string = "memory"
	lxcApiKeyName            string = "name"
	lxcApiKeyOperatingSystem string = "ostype"
	lxcApiKeyOsTemplate      string = "ostemplate"
	lxcApiKeyPassword        string = "password"
	lxcApiKeyPool            string = "pool"
	lxcApiKeyRootFS          string = "rootfs"
	lxcApiKeySSHPublicKeys   string = "ssh-public-keys"
	lxcApiKeySwap            string = "swap"
	lxcApiKeyTags            string = "tags"
	lxcApiKeyUnprivileged    string = "unprivileged"
	lxcPrefixApiKeyNetwork   string = "net"
)

// These settings are only available during creation and can not be changed afterwards, or returned by the API
type LxcCreateOptions struct {
	OsTemplate    *LxcTemplate    `json:"os_template,omitempty"`
	UserPassword  *string         `json:"password,omitempty"`
	PublicSSHkeys []AuthorizedKey `json:"sshkeys,omitempty"`
}

const (
	LxcCreateOptions_Error_TemplateMissing = "os template is required during creation"
)

func (config LxcCreateOptions) mapToAPI(params map[string]any) {
	if config.OsTemplate != nil {
		params[lxcApiKeyOsTemplate] = config.OsTemplate.String()
	}
	if config.UserPassword != nil {
		params[lxcApiKeyPassword] = *config.UserPassword
	}
	if len(config.PublicSSHkeys) != 0 {
		params[lxcApiKeySSHPublicKeys] = sshKeyUrlEncode(config.PublicSSHkeys)
	}
}

func (config LxcCreateOptions) Validate() error {
	if config.OsTemplate == nil {
		return errors.New(LxcCreateOptions_Error_TemplateMissing)
	}
	if err := config.OsTemplate.Validate(); err != nil {
		return err
	}
	return nil
}

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

type LxcTemplate struct {
	Storage string `json:"storage"`
	File    string `json:"template"`
}

const (
	LxcTemplate_Error_StorageMissing = "storage is required"
	LxcTemplate_Error_FileMissing    = "file is required"
)

func (template LxcTemplate) String() string { // String is for fmt.Stringer.
	return template.Storage + ":vztmpl/" + strings.TrimPrefix(template.File, "/")
}

func (template LxcTemplate) Validate() error {
	if template.Storage == "" {
		return errors.New(LxcTemplate_Error_StorageMissing)
	}
	if template.File == "" {
		return errors.New(LxcTemplate_Error_FileMissing)
	}
	return nil
}

type LxcSwap uint

func (swap LxcSwap) String() string { return strconv.Itoa(int(swap)) } // String is for fmt.Stringer.

func NewConfigLXCFromApi(ctx context.Context, vmr *VmRef, c *Client) (RawConfigLXC, error) {
	rawConfig, err := c.GetVmConfig(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return rawConfig, nil
}
