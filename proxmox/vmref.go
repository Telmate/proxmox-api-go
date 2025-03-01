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
		id, err = guestCreateLoop(ctx, "newid", url, params, c)
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
		vmType: vmRefLXC}, nil
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
		id, err = guestCreateLoop(ctx, "newid", url, params, c)
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
		vmType: vmRefQemu}, nil
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
	Full   *CloneQemuFull
	Linked *CloneLinked
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
	Node NodeName
	ID   *GuestID  // Optional
	Name *string   // Optional // TODO replace one we have a type for it
	Pool *PoolName // Optional
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
	Node    NodeName
	ID      *GuestID  // Optional
	Name    *string   // Optional // TODO replace one we have a type for it
	Pool    *PoolName // Optional
	Storage *string   // Optional // TODO replace one we have a type for it
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
	Node          NodeName
	ID            *GuestID        // Optional
	Name          *string         // Optional // TODO replace one we have a type for it
	Pool          *PoolName       // Optional
	Storage       *string         // Optional // TODO replace one we have a type for it
	StorageFormat *QemuDiskFormat // Optional
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
