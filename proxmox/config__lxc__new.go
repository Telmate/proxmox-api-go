package proxmox

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	LxcGuestInterface interface {
		Clone(context.Context, VmRef, CloneLxcTarget) (*VmRef, error)
		CloneNoCheck(context.Context, VmRef, CloneLxcTarget) (*VmRef, error)

		// VmRef can be set while there was an error, to indicate a guest was created.
		Create(context.Context, ConfigLXC) (*VmRef, error)
		CreateNoCheck(context.Context, ConfigLXC) (*VmRef, error)

		Migrate(context.Context, VmRef, NodeName) error
		MigrateNoCheck(context.Context, VmRef, NodeName) error

		// Read returns the configuration of the LXC guest.
		// Including pending changes.
		Read(context.Context, VmRef) (RawConfigLXC, error)
		ReadNoCheck(context.Context, VmRef) (RawConfigLXC, error)

		// ReadActive returns the active/running configuration of the Lxc guest.
		// Without pending changes.
		ReadActive(context.Context, VmRef) (raw RawConfigLXC, pending bool, err error)
		ReadActiveNoCheck(context.Context, VmRef) (raw RawConfigLXC, pending bool, err error)

		// Only available from PVE 9.1 and onwards.
		ReadNetworkInterfaceInfo(context.Context, VmRef) (RawLxcInfoNetworkInterfaces, error)
		ReadNetworkInterfaceInfoNoCheck(context.Context, VmRef) (RawLxcInfoNetworkInterfaces, error)

		// When allowRestart is false an error is return if the update requires a reboot or shutdown.
		Update(ctx context.Context, vmr VmRef, allowRestart bool, config ConfigLXC) error
		UpdateNoCheck(ctx context.Context, vmr VmRef, allowRestart bool, config ConfigLXC) error
	}

	lxcGuestClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ LxcGuestInterface = (*lxcGuestClient)(nil)

func (c *lxcGuestClient) Clone(ctx context.Context, vmr VmRef, config CloneLxcTarget) (*VmRef, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	if _, err := ref.checkQemu(ctx, c.api); err != nil {
		return nil, err
	}
	return c.CloneNoCheck(ctx, ref.legacy(), config)
}

func (c *lxcGuestClient) CloneNoCheck(ctx context.Context, vmr VmRef, config CloneLxcTarget) (*VmRef, error) {
	return LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize().cloneLxc(ctx, config, c.oldClient, c.api)
}

func (c *lxcGuestClient) Create(ctx context.Context, config ConfigLXC) (*VmRef, error) {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return nil, err
	}
	if err = config.Validate(nil, version); err != nil {
		return nil, err
	}
	return c.CreateNoCheck(ctx, config)
}

func (c *lxcGuestClient) CreateNoCheck(ctx context.Context, config ConfigLXC) (*VmRef, error) {
	return config.create(ctx, c.oldClient, c.api)
}

func (c *lxcGuestClient) Migrate(ctx context.Context, vmr VmRef, new NodeName) error {
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	ref.checkLxc(ctx, c.api)
	return c.MigrateNoCheck(ctx, ref.legacy(), new)
}

func (c *lxcGuestClient) MigrateNoCheck(ctx context.Context, vmr VmRef, new NodeName) error {
	return LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize().migrate(ctx, c.api, new, false)
}

func (c *lxcGuestClient) Read(ctx context.Context, vmr VmRef) (RawConfigLXC, error) {
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	if _, err := ref.checkLxc(ctx, c.api); err != nil {
		return nil, err
	}
	return c.ReadNoCheck(ctx, ref.legacy())
}

func (c *lxcGuestClient) ReadNoCheck(ctx context.Context, vmr VmRef) (RawConfigLXC, error) {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return nil, err
	}
	return readLxcConfig(ctx, LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize(), version, c.api)
}

func (c *lxcGuestClient) ReadActive(ctx context.Context, vmr VmRef) (RawConfigLXC, bool, error) {
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	if _, err := ref.checkLxc(ctx, c.api); err != nil {
		return nil, false, err
	}
	return c.ReadActiveNoCheck(ctx, ref.legacy())
}

func (c *lxcGuestClient) ReadActiveNoCheck(ctx context.Context, vmr VmRef) (RawConfigLXC, bool, error) {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return nil, false, err
	}
	return guestGetActiveRawLxcConfig(ctx, vmr.toGuestRef(), version, c.api)
}

