package proxmox

import (
	"errors"
	"strings"
)

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

func (target CloneLxcTarget) mapToAPI() (GuestID, NodeName, PoolName, *[]byte) {
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

func (target CloneQemuTarget) mapToAPI() (GuestID, NodeName, PoolName, *[]byte) {
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
	ID   *GuestID   `json:"id,omitempty"`   // Optional
	Name *GuestName `json:"name,omitempty"` // Optional
	Pool *PoolName  `json:"pool,omitempty"` // Optional
	Node NodeName   `json:"node"`
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

func (linked CloneLinked) mapToAPI(nameFlag string) (GuestID, NodeName, PoolName, *[]byte) {
	return cloneSettings{
		FullClone: false,
		ID:        linked.ID,
		nameFlag:  nameFlag,
		Name:      linked.Name,
		Node:      linked.Node,
		Pool:      linked.Pool}.mapToAPI()
}

type CloneLxcFull struct {
	ID      *GuestID     `json:"id,omitempty"`      // Optional
	Name    *GuestName   `json:"name,omitempty"`    // Optional
	Pool    *PoolName    `json:"pool,omitempty"`    // Optional
	Storage *StorageName `json:"storage,omitempty"` // OptionaL
	Node    NodeName     `json:"node"`
}

func (full CloneLxcFull) Validate() error {
	if full.ID != nil {
		if err := full.ID.Validate(); err != nil {
			return err
		}
	}
	if full.Name != nil {
		if err := full.Name.Validate(); err != nil {
			return err
		}
	}
	if full.Pool != nil {
		if err := full.Pool.Validate(); err != nil {
			return err
		}
	}
	if full.Storage != nil {
		if err := full.Storage.Validate(); err != nil {
			return err
		}
	}
	return full.Node.Validate()
}

func (full CloneLxcFull) mapToAPI() (GuestID, NodeName, PoolName, *[]byte) {
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
	ID            *GuestID        `json:"id,omitempty"`      // Optional
	Name          *GuestName      `json:"name,omitempty"`    // Optional
	Pool          *PoolName       `json:"pool,omitempty"`    // Optional
	Storage       *StorageName    `json:"storage,omitempty"` // Optional
	StorageFormat *QemuDiskFormat `json:"format,omitempty"`  // Optional
	Node          NodeName        `json:"node"`
}

func (full CloneQemuFull) Validate() error {
	if full.ID != nil {
		if err := full.ID.Validate(); err != nil {
			return err
		}
	}
	if full.Name != nil {
		if err := full.Name.Validate(); err != nil {
			return err
		}
	}
	if full.Pool != nil {
		if err := full.Pool.Validate(); err != nil {
			return err
		}
	}
	if full.Storage != nil {
		if err := full.Storage.Validate(); err != nil {
			return err
		}
	}
	if full.StorageFormat != nil {
		if err := full.StorageFormat.Validate(); err != nil {
			return err
		}
	}
	return full.Node.Validate()
}

func (full CloneQemuFull) mapToAPI() (GuestID, NodeName, PoolName, *[]byte) {
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
	ID            *GuestID
	Name          *GuestName
	Pool          *PoolName
	Storage       *StorageName
	StorageFormat *QemuDiskFormat
	Node          NodeName
	nameFlag      string
	FullClone     bool
}

func (clone cloneSettings) mapToAPI() (GuestID, NodeName, PoolName, *[]byte) {
	var b strings.Builder
	b.WriteString("target=")
	b.WriteString(clone.Node.String())
	if clone.FullClone {
		b.WriteString("&full=1")
	}
	var id GuestID
	if clone.ID != nil {
		id = *clone.ID
		b.WriteString("&newid=")
		b.WriteString(id.String())
	}
	if clone.Name != nil {
		b.WriteByte('&')
		b.WriteString(clone.nameFlag)
		b.WriteByte('=')
		b.WriteString(clone.Name.String())
	}
	var pool PoolName
	if clone.Pool != nil {
		pool = *clone.Pool
		b.WriteString("&pool=")
		b.WriteString(pool.String())
	}
	if clone.Storage != nil {
		b.WriteString("&storage=")
		b.WriteString(clone.Storage.String())
	}
	if clone.StorageFormat != nil {
		b.WriteString("&format=")
		b.WriteString(clone.StorageFormat.String())
	}
	return id, clone.Node, pool, new([]byte(b.String()))
}
