package proxmox

import (
	"context"
	"errors"
)

const (
	vmRefQemu                     string = "qemu"
	vmRefLXC                      string = "lxc"
	clone_Error_MutuallyExclusive string = "linked and full clone are mutually exclusive"
	clone_Error_NoneSet           string = "either linked nor full clone must be set"
)

// CloneLxc clones a new LXC container by cloning current container
func (vmr *VmRef) CloneLxc(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, Task, error) {
	if vmr == nil {
		return nil, nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, nil, err
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

// CloneLxcNoCheck creates a new LXC container by cloning the current container, without input validation.
func (vmr *VmRef) CloneLxcNoCheck(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, Task, error) {
	if vmr == nil {
		return nil, nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneLxc_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneLxc_Unsafe(ctx context.Context, settings CloneLxcTarget, c *Client) (*VmRef, Task, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/lxc/" + vmr.vmId.String() + "/clone"
	var task Task
	if id == 0 {
		id, task, err = guestCreateLoop(ctx, "newid", url, params, c)
	} else {
		task, err = c.postWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: vmRefLXC}, task, nil
}

// CloneQemu creates a new Qemu VM by cloning the current VM.
func (vmr *VmRef) CloneQemu(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, Task, error) {
	if vmr == nil {
		return nil, nil, errors.New(VmRef_Error_Nil)
	}
	err := settings.Validate()
	if err != nil {
		return nil, nil, err
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

// CloneQemuNoCheck creates a new VM by cloning the current VM, without input validation.
func (vmr *VmRef) CloneQemuNoCheck(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, Task, error) {
	if vmr == nil {
		return nil, nil, errors.New(VmRef_Error_Nil)
	}
	return vmr.cloneQemu_Unsafe(ctx, settings, c)
}

func (vmr *VmRef) cloneQemu_Unsafe(ctx context.Context, settings CloneQemuTarget, c *Client) (*VmRef, Task, error) {
	id, node, pool, params := settings.mapToAPI()
	var err error
	url := "/nodes/" + vmr.node.String() + "/qemu/" + vmr.vmId.String() + "/clone"
	var task Task
	if id == 0 {
		id, task, err = guestCreateLoop(ctx, "newid", url, params, c)
	} else {
		task, err = c.postWithTask(ctx, params, url)
	}
	if err != nil {
		return nil, nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: vmRefQemu}, task, nil
}

func (vmr *VmRef) Migrate(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) (Task, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	if err := newNode.Validate(); err != nil {
		return nil, err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) MigrateNoCheck(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) (Task, error) {
	if vmr == nil {
		return nil, errors.New(VmRef_Error_Nil)
	}
	if err := c.checkInitialized(); err != nil {
		return nil, err
	}
	return vmr.migrate_Unsafe(ctx, c, newNode, LiveMigrate)
}

func (vmr *VmRef) migrate_Unsafe(ctx context.Context, c *Client, newNode NodeName, LiveMigrate bool) (Task, error) {
	params := map[string]interface{}{
		"target":           newNode.String(),
		"with-local-disks": 1,
	}
	if LiveMigrate {
		params["online"] = 1
	}
	return c.postWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/migrate")
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
	Node NodeName  `json:"node"`
	ID   *GuestID  `json:"id,omitempty"`   // Optional
	Name *string   `json:"name,omitempty"` // Optional // TODO replace one we have a type for it
	Pool *PoolName `json:"pool,omitempty"` // Optional
}

func (linked CloneLinked) Validate() (err error) {
	if linked.ID != nil {
		if err = linked.ID.Validate(); err != nil {
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
	Node    NodeName  `json:"node"`
	ID      *GuestID  `json:"id,omitempty"`      // Optional
	Name    *string   `json:"name,omitempty"`    // Optional // TODO replace one we have a type for it
	Pool    *PoolName `json:"pool,omitempty"`    // Optional
	Storage *string   `json:"storage,omitempty"` // Optional // TODO replace one we have a type for it
}

func (full CloneLxcFull) Validate() (err error) {
	if full.ID != nil {
		if err = full.ID.Validate(); err != nil {
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
	Name          *string         `json:"name,omitempty"`    // Optional // TODO replace one we have a type for it
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
	Name          *string // TODO replace one we have a type for it
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
		params[clone.nameFlag] = *clone.Name
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