func (c *lxcGuestClient) ReadNetworkInterfaceInfo(ctx context.Context, vmr VmRef) (RawLxcInfoNetworkInterfaces, error) {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return nil, err
	}
	if version.Encode() < version_9_1_0 {
		return nil, newVersionError("lxc network information", Version{Major: 9, Minor: 1})
	}
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	if _, err = ref.checkLxc(ctx, c.api); err != nil {
		return nil, err
	}
	return c.ReadNetworkInterfaceInfoNoCheck(ctx, ref.legacy())
}

func (c *lxcGuestClient) ReadNetworkInterfaceInfoNoCheck(ctx context.Context, vmr VmRef) (RawLxcInfoNetworkInterfaces, error) {
	return readLxcInterfaceStatus(ctx, vmr.toGuestRef(), c.api)
}

func (c *lxcGuestClient) Update(ctx context.Context, vmr VmRef, allowRestart bool, config ConfigLXC) error {
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	rawGuest, err := ref.checkLxc(ctx, c.api)
	if err != nil {
		return err
	}
	var state PowerState
	var pool PoolName
	if rawGuest != nil {
		state = rawGuest.GetState()
		if config.Pool != nil {
			pool = rawGuest.GetPool()
		}
	} else {
		if config.Pool != nil {
			raws, err := c.api.listGuestResources(ctx)
			if err != nil {
				return err
			}
			raw, ok := raws.selectID(ref.ID)
			if !ok {
				return errorMsg{}.guestDoesNotExist(ref.ID)
			}
			state = raw.GetState()
			pool = raw.GetPool()
		} else {
			rawStatus, err := vmr.getRawGuestStatus_Unsafe(ctx, c.api)
			if err != nil {
				return err
			}
			state = rawStatus.GetState()
		}
	}
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	raw, err := readLxcConfig(ctx, ref, version, c.api)
	if err != nil {
		return err
	}
	current := raw.get(pool)
	if err := config.Validate(current, version); err != nil {
		return err
	}
	return config.update_Unsafe(ctx, allowRestart, ref, current, state, version, c.api)
}

func (c *lxcGuestClient) UpdateNoCheck(ctx context.Context, vmr VmRef, allowRestart bool, config ConfigLXC) error {
	ref := LxcRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	var state PowerState
	var pool PoolName
	if config.Pool != nil {
		raws, err := c.api.listGuestResources(ctx)
		if err != nil {
			return err
		}
		raw, ok := raws.selectID(ref.ID)
		if !ok {
			return errorMsg{}.guestDoesNotExist(ref.ID)
		}
		state = raw.GetState()
		pool = raw.GetPool()
	} else {
		raw, err := ref.status(ctx, c.api)
		if err != nil {
			return err
		}
		state = raw.GetState()
	}
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	var raw *rawConfigLXC
	raw, err = readLxcConfig(ctx, ref, version, c.api)
	if err != nil {
		return err
	}
	return config.update_Unsafe(ctx, allowRestart, ref, raw.get(pool), state, version, c.api)
}

type LxcCpuArchitecture string

func (arch LxcCpuArchitecture) String() string { return string(arch) } // String is for fmt.Stringer.

type OperatingSystem string

type ConfigLXC struct {
	BootMount       *LxcBootMount       `json:"boot_mount,omitempty"`
	CPU             *LxcCPU             `json:"cpu,omitempty"`
	CreateOptions   *LxcCreateOptions   `json:"create,omitempty"` // only used during creation, never returned
	DNS             *GuestDNS           `json:"dns,omitempty"`
	Description     *string             `json:"description,omitempty"`
	Features        *LxcFeatures        `json:"features,omitempty"`
	ID              *GuestID            `json:"id"`               // only used during creation
	Memory          *LxcMemory          `json:"memory,omitempty"` // Never nil when returned
	Mounts          LxcMounts           `json:"mounts,omitempty"`
	Name            *GuestName          `json:"name,omitempty"` // Never nil when returned
	Networks        LxcNetworks         `json:"networks,omitempty"`
	Node            *NodeName           `json:"node,omitempty"` // only used during creation
	Pool            *PoolName           `json:"pool,omitempty"`
	Privileged      *bool               `json:"privileged,omitempty"`         // only used during creation, defaults to false ,never nil when returned
	Protection      *bool               `json:"protection,omitempty"`         // Never nil when returned
	StartAtNodeBoot *bool               `json:"start_at_node_boot,omitempty"` // Never nil when returned
	StartupShutdown *StartupAndShutdown `json:"startup_shutdown,omitempty"`
	State           *PowerState         `json:"state,omitempty"`
	Swap            *LxcSwap            `json:"swap,omitempty"` // Never nil when returned
	Tags            *Tags               `json:"tags,omitempty"` // Never nil when returned
	Architecture    LxcCpuArchitecture  `json:"architecture"`   // only returned
	OperatingSystem OperatingSystem     `json:"os"`             // only returned
	rawDigest       digest              `json:"-"`
	Digest          [sha1.Size]byte     `json:"digest,omitempty"` // only returned.
}

