package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
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
		Container: util.Pointer(true),
		DiskImage: util.Pointer(true),
	},
}

var LVMThinEmpty = proxmox.ConfigStorage{
	Type: "lvm-thin",
	LVMThin: &proxmox.ConfigStorageLVMThin{
		VGname:   "pve",
		Thinpool: "data",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: util.Pointer(true),
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
	s.Content.DiskImage = util.Pointer(false)
	Get(s, name, t)
}
