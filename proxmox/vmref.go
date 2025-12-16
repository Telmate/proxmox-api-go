package proxmox

import (
	"context"
	"errors"
	"time"
)

const (
	vmRefQemu                     string = "qemu"
	vmRefLXC                      string = "lxc"
	clone_Error_MutuallyExclusive string = "linked and full clone are mutually exclusive"
	clone_Error_NoneSet           string = "either linked nor full clone must be set"
)

// CloneLxc clones a new LXC container by cloning current container
func (vmr *VmRef) CloneLxc(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, err
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

// CloneLxcNoCheck creates a new LXC container by cloning the current container, without input validation.
func (vmr *VmRef) CloneLxcNoCheck(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneLxc_Unsafe(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/lxc/" + vmr.vmId.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, params, c)
	} else {
		_, err = c.PostWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestLxc}, nil
}

// CloneQemu creates a new Qemu VM by cloning the current VM.
func (vmr *VmRef) CloneQemu(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, err
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

// CloneQemuNoCheck creates a new VM by cloning the current VM, without input validation.
func (vmr *VmRef) CloneQemuNoCheck(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneQemu_Unsafe(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/qemu/" + vmr.vmId.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, params, c)
	} else {
		_, err = c.PostWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestQemu}, nil
}

func (vmr VmRef) Delete(ctx context.Context, c *Client) error { return c.new().guestDelete(ctx, &vmr) }

func (c *clientNew) guestDelete(ctx context.Context, vmr *VmRef) error {
	guestID := vmr.VmId()
	if guestID == 0 {
		return errors.New(VmRef_Error_IDnotSet)
	}
	ca := c.apiGet()
	rawGuests, err := listGuests_Unsafe(ctx, ca)
	if err != nil {
		return err
	}

	rawGuest, ok := rawGuests.SelectID(guestID)
	if !ok {
		return errorMsg{}.guestDoesNotExist(vmr.vmId)
	}

	guestType := rawGuest.GetType()
	vmr.node = rawGuest.GetNode()
	vmr.vmType = guestType

	var protection bool // Check if guest is protected
	switch guestType {
	case GuestQemu:
		rawConfig, err := guestGetRawQemuConfig_Unsafe(ctx, vmr, ca)
		if err != nil {
			return err
		}
		protection = rawConfig.GetProtection()
	case GuestLxc:
		rawConfig, err := guestGetLxcRawConfig_Unsafe(ctx, vmr, ca)
		if err != nil {
			return err
		}
		protection = rawConfig.GetProtection()
	}
	if protection {
		return errorMsg{}.guestIsProtectedCantDelete(guestID)
	}

	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}

	haEnabled := rawGuest.GetHaState() != ""
	if rawGuest.GetStatus() != PowerStateStopped { // Check if guest is running
		if haEnabled {
			if _, err = guestID.deleteHaResource(ctx, ca); err != nil { // It's faster to delete HA resource first, instead of stopping via HA
				return err
			}
			haEnabled = false
		}
		for {
			if version.Encode() >= version_8_0_0 { // Try to force stop the guest if supported
				err = vmr.forceStop_Unsafe(ctx, ca)
			} else {
				err = vmr.stop_Unsafe(ctx, ca)
			}
			if err != nil {
				return err
			}
			var guestStatus RawGuestStatus
			guestStatus, err = vmr.getRawGuestStatus_Unsafe(ctx, c.oldClient)
			if err != nil {
				return err
			}
			if guestStatus.GetState() == PowerStateStopped {
				break
			}
		}
	}
	return vmr.delete_Unsafe(ctx, c.apiGet(), haEnabled)
}

func (vmr VmRef) DeleteNoCheck(ctx context.Context, c *Client) error {
	if err := c.checkInitialized(); err != nil {
		return err
	}
	return vmr.delete_Unsafe(ctx, c.new().apiGet(), false)
}

func (vmr *VmRef) delete_Unsafe(ctx context.Context, c clientApiInterface, purge bool) error {
	return c.deleteGuest(ctx, vmr, purge)
}

func (c *clientAPI) deleteGuest(ctx context.Context, vmr *VmRef, purge bool) error {
	var urlOptions string
	if purge {
		urlOptions = "?purge=1"
	}
	return c.deleteTask(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+urlOptions)
}

// ForceStop stops the guest immediately without a graceful shutdown and cancels any stop/shutdown operations in progress.
// This function requires Proxmox VE 8.0 or later.
func (vmr *VmRef) ForceStop(ctx context.Context, c *Client) error {
	return c.new().guestStopForce(ctx, vmr)
}

func (c *clientNew) guestStopForce(ctx context.Context, vmr *VmRef) error {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	if version.Encode() < version_8_0_0 {
		return functionalityNotSupportedInVersion("force stop", version)
	}
	if err := c.oldClient.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return vmr.forceStop_Unsafe(ctx, c.apiGet())
}

// forceStop stops the guest immediately without a graceful shutdown and cancels any stop/shutdown operations in progress.
// This function requires Proxmox VE 8.0 or later.
// Gives error when HA is enabled.
func (vmr *VmRef) forceStop_Unsafe(ctx context.Context, c clientApiInterface) error {
	return c.updateGuestStatus(ctx, vmr, "stop", map[string]any{"overrule-shutdown": int(1)})
}

func (vmr *VmRef) GetRawGuestStatus(ctx context.Context, c *Client) (RawGuestStatus, error) {
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return vmr.getRawGuestStatus_Unsafe(ctx, c)
}

func (vmr *VmRef) getRawGuestStatus_Unsafe(ctx context.Context, c *Client) (RawGuestStatus, error) {
	return c.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/status/current", "vm", "STATE")
}

func (vmr *VmRef) Migrate(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	if err := newNode.Validate(); err != nil {
		return err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) MigrateNoCheck(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) migrate_Unsafe(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) error {
	params := map[string]interface{}{
		"target":           newNode.String(),
		"with-local-disks": 1,
	}
	if LiveMigrate {
		params["online"] = 1
	}
	_, err := c.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/migrate")
	return err
}

func (vmr *VmRef) PendingChanges(ctx context.Context, c *Client) (bool, error) {
	return c.new().guestCheckPendingChanges(ctx, vmr)
}

func (vmr *VmRef) pendingChanges(ctx context.Context, c clientApiInterface) (bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, vmr)
	if err != nil {
		return false, err
	}
	for _, item := range changes {
		m := item.(map[string]any)
		// we always have the key `key`
		if _, ok := m[pendingApiValueKey]; ok {
			if len(m) > 2 {
				return true, nil
			}
		} else if len(m) > 1 {
			return true, nil
		}
	}
	return false, nil
}

func (c *clientNew) guestCheckPendingChanges(ctx context.Context, vmr *VmRef) (bool, error) {
	return vmr.pendingChanges(ctx, c.apiGet())
}

func (vmr *VmRef) pendingConfig(ctx context.Context, c clientApiInterface) (map[string]any, bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, vmr)
	if err != nil {
		return nil, false, err
	}
	var pending bool
	config := make(map[string]any, len(changes))
	for _, item := range changes {
		m := item.(map[string]any)
		// we always have the key `key`
		if v, ok := m[pendingApiValueKey]; ok {
			config[m[pendingApiKeyKey].(string)] = v
			if len(m) > 2 {
				pending = true
			}
		} else if len(m) > 1 {
			pending = true
		}
	}
	return config, pending, nil
}