const (
	lxcDefaultPrivilege bool = false
)

const (
	ConfigLXC_Error_BootMountMissing     = "boot mount is required during creation"
	ConfigLXC_Error_CreateOptionsMissing = "create options are required during creation"
)

// Deprecated: LxcGuestInterface.Create() use instead.
func (config ConfigLXC) Create(ctx context.Context, c *Client) (*VmRef, error) {
	return c.New().LxcGuest.Create(ctx, config)
}

// Deprecated: LxcGuestInterface.CreateNoCheck() use instead.
func (config ConfigLXC) CreateNoCheck(ctx context.Context, c *Client) (*VmRef, error) {
	return config.create(ctx, c, c.api())
}

func (config ConfigLXC) create(ctx context.Context, c *Client, ca *clientAPI) (*VmRef, error) {
	params, body, pool := config.mapToApiCreate()

	var err error
	var id GuestID
	var node NodeName
	if config.Node != nil {
		node = *config.Node
	}
	url := "/nodes/" + node.String() + "/lxc"
	if config.ID == nil {
		id, err = guestCreateLoop_Unsafe(ctx, lxcApiKeyGuestID, url, params, body, c, ca)
		if err != nil {
			return nil, err
		}
	} else {
		if err = ca.postTask(ctx, url, combineParamsAndBody(params, body)); err != nil {
			return nil, fmt.Errorf("error creating LXC: %v", err)
		}
		id = *config.ID
	}

	vmRef := &VmRef{
		node:   node,
		vmId:   id,
		pool:   pool,
		vmType: GuestLxc}
	if config.State != nil && *config.State == PowerStateRunning {
		err = LxcRef{ID: id,Node: node}.generalize().start(ctx, ca)
		if err != nil {
			return vmRef, err
		}
	}

	return vmRef, nil
}

func (config ConfigLXC) mapToApiCreate() (map[string]any, *[]byte, PoolName) {
	builder := strings.Builder{}
	builderPtr := &builder
	params := config.mapToApiShared()
	var privileged bool
	builder.WriteString("&" + lxcApiKeyUnprivileged + "=")
	if config.Privileged != nil && *config.Privileged {
		builder.WriteString("0")
		privileged = true
	} else {
		builder.WriteString("1")
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
		params[lxcApiKeyGuestID] = int(*config.ID)
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
		config.Networks.mapToApiCreate(builderPtr)
	}
	if config.Pool != nil {
		pool = *config.Pool
		params[lxcApiKeyPool] = string(pool)
	}
	if config.Protection != nil && *config.Protection {
		params[lxcAPIKeyProtection] = "1"
	}
	if config.StartAtNodeBoot != nil {
		startAtNodeBootMapToApiCreate(params, *config.StartAtNodeBoot)
	}
	if config.StartupShutdown != nil {
		config.StartupShutdown.mapToApiCreate(params)
	}
	if config.Swap != nil {
		params[lxcApiKeySwap] = int(*config.Swap)
	}
	if config.Tags != nil {
		if v := config.Tags.mapToApiCreate(); v != "" {
			builder.WriteString("&" + qemuApiKeyTags + "=")
			builder.WriteString(v)
		}
	}
	var body *[]byte
	if builder.Len() > 0 {
		body = new(bytes.NewBufferString(builder.String()[1:]).Bytes())
	}
	if len(params) == 0 { // Temporarily clear params
		params = nil
	}
	return params, body, pool
}

func (config ConfigLXC) mapToApiShared() map[string]any {
	params := make(map[string]any)
	return params
}

