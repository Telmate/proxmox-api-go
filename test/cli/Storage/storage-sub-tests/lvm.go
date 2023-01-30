package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var LVMFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "lvm",
	LVM: &proxmox.ConfigStorageLVM{
		VGname: "TestVolumeGroup",
		Shared: true,
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
	},
}

var LVMEmpty = proxmox.ConfigStorage{
	Type: "lvm",
	LVM: &proxmox.ConfigStorageLVM{
		VGname: "TestVolumeGroup",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

func LVMGetFull(name string, t *testing.T) {
	s := CloneJson(LVMFull)
	s.ID = name
	Get(s, name, t)
}

func LVMGetEmpty(name string, t *testing.T) {
	s := CloneJson(LVMEmpty)
	s.ID = name
	s.Content.Container = proxmox.PointerBool(false)
	Get(s, name, t)
}