const (
	pendingApiKeyKey   string = "key"
	pendingApiValueKey string = "value"
)

func (vmr *VmRef) Stop(ctx context.Context, c *Client) error { return c.new().guestStop(ctx, vmr) }

func (c *clientNew) guestStop(ctx context.Context, vmr *VmRef) error {
	if err := c.oldClient.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	return vmr.stop_Unsafe(ctx, c.apiGet())
}

func (vmr *VmRef) stop_Unsafe(ctx context.Context, c clientApiInterface) error {
	return c.updateGuestStatus(ctx, vmr, "stop", nil)
}

const (
	cloneLxcFlagName  string = "hostname"
	cloneQemuFlagName string = "name"
)

type CloneLxcTarget struct {
	Full   *CloneLxcFull
	Linked *CloneLinked
}

const (
	CloneLxcTarget_Error_MutualExclusivity = clone_Error_MutuallyExclusive
	CloneLxcTarget_Error_NoneSet           = clone_Error_NoneSet
)

func (target CloneLxcTarget) Validate() error {
	if target.Full == nil && target.Linked == nil {
		return errors.New(CloneQemuTarget_Error_NoneSet)
	}
	if target.Full != nil && target.Linked != nil {
		return errors.New(CloneQemuTarget_Error_MutualExclusivity)
	}
	if target.Full != nil {
		return target.Full.Validate()
	}
	return target.Linked.Validate()
}