func (config ConfigLXC) mapToApiUpdate(current ConfigLXC) (map[string]any, *[]byte) {
	var builder, deleteBuilder strings.Builder
	builderPtr := &builder
	deleteBuilderPtr := &deleteBuilder
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
			config.Networks.mapToApiUpdate(current.Networks, builderPtr, deleteBuilderPtr)
		} else {
			config.Networks.mapToApiCreate(builderPtr)
		}
	}
	if config.Protection != nil && (current.Protection == nil || *config.Protection != *current.Protection) {
		if *config.Protection {
			params[lxcAPIKeyProtection] = "1"
		} else {
			delete += "," + lxcAPIKeyProtection
		}
	}
	if config.StartAtNodeBoot != nil {
		if current.StartAtNodeBoot != nil {
			delete += startAtNodeBootMapToApiUpdate(params, *config.StartAtNodeBoot, *current.StartAtNodeBoot)
		} else {
			startAtNodeBootMapToApiCreate(params, *config.StartAtNodeBoot)
		}
	}
	if config.StartupShutdown != nil {
		if current.StartupShutdown != nil {
			delete += config.StartupShutdown.mapToApiUpdate(current.StartupShutdown, params)
		} else {
			config.StartupShutdown.mapToApiCreate(params)
		}
	}
	if config.Swap != nil && (current.Swap == nil || *config.Swap != *current.Swap) {
		params[lxcApiKeySwap] = int(*config.Swap)
	}
	if config.Tags != nil {
		if current.Tags != nil {
			if v, ok := config.Tags.mapToApiUpdate(*current.Tags); ok {
				builder.WriteString("&" + qemuApiKeyTags + "=")
				builder.WriteString(v)
			}
		} else {
			if v := config.Tags.mapToApiCreate(); v != "" {
				builder.WriteString("&" + qemuApiKeyTags + "=")
				builder.WriteString(v)
			}
		}
	}
	if delete != "" {
		params[lxcApiKeyDelete] = strings.TrimPrefix(delete, ",")
	}
	if deleteBuilder.Len() > 0 {
		if v, ok := params["delete"]; ok {
			params["delete"] = v.(string) + deleteBuilder.String()
		} else {
			builder.WriteString("&delete=")
			builder.WriteString(deleteBuilder.String()[1:]) // remove leading '&'
		}
	}
	if len(params) > 0 {
		params[lxcApiKeyDigest] = current.rawDigest.String()
	}
	var body *[]byte
	if builder.Len() > 0 {
		body = new(bytes.NewBufferString(builder.String()[1:]).Bytes())
	}
	if len(params) == 0 { // Temporarily clear params
		params = nil
	}
	return params, body
}

// Deprecated: use LxcGuestInterface.Update() instead.
func (config ConfigLXC) Update(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return c.New().LxcGuest.Update(ctx, *vmr, allowRestart, config)
}

// Deprecated: use LxcGuestInterface.UpdateNoCheck() instead.
func (config ConfigLXC) UpdateNoCheck(ctx context.Context, allowRestart bool, vmr *VmRef, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	return c.New().LxcGuest.UpdateNoCheck(ctx, *vmr, allowRestart, config)
}

