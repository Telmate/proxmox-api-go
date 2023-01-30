package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var LVMThinFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "lvm-thin",
	LVMThin: &proxmox.ConfigStorageLVMThin{
		VGname:   "pve",
		Thinpool: "data",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
	},
}

var LVMThinEmpty = proxmox.ConfigStorage{
	Type: "lvm-thin",
	LVMThin: &proxmox.ConfigStorageLVMThin{
		VGname:   "pve",
		Thinpool: "data",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
	},
}

func LVMThinGetFull(name string, t *testing.T) {
	s := CloneJson(LVMThinFull)
	s.ID = name
	Get(s, name, t)
}

func LVMThinGetEmpty(name string, t *testing.T) {
	s := CloneJson(LVMThinEmpty)
	s.ID = name
	s.Content.DiskImage = proxmox.PointerBool(false)
	Get(s, name, t)
}