func (target CloneLxcTarget) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	if target.Linked != nil {
		return target.Linked.mapToAPI(cloneLxcFlagName)
	}
	if target.Full != nil {
		return target.Full.mapToAPI()
	}
	return 0, "", "", nil
}

type CloneQemuTarget struct {
	Full   *CloneQemuFull `json:"full,omitempty"`
	Linked *CloneLinked   `json:"linked,omitempty"`
}

const (
	CloneQemuTarget_Error_MutualExclusivity = clone_Error_MutuallyExclusive
	CloneQemuTarget_Error_NoneSet           = clone_Error_NoneSet
)

func (target CloneQemuTarget) Validate() error {
	if target.Full == nil && target.Linked == nil {
		return errors.New(CloneQemuTarget_Error_NoneSet)
	}
	if target.Full != nil && target.Linked != nil {
		return errors.New(CloneQemuTarget_Error_MutualExclusivity)
	}
	if target.Full != nil {
		return target.Full.Validate()
	}
	return target.Linked.Validate()
}

func (target CloneQemuTarget) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	if target.Linked != nil {
		return target.Linked.mapToAPI(cloneQemuFlagName)
	}
	if target.Full != nil {
		return target.Full.mapToAPI()
	}
	return 0, "", "", nil
}

// Linked Clone in the same for both LXC and QEMU
type CloneLinked struct {
	Node NodeName   `json:"node"`
	ID   *GuestID   `json:"id,omitempty"`   // Optional
	Name *GuestName `json:"name,omitempty"` // Optional
	Pool *PoolName  `json:"pool,omitempty"` // Optional
}

func (linked CloneLinked) Validate() (err error) {
	if linked.ID != nil {
		if err = linked.ID.Validate(); err != nil {
			return
		}
	}
	if linked.Name != nil {
		if err = linked.Name.Validate(); err != nil {
			return
		}
	}
	if linked.Pool != nil {
		if err = linked.Pool.Validate(); err != nil {
			return
		}
	}
	return linked.Node.Validate()
}

func (linked CloneLinked) mapToAPI(nameFlag string) (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone: false,
		ID:        linked.ID,
		nameFlag:  nameFlag,
		Name:      linked.Name,
		Node:      linked.Node,
		Pool:      linked.Pool}.mapToAPI()
}

type CloneLxcFull struct {
	Node    NodeName   `json:"node"`
	ID      *GuestID   `json:"id,omitempty"`      // Optional
	Name    *GuestName `json:"name,omitempty"`    // Optional
	Pool    *PoolName  `json:"pool,omitempty"`    // Optional
	Storage *string    `json:"storage,omitempty"` // Optional // TODO replace one we have a type for it
}

func (full CloneLxcFull) Validate() (err error) {
	if full.ID != nil {
		if err = full.ID.Validate(); err != nil {
			return
		}
	}
	if full.Name != nil {
		if err = full.Name.Validate(); err != nil {
			return
		}
	}
	if full.Pool != nil {
		if err = full.Pool.Validate(); err != nil {
			return
		}
	}
	return full.Node.Validate()
}