func (config ConfigLXC) update_Unsafe(
	ctx context.Context,
	allowRestart bool,
	ref guestRef,
	current *ConfigLXC,
	currentState PowerState,
	version Version,
	c *clientAPI) error {

	var move []lxcMountMove
	var resize []lxcMountResize
	var getRootMount, getMounts, requiresOffStateForMountActions bool

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
		if err = ref.shutdown(ctx, c); err != nil {
			return err
		}
		currentState = PowerStateStopped // We assume the guest is stopped now
	}

	if requiresOffStateForMountActions || len(move) > 0 { // turn the guest off
		if currentState == PowerStateRunning || currentState == PowerStateUnknown { // Stop guest before moving disks
			if !allowRestart {
				return errors.New("guest has to be stopped before moving disks")
			}
			if err = ref.shutdownForce(ctx, c); err != nil { // We have to stop the guest before moving disks
				return err
			}
			currentState = PowerStateStopped // We assume the guest is stopped now
		}
	}

	if len(resize) > 0 || len(move) > 0 {

		for i := range move { // Move mounts
			if err = move[i].move(ctx, true, ref, c); err != nil {
				return err
			}
		}

		for i := range resize { // Resize mounts
			if err = resize[i].resize(ctx, ref, c); err != nil {
				return err
			}
		}

		var newCurrent *rawConfigLXC
		newCurrent, err = readLxcConfig(ctx, ref, version, c) // We have to refetch part of the current config
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
	params, body := config.mapToApiUpdate(*current)

	if body = combineParamsAndBody(params, body); body != nil {
		var urlBuilder strings.Builder
		urlBuilder.WriteString("/nodes/")
		urlBuilder.WriteString(ref.Node.String())
		urlBuilder.WriteString("/lxc/")
		urlBuilder.WriteString(ref.ID.String())
		urlBuilder.WriteString("/config")
		if err = c.putRawRetry(ctx, urlBuilder.String(), body, 3); err != nil {
			return err
		}
	}
	if currentState == PowerStateRunning || currentState == PowerStateUnknown { // If the guest is running, we have to check if it has pending changes
		var pendingChanges bool
		pendingChanges, err = ref.pendingChanges(ctx, c)
		if err != nil {
			return fmt.Errorf("error checking for pending changes: %w", err)
		}
		if pendingChanges {
			if !allowRestart {
				// TODO revert pending changes
				return errors.New("guest has to be restarted to apply changes")
			}
			if err = ref.reboot(ctx, c); err != nil {
				return fmt.Errorf("error restarting guest: %w", err)
			}
			currentState = PowerStateRunning // We assume the guest is running now
		}
	}
	if currentState != PowerStateRunning && targetState == PowerStateRunning { // We want the guest to be running, so we start it now
		if err != nil {
			ref.start(ctx, c)
		}
	}
	if config.Pool != nil {
		return ref.ID.setPool(ctx, c, *config.Pool, current.Pool, version)
	}
	return nil
}

