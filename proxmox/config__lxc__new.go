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

func (arch CpuArchitecture) String() string { return string(arch) } // String is for fmt.Stringer.

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
	Mounts          LxcMounts         `json:"mounts,omitempty"`
	Name            *GuestName        `json:"name,omitempty"` // Never nil when returned
	Networks        LxcNetworks       `json:"networks,omitempty"`
	Node            *NodeName         `json:"node,omitempty"` // only used during creation
	OperatingSystem OperatingSystem   `json:"os"`             // only returned
	Pool            *PoolName         `json:"pool,omitempty"`
	Privileged      *bool             `json:"privileged,omitempty"` // only used during creation, defaults to false ,never nil when returned
	Protection      *bool             `json:"protection,omitempty"` // Never nil when returned
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
		id, err = guestCreateLoop_Unsafe(ctx, "vmid", url, params, c)
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
		vmType: GuestLxc}
	if config.State != nil && *config.State == PowerStateRunning {
		if err := GuestStart(ctx, vmRef, c); err != nil {
			return nil, err
		}
	}

	return vmRef, nil
}

func (config ConfigLXC) mapToApiCreate() (map[string]any, PoolName) {
	params := config.mapToApiShared()
	privileged := true
	if config.Privileged == nil || !*config.Privileged {
		params[lxcApiKeyUnprivileged] = 1
		privileged = false
	}
	var pool PoolName
	if config.BootMount != nil {
		params[lxcApiKeyRootFS] = config.BootMount.mapToApiCreate(privileged)
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
	if config.Mounts != nil {
		config.Mounts.mapToAPICreate(privileged, params)
	}
	if config.Networks != nil {
		config.Networks.mapToApiCreate(params)
	}
	if config.Pool != nil {
		pool = *config.Pool
		params[lxcApiKeyPool] = string(pool)
	}
	if config.Protection != nil && *config.Protection {
		params[lxcAPIKeyProtection] = "1"
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
	privileged := lxcDefaultPrivilege
	if current.Privileged != nil {
		privileged = *current.Privileged
	}
	params := config.mapToApiShared()
	var delete string
	if config.BootMount != nil && current.BootMount != nil {
		config.BootMount.mapToApiUpdate(*current.BootMount, privileged, params)
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
	if len(config.Mounts) > 0 {
		if len(current.Mounts) > 0 {
			delete += config.Mounts.mapToAPIUpdate(current.Mounts, privileged, params)
		} else {
			config.Mounts.mapToAPICreate(privileged, params)
		}
	}
	if len(config.Networks) > 0 {
		if len(current.Networks) > 0 {
			delete += config.Networks.mapToApiUpdate(current.Networks, params)
		} else {
			config.Networks.mapToApiCreate(params)
		}
	}
	if config.Protection != nil && (current.Protection == nil || *config.Protection != *current.Protection) {
		if *config.Protection {
			params[lxcAPIKeyProtection] = "1"
		} else {
			delete += "," + lxcAPIKeyProtection
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
	if len(params) > 0 {
		params[lxcApiKeyDigest] = current.rawDigest.String()
	}
	return params
}

func (config ConfigLXC) Update(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	rawStatus, err := vmr.getRawGuestStatus_Unsafe(ctx, c)
	if err != nil {
		return err
	}

	raw, err := guestGetLxcRawConfig_Unsafe(ctx, vmr, c.new().apiGet())
	if err != nil {
		return err
	}
	current := raw.get(*vmr)
	if err := config.Validate(current); err != nil {
		return err
	}
	version, err := c.Version(ctx)
	if err != nil {
		return err
	}
	return config.update_Unsafe(ctx, allowRestart, vmr, current, rawStatus.GetState(), version.Encode(), c)
}

func (config ConfigLXC) UpdateNoCheck(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	rawStatus, err := vmr.getRawGuestStatus_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	raw, err := c.new().guestGetLxcRawConfig(ctx, vmr)
	if err != nil {
		return err
	}
	version, err := c.Version(ctx)
	if err != nil {
		return err
	}
	return config.update_Unsafe(ctx, allowRestart, vmr, raw.get(*vmr), rawStatus.GetState(), version.Encode(), c)
}

func (config ConfigLXC) update_Unsafe(
	ctx context.Context,
	allowRestart bool,
	vmr *VmRef,
	current *ConfigLXC,
	currentState PowerState,
	version EncodedVersion,
	c *Client) error {

	ca := c.new().apiGet()

	var move []lxcMountMove
	var resize []lxcMountResize
	var getRootMount, getMounts, requiresOffStateForMountActions bool

	url := "/nodes/" + vmr.node.String() + "/lxc/" + vmr.vmId.String()

	targetState := config.State.combine(currentState)

	// Check if we have to move or resize any mounts
	if config.BootMount != nil && current.BootMount != nil {
		getRootMount = true
		markedMounts := config.BootMount.markMountChanges_Unsafe(current.BootMount)
		move = markedMounts.move
		resize = markedMounts.resize
	}
	if config.Mounts != nil && current.Mounts != nil {
		getMounts = true
		markedMounts := config.Mounts.markMountChanges(current.Mounts)
		move = append(move, markedMounts.move...)
		resize = append(resize, markedMounts.resize...)
		requiresOffStateForMountActions = markedMounts.offState
	}

	var err error
	if targetState == PowerStateStopped && currentState != PowerStateStopped { // We want the vm to be stopped, better to do this before we start making other api calls
		if !allowRestart {
			return errors.New("guest has to be stopped before applying changes")
		}
		if _, err = c.ShutdownVm(ctx, vmr); err != nil {
			return err
		}
		currentState = PowerStateStopped // We assume the guest is stopped now
	}

	if requiresOffStateForMountActions || len(move) > 0 { // turn the guest off
		if currentState == PowerStateRunning || currentState == PowerStateUnknown { // Stop guest before moving disks
			if !allowRestart {
				return errors.New("guest has to be stopped before moving disks")
			}
			if err = GuestShutdown(ctx, vmr, c, true); err != nil { // We have to stop the guest before moving disks
				return err
			}
			currentState = PowerStateStopped // We assume the guest is stopped now
		}
	}

	if len(resize) > 0 || len(move) > 0 {

		for i := range move { // Move mounts
			if _, err = move[i].move(ctx, true, vmr, c); err != nil {
				return err
			}
		}

		for i := range resize { // Resize mounts
			if _, err = resize[i].resize(ctx, vmr, c); err != nil {
				return err
			}
		}

		var newCurrent RawConfigLXC
		newCurrent, err = c.new().guestGetLxcRawConfig(ctx, vmr) // We have to refetch part of the current config
		if err != nil {
			return err
		}
		current.rawDigest = newCurrent.getDigest()

		if len(move) > 0 {
			if getRootMount {
				current.BootMount = newCurrent.GetBootMount()
			}
			if getMounts {
				current.Mounts = newCurrent.GetMounts()
			}
		}
	}

	if params := config.mapToApiUpdate(*current); len(params) > 0 {
		if err = c.Put(ctx, params, url+"/config"); err != nil {
			return err
		}
		if currentState == PowerStateRunning || currentState == PowerStateUnknown { // If the guest is running, we have to check if it has pending changes
			var pendingChanges bool
			pendingChanges, err = vmr.pendingChanges(ctx, ca)
			if err != nil {
				return fmt.Errorf("error checking for pending changes: %w", err)
			}
			if pendingChanges {
				if !allowRestart {
					// TODO revert pending changes
					return errors.New("guest has to be restarted to apply changes")
				}
				if err = GuestReboot(ctx, vmr, c); err != nil {
					return fmt.Errorf("error restarting guest: %w", err)
				}
				currentState = PowerStateRunning // We assume the guest is running now
			}
		}
	}
	if currentState != PowerStateRunning && targetState == PowerStateRunning { // We want the guest to be running, so we start it now
		if err = GuestStart(ctx, vmr, c); err != nil {
			return err
		}
	}
	if config.Pool != nil {
		err = guestSetPoolNoCheck(ctx, c, vmr.vmId, *config.Pool, current.Pool, version)
	}
	return err
}

func (config ConfigLXC) Validate(current *ConfigLXC) (err error) {
	if current != nil { // Update
		err = config.validateUpdate(*current)
	} else { // Create
		privileged := lxcDefaultPrivilege
		if config.Privileged != nil {
			privileged = *config.Privileged
		}
		err = config.validateCreate(privileged)
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

func (config ConfigLXC) validateCreate(privileged bool) (err error) {
	if config.BootMount == nil {
		return errors.New(ConfigLXC_Error_BootMountMissing)
	}
	if err = config.BootMount.Validate(nil, privileged); err != nil {
		return
	}
	if config.CreateOptions == nil {
		return errors.New(ConfigLXC_Error_CreateOptionsMissing)
	}
	if err = config.CreateOptions.Validate(); err != nil {
		return
	}
	privilege := lxcDefaultPrivilege
	if config.Privileged != nil {
		privilege = *config.Privileged
	}
	if config.Features != nil {
		if err = config.Features.Validate(privilege); err != nil {
			return
		}
	}
	if config.Mounts != nil {
		if err = config.Mounts.validateCreate(privileged); err != nil {
			return
		}
	}
	return config.Networks.Validate(nil)
}

func (config ConfigLXC) validateUpdate(current ConfigLXC) (err error) {
	privileged := lxcDefaultPrivilege
	if current.Privileged != nil {
		privileged = *current.Privileged
	}
	if config.BootMount != nil {
		if err = config.BootMount.Validate(current.BootMount, privileged); err != nil {
			return
		}
	}
	if config.Features != nil {
		if err = config.Features.Validate(*current.Privileged); err != nil {
			return
		}
	}
	if config.Mounts != nil {
		if current.Mounts != nil {
			if err := config.Mounts.validateUpdate(current.Mounts, privileged); err != nil {
				return err
			}
		} else {
			if err := config.Mounts.validateCreate(privileged); err != nil {
				return err
			}
		}
	}
	return config.Networks.Validate(current.Networks)
}

type RawConfigLXC interface {
	Get(vmr VmRef, state PowerState) *ConfigLXC
	GetArchitecture() CpuArchitecture
	GetBootMount() *LxcBootMount
	GetDNS() *GuestDNS
	GetDescription() *string
	GetDigest() [sha1.Size]byte
	GetMemory() LxcMemory
	GetMounts() LxcMounts
	GetName() GuestName
	GetOperatingSystem() OperatingSystem
	GetPrivileged() bool
	GetProtection() bool
	GetSwap() LxcSwap
	GetTags() *Tags
	get(vmr VmRef) *ConfigLXC
	getBootMount(privileged bool) *LxcBootMount
	getDigest() digest
	getMounts(privileged bool) LxcMounts
}

type rawConfigLXC struct{ a map[string]any }

func (raw *rawConfigLXC) Get(vmr VmRef, state PowerState) *ConfigLXC {
	config := raw.get(vmr)
	config.Digest = config.rawDigest.sha1()
	if state != PowerStateUnknown {
		config.State = &state
	}
	return config
}

func (raw *rawConfigLXC) get(vmr VmRef) *ConfigLXC {
	privileged := raw.isPrivileged()
	config := ConfigLXC{
		Architecture:    raw.GetArchitecture(),
		BootMount:       raw.getBootMount(privileged),
		CPU:             raw.GetCPU(),
		DNS:             raw.GetDNS(),
		Description:     raw.GetDescription(),
		Features:        raw.getFeatures(privileged),
		ID:              util.Pointer(vmr.vmId),
		Memory:          util.Pointer(raw.GetMemory()),
		Mounts:          raw.getMounts(privileged),
		Name:            util.Pointer(raw.GetName()),
		Networks:        raw.GetNetworks(),
		Node:            util.Pointer(vmr.node),
		OperatingSystem: raw.GetOperatingSystem(),
		Privileged:      &privileged,
		Protection:      util.Pointer(raw.GetProtection()),
		Swap:            util.Pointer(raw.GetSwap()),
		Tags:            raw.GetTags(),
		rawDigest:       raw.getDigest()}
	if vmr.pool != "" {
		config.Pool = util.Pointer(PoolName(vmr.pool))
	}
	return &config
}

func (raw *rawConfigLXC) GetArchitecture() CpuArchitecture {
	if v, isSet := raw.a[lxcApiKeyArchitecture]; isSet {
		return CpuArchitecture(v.(string))
	}
	return ""
}

func (raw *rawConfigLXC) GetDescription() *string {
	if v, isSet := raw.a[lxcApiKeyDescription]; isSet {
		return util.Pointer(v.(string))
	}
	return nil
}

func (raw *rawConfigLXC) GetDigest() [sha1.Size]byte {
	return raw.getDigest().sha1()
}

func (raw *rawConfigLXC) getDigest() digest {
	if v, isSet := raw.a[lxcApiKeyDigest]; isSet {
		return digest(v.(string))
	}
	return ""
}

func (raw *rawConfigLXC) GetDNS() *GuestDNS {
	return GuestDNS{}.mapToSDK(raw.a)
}

func (raw *rawConfigLXC) GetMemory() LxcMemory {
	if v, isSet := raw.a[lxcApiKeyMemory]; isSet {
		return LxcMemory(v.(float64))
	}
	return 0
}

func (raw *rawConfigLXC) GetName() GuestName {
	if v, isSet := raw.a[lxcApiKeyName]; isSet {
		return GuestName(v.(string))
	}
	return ""
}

func (raw *rawConfigLXC) GetOperatingSystem() OperatingSystem {
	if v, isSet := raw.a[lxcApiKeyOperatingSystem]; isSet {
		return OperatingSystem(v.(string))
	}
	return ""
}

// GetPrivileged returns true if the container is privileged, false if it is unprivileged.
func (raw *rawConfigLXC) GetPrivileged() bool {
	return raw.isPrivileged()
}

func (raw *rawConfigLXC) isPrivileged() bool {
	if v, isSet := raw.a[lxcApiKeyUnprivileged]; isSet {
		return v.(float64) == 0
	}
	return true // when privileged the API does not return the key at all, so we assume it is privileged
}

func (raw *rawConfigLXC) GetProtection() bool {
	if v, isSet := raw.a[lxcAPIKeyProtection]; isSet {
		return int(v.(float64)) == 1
	}
	return false
}

func (raw *rawConfigLXC) GetSwap() LxcSwap {
	if v, isSet := raw.a[lxcApiKeySwap]; isSet {
		return LxcSwap(v.(float64))
	}
	return 0
}

func (raw *rawConfigLXC) GetTags() *Tags {
	if v, isSet := raw.a[lxcApiKeyTags]; isSet {
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
	lxcAPIKeyProtection      string = "protection"
	lxcApiKeyRootFS          string = "rootfs"
	lxcApiKeySSHPublicKeys   string = "ssh-public-keys"
	lxcApiKeySwap            string = "swap"
	lxcApiKeyTags            string = "tags"
	lxcApiKeyUnprivileged    string = "unprivileged"
	lxcPrefixApiKeyMount     string = "mp"
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

// NewRawConfigLXCFromAPI returns the configuration of the LXC guest.
// Including pending changes.
func NewRawConfigLXCFromAPI(ctx context.Context, vmr *VmRef, c *Client) (RawConfigLXC, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	if c == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	return c.new().guestGetLxcRawConfig(ctx, vmr)
}

func guestGetLxcRawConfig_Unsafe(ctx context.Context, vmr *VmRef, c clientApiInterface) (RawConfigLXC, error) {
	rawConfig, err := c.getGuestConfig(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return &rawConfigLXC{a: rawConfig}, nil
}

func (c *clientNew) guestGetLxcRawConfig(ctx context.Context, vmr *VmRef) (RawConfigLXC, error) {
	return guestGetLxcRawConfig_Unsafe(ctx, vmr, c.api)
}

// NewActiveRawConfigLXCFromApi returns the active configuration of the LXC guest.
// Without pending changes.
func NewActiveRawConfigLXCFromApi(ctx context.Context, vmr *VmRef, c *Client) (raw RawConfigLXC, pending bool, err error) {
	return c.new().guestGetLxcActiveRawConfig(ctx, vmr)
}

func guestGetActiveRawLxcConfig_Unsafe(ctx context.Context, vmr *VmRef, c clientApiInterface) (raw RawConfigLXC, pending bool, err error) {
	var tmpConfig map[string]any
	tmpConfig, pending, err = vmr.pendingActiveConfig(ctx, c)
	if err != nil {
		return nil, false, err
	}
	return &rawConfigLXC{a: tmpConfig}, pending, nil
}

func (c *clientNew) guestGetLxcActiveRawConfig(ctx context.Context, vmr *VmRef) (raw RawConfigLXC, pending bool, err error) {
	return guestGetActiveRawLxcConfig_Unsafe(ctx, vmr, c.api)
}
