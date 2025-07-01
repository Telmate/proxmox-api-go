package proxmox

import (
	"context"
	"crypto/sha1"
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
	Digest          [sha1.Size]byte   `json:"digest,omitempty"` // only returned.
	Features        *LxcFeatures      `json:"features,omitempty"`
	ID              *GuestID          `json:"id"`               // only used during creation
	Memory          *LxcMemory        `json:"memory,omitempty"` // Never nil when returned
	Name            *GuestName        `json:"name,omitempty"`   // Never nil when returned
	Networks        LxcNetworks       `json:"networks,omitempty"`
	Node            *NodeName         `json:"node,omitempty"` // only used during creation
	OperatingSystem OperatingSystem   `json:"os"`             // only returned
	Pool            *PoolName         `json:"pool,omitempty"`
	Privileged      *bool             `json:"privileged,omitempty"` // only used during creation, defaults to false ,never nil when returned
	State           *PowerState       `json:"state,omitempty"`
	Swap            *LxcSwap          `json:"swap,omitempty"` // Never nil when returned
	Tags            *Tags             `json:"tags,omitempty"`
	rawDigest       digest            `json:"-"`
}

const (
	lxcDefaultPrivilege bool = false
)

const (
	ConfigLXC_Error_BootMountMissing     = "boot mount is required during creation"
	ConfigLXC_Error_CreateOptionsMissing = "create options are required during creation"
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

	vmRef := &VmRef{
		node:   node,
		vmId:   id,
		pool:   pool,
		vmType: vmRefLXC}
	if config.PowerState != nil && *config.PowerState == PowerStateRunning {
		if err := GuestStart(ctx, vmRef, c); err != nil {
			return nil, err
		}
	}

	return vmRef, nil
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
	if config.Privileged == nil || !*config.Privileged {
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
	params[lxcApiKeyDigest] = current.rawDigest.String()
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

func (config ConfigLXC) Update(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	rawStatus, err := vmr.GetRawGuestStatus(ctx, c)
	if err != nil {
		return err
	}
	raw, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	current := raw.all(*vmr)
	if err := config.Validate(current); err != nil {
		return err
	}
	return config.updateNoCheck(ctx, allowRestart, vmr, current, rawStatus.State(), c)
}

func (config ConfigLXC) UpdateNoCheck(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	rawStatus, err := vmr.GetRawGuestStatus(ctx, c)
	if err != nil {
		return err
	}
	raw, err := NewConfigLXCFromApi(ctx, vmr, c)
	if err != nil {
		return err
	}
	return config.updateNoCheck(ctx, allowRestart, vmr, raw.all(*vmr), rawStatus.State(), c)
}

func (config ConfigLXC) updateNoCheck(
	ctx context.Context,
	allowRestart bool,
	vmr *VmRef,
	current *ConfigLXC,
	currentState PowerState,
	c *Client) error {

	var getRootMount, getMounts bool

	url := "/nodes/" + vmr.node.String() + "/lxc/" + vmr.vmId.String()

	targetState := config.State.combine(currentState)

	if targetState == PowerStateStopped && currentState != PowerStateStopped { // We want the vm to be stopped, better to do this before we start making other api calls
		if !allowRestart {
			return errors.New("guest has to be stopped before applying changes")
		}
		if _, err := c.ShutdownVm(ctx, vmr); err != nil {
			return err
		}
		currentState = PowerStateStopped // We assume the guest is stopped now
	}

	if params := config.mapToApiUpdate(*current); len(params) > 0 {
		if err := c.Put(ctx, params, url+"/config"); err != nil {
			return err
		}
		if currentState == PowerStateRunning || currentState == PowerStateUnknown { // If the gest is running, we have to check if it has pending changes
			pendingChanges, err := GuestHasPendingChanges(ctx, vmr, c)
			if err != nil {
				return fmt.Errorf("error checking for pending changes: %w", err)
			}
			if pendingChanges {
				if !allowRestart {
					// TODO revert pending changes
					return errors.New("guest has to be restarted to apply changes")
				}
				if err := GuestReboot(ctx, vmr, c); err != nil {
					return fmt.Errorf("error restarting guest: %w", err)
				}
				currentState = PowerStateRunning // We assume the guest is running now
			}
		}
	}
	if currentState != PowerStateRunning && targetState == PowerStateRunning { // We want the guest to be running, so we start it now
		if err := GuestStart(ctx, vmr, c); err != nil {
			return err
		}
	}
	return nil
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
	if config.Pool != nil && config.Pool.String() != "" {
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
	if config.Features != nil {
		privilege := lxcDefaultPrivilege
		if config.Privileged != nil {
			privilege = *config.Privileged
		}
		if err = config.Features.Validate(privilege); err != nil {
			return
		}
	}
	return config.Networks.Validate(nil)
}

func (config ConfigLXC) validateUpdate(current ConfigLXC) (err error) {
	if config.BootMount != nil {
		if err = config.BootMount.Validate(current.BootMount); err != nil {
			return
		}
	}
	if config.Features != nil {
		if err = config.Features.Validate(*current.Privileged); err != nil {
			return
		}
	}
	return config.Networks.Validate(current.Networks)
}

type RawConfigLXC map[string]any

func (raw RawConfigLXC) ALL(vmr VmRef, state PowerState) *ConfigLXC {
	config := raw.all(vmr)
	config.Digest = config.rawDigest.sha1()
	if state != PowerStateUnknown {
		config.State = &state
	}
	return config
}

func (raw RawConfigLXC) all(vmr VmRef) *ConfigLXC {
	privileged := raw.isPrivileged()
	config := ConfigLXC{
		Architecture:    raw.Architecture(),
		BootMount:       raw.BootMount(),
		CPU:             raw.CPU(),
		DNS:             raw.DNS(),
		Description:     raw.Description(),
		Features:        raw.features(privileged),
		ID:              util.Pointer(vmr.vmId),
		Memory:          util.Pointer(raw.Memory()),
		Name:            util.Pointer(raw.Name()),
		Networks:        raw.Networks(),
		Node:            util.Pointer(vmr.node),
		OperatingSystem: raw.OperatingSystem(),
		Privileged:      &privileged,
		Swap:            util.Pointer(raw.Swap()),
		Tags:            raw.Tags(),
		rawDigest:       raw.digest()}
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

func (raw RawConfigLXC) Digest() [sha1.Size]byte {
	return raw.digest().sha1()
}

func (raw RawConfigLXC) digest() digest {
	if v, isSet := raw[lxcApiKeyDigest]; isSet {
		return digest(v.(string))
	}
	return ""
}

func (raw RawConfigLXC) DNS() *GuestDNS {
	return GuestDNS{}.mapToSDK(raw)
}

func (raw RawConfigLXC) Memory() LxcMemory {
	if v, isSet := raw[lxcApiKeyMemory]; isSet {
		return LxcMemory(v.(float64))
	}
	return 0
}

func (raw RawConfigLXC) Name() GuestName {
	if v, isSet := raw[lxcApiKeyName]; isSet {
		return GuestName(v.(string))
	}
	return ""
}

func (raw RawConfigLXC) OperatingSystem() OperatingSystem {
	if v, isSet := raw[lxcApiKeyOperatingSystem]; isSet {
		return OperatingSystem(v.(string))
	}
	return ""
}

// Privileged returns true if the container is privileged, false if it is unprivileged.
func (raw RawConfigLXC) Privileged() bool {
	return raw.isPrivileged()
}

func (raw RawConfigLXC) isPrivileged() bool {
	if v, isSet := raw[lxcApiKeyUnprivileged]; isSet {
		return v.(float64) == 0
	}
	return true // when privileged the API does not return the key at all, so we assume it is privileged
}

func (raw RawConfigLXC) Swap() LxcSwap {
	if v, isSet := raw[lxcApiKeySwap]; isSet {
		return LxcSwap(v.(float64))
	}
	return 0
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
	lxcApiKeyDigest          string = "digest"
	lxcApiKeyFeatures        string = "features"
	lxcApiKeyGuestID         string = "vmid"
	lxcApiKeyMemory          string = "memory"
	lxcApiKeyName            string = "hostname"
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