func (config ConfigLXC) Validate(current *ConfigLXC, version Version) (err error) {
	if current != nil { // Update
		err = config.validateUpdate(*current, version)
	} else { // Create
		privileged := lxcDefaultPrivilege
		if config.Privileged != nil {
			privileged = *config.Privileged
		}
		err = config.validateCreate(privileged, version)
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

func (config ConfigLXC) validateCreate(privileged bool, version Version) (err error) {
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
	return config.Networks.Validate(nil, version)
}

func (config ConfigLXC) validateUpdate(current ConfigLXC, version Version) (err error) {
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
	return config.Networks.Validate(current.Networks, version)
}

type RawConfigLXC interface {
	Get(*PoolName, PowerState) *ConfigLXC
	GetArchitecture() LxcCpuArchitecture
	GetBootMount() *LxcBootMount
	GetCPU() *LxcCPU
	GetDNS() *GuestDNS
	GetDescription() *string
	GetDigest() [sha1.Size]byte
	GetFeatures() *LxcFeatures
	GetID() GuestID
	GetMemory() LxcMemory
	GetMounts() LxcMounts
	GetName() GuestName
	GetNetworks() LxcNetworks
	GetNode() NodeName
	GetOperatingSystem() OperatingSystem
	GetPrivileged() bool
	GetProtection() bool
	GetStartAtNodeBoot() bool
	GetStartupShutdown() *StartupAndShutdown
	GetSwap() LxcSwap
	GetTags() Tags
}

type rawConfigLXC struct {
	a       map[string]any
	node    NodeName
	guestID GuestID
	version Version
}

var _ RawConfigLXC = (*rawConfigLXC)(nil)

func (raw *rawConfigLXC) Get(pool *PoolName, state PowerState) *ConfigLXC {
	var poolName PoolName
	if pool != nil {
		poolName = *pool
	}
	config := raw.get(poolName)
	config.Digest = config.rawDigest.sha1()
	if state != PowerStateUnknown {
		config.State = &state
	}
	return config
}

func (raw *rawConfigLXC) get(pool PoolName) *ConfigLXC {
	privileged := raw.isPrivileged()
	config := ConfigLXC{
		Architecture:    raw.GetArchitecture(),
		BootMount:       raw.getBootMount(privileged),
		CPU:             raw.GetCPU(),
		DNS:             raw.GetDNS(),
		Description:     raw.GetDescription(),
		Features:        raw.getFeatures(privileged),
		ID:              new(raw.GetID()),
		Memory:          new(raw.GetMemory()),
		Mounts:          raw.getMounts(privileged),
		Name:            new(raw.GetName()),
		Networks:        raw.GetNetworks(),
		Node:            new(raw.GetNode()),
		OperatingSystem: raw.GetOperatingSystem(),
		Privileged:      &privileged,
		Protection:      new(raw.GetProtection()),
		StartAtNodeBoot: new(raw.GetStartAtNodeBoot()),
		StartupShutdown: raw.GetStartupShutdown(),
		Swap:            new(raw.GetSwap()),
		Tags:            new(raw.GetTags()),
		rawDigest:       raw.getDigest()}
	if pool != "" {
		config.Pool = &pool
	}
	return &config
}

func (raw *rawConfigLXC) GetArchitecture() LxcCpuArchitecture {
	if v, isSet := raw.a[lxcApiKeyArchitecture]; isSet {
		return LxcCpuArchitecture(v.(string))
	}
	return ""
}

func (raw *rawConfigLXC) GetDescription() *string {
	if v, isSet := raw.a[lxcApiKeyDescription]; isSet {
		return new(strings.TrimSpace(v.(string)))
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

func (raw *rawConfigLXC) GetID() GuestID { return raw.guestID }

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

func (raw *rawConfigLXC) GetNode() NodeName { return raw.node }

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

func (raw *rawConfigLXC) GetStartAtNodeBoot() bool { return startAtNodeBootMapToSDK(raw.a) }

func (raw *rawConfigLXC) GetStartupShutdown() *StartupAndShutdown {
	return StartupAndShutdown{}.mapToSDK(raw.a)
}

func (raw *rawConfigLXC) GetSwap() LxcSwap {
	if v, isSet := raw.a[lxcApiKeySwap]; isSet {
		return LxcSwap(v.(float64))
	}
	return 0
}

func (raw *rawConfigLXC) GetTags() Tags {
	var t Tags
	if v, isSet := raw.a[lxcApiKeyTags]; isSet {
		t.mapToSDK(v.(string))
	}
	return t
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
		builder := strings.Builder{}
		first := true
		for i := range config.PublicSSHkeys {
			if v := config.PublicSSHkeys[i].String(); v != "" {
				if !first {
					builder.WriteString("\n")
				}
				first = false
				builder.WriteString(v)
			}
		}
		params[lxcApiKeySSHPublicKeys] = builder.String()
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

// Deprecated: use LxcGuestInterface.Read() instead.
func NewRawConfigLXCFromAPI(ctx context.Context, vmr *VmRef, c *Client) (RawConfigLXC, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	if c == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	return c.New().LxcGuest.Read(ctx, *vmr)
}

// Deprecated: LxcGuestInterface.ReadActive() instead.
// NewActiveRawConfigLXCFromApi returns the active configuration of the LXC guest.
// Without pending changes.
func NewActiveRawConfigLXCFromApi(ctx context.Context, vmr *VmRef, c *Client) (raw RawConfigLXC, pending bool, err error) {
	return c.New().LxcGuest.ReadActive(ctx, *vmr)
}

func guestGetActiveRawLxcConfig(ctx context.Context, ref guestRef, version Version, c *clientAPI) (raw RawConfigLXC, pending bool, err error) {
	var tmpConfig map[string]any
	tmpConfig, pending, err = ref.pendingConfig(ctx, c)
	if err != nil {
		return nil, false, err
	}
	return &rawConfigLXC{
		a:       tmpConfig,
		node:    ref.Node,
		guestID: ref.ID,
		version: version,
	}, pending, nil
}

func readLxcConfig(ctx context.Context, ref guestRef, version Version, c *clientAPI) (*rawConfigLXC, error) {
	rawConfig, err := c.getGuestConfig(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &rawConfigLXC{
		a:       rawConfig,
		guestID: ref.ID,
		node:    ref.Node,
		version: version,
	}, nil
}

func readLxcInterfaceStatus(ctx context.Context, ref guestRef, c *clientAPI) (*rawLxcInfoNetworkInterfaces, error) {
	raw, err := c.getList(ctx, "/nodes/"+ref.Node.String()+"/lxc/"+ref.ID.String()+"/interfaces", "", "")
	if err != nil {
		return nil, err
	}
	return &rawLxcInfoNetworkInterfaces{a: raw}, nil
}