func (full CloneLxcFull) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone: true,
		ID:        full.ID,
		nameFlag:  cloneLxcFlagName,
		Name:      full.Name,
		Node:      full.Node,
		Pool:      full.Pool,
		Storage:   full.Storage}.mapToAPI()
}

type CloneQemuFull struct {
	Node          NodeName        `json:"node"`
	ID            *GuestID        `json:"id,omitempty"`      // Optional
	Name          *GuestName      `json:"name,omitempty"`    // Optional
	Pool          *PoolName       `json:"pool,omitempty"`    // Optional
	Storage       *string         `json:"storage,omitempty"` // Optional // TODO replace one we have a type for it
	StorageFormat *QemuDiskFormat `json:"format,omitempty"`  // Optional
}

func (full CloneQemuFull) Validate() (err error) {
	if full.ID != nil {
		if err = full.ID.Validate(); err != nil {
			return
		}
	}
	if full.Name != nil {
		if err = full.Name.Validate(); err != nil {
			return
		}
	}
	if full.Pool != nil {
		if err = full.Pool.Validate(); err != nil {
			return
		}
	}
	if full.StorageFormat != nil {
		if err = full.StorageFormat.Validate(); err != nil {
			return
		}
	}
	return full.Node.Validate()
}

func (full CloneQemuFull) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	return cloneSettings{
		FullClone:     true,
		ID:            full.ID,
		nameFlag:      cloneQemuFlagName,
		Name:          full.Name,
		Node:          full.Node,
		Pool:          full.Pool,
		Storage:       full.Storage,
		StorageFormat: full.StorageFormat}.mapToAPI()
}

type cloneSettings struct {
	FullClone     bool
	ID            *GuestID
	nameFlag      string
	Name          *GuestName
	Node          NodeName
	Pool          *PoolName
	Storage       *string // TODO replace one we have a type for it
	StorageFormat *QemuDiskFormat
}

func (clone cloneSettings) mapToAPI() (GuestID, NodeName, PoolName, map[string]interface{}) {
	params := map[string]interface{}{
		"target": clone.Node.String(),
		"full":   clone.FullClone,
	}
	var id GuestID
	if clone.ID != nil {
		id = *clone.ID
		params["newid"] = int(id)
	}
	if clone.Name != nil {
		params[clone.nameFlag] = (*clone.Name).String()
	}
	var pool PoolName
	if clone.Pool != nil {
		pool = *clone.Pool
		params["pool"] = pool.String()
	}
	if clone.Storage != nil {
		params["storage"] = *clone.Storage
	}
	if clone.StorageFormat != nil {
		params["format"] = clone.StorageFormat.String()
	}
	return id, clone.Node, pool, params
}

// TODO add more properties to GuestStatus
type GuestStatus struct {
	Name   GuestName     `json:"name"`
	State  PowerState    `json:"state"`
	Uptime time.Duration `json:"uptime"`
}

type RawGuestStatus map[string]any

func (raw RawGuestStatus) GetName() GuestName {
	if v, isSet := raw["name"]; isSet {
		if name, ok := v.(string); ok {
			return GuestName(name)
		}
	}
	return ""
}

func (raw RawGuestStatus) Get() GuestStatus {
	return GuestStatus{
		Name:   raw.GetName(),
		State:  raw.GetState(),
		Uptime: raw.GetUptime()}
}

func (raw RawGuestStatus) GetState() PowerState {
	if v, isSet := raw["status"]; isSet {
		if state, ok := v.(string); ok {
			return PowerState(0).parse(state)
		}
	}
	return PowerStateUnknown
}

func (raw RawGuestStatus) GetUptime() time.Duration {
	if v, isSet := raw["uptime"]; isSet {
		if uptime, ok := v.(float64); ok {
			return time.Duration(uptime) * time.Second
		}
	}
	return 0
}
