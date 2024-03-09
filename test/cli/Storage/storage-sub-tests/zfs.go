package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var ZFSFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs",
	ZFS: &proxmox.ConfigStorageZFS{
		Pool:          "test-pool",
		Blocksize:     util.Pointer("4k"),
		Thinprovision: true,
	},
	Content: &proxmox.ConfigStorageContent{
		Container: util.Pointer(true),
		DiskImage: util.Pointer(true),
	},
}

var ZFSEmpty = proxmox.ConfigStorage{
	Type: "zfs",
	ZFS: &proxmox.ConfigStorageZFS{
		Pool: "test-pool",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: util.Pointer(true),
	},
}

func ZFSGetFull(name string, t *testing.T) {
	s := CloneJson(ZFSFull)
	s.ID = name
	Get(s, name, t)
}

func ZFSGetEmpty(name string, t *testing.T) {
	s := CloneJson(ZFSEmpty)
	s.ID = name
	s.ZFS.Blocksize = util.Pointer("8k")
	s.Content.Container = util.Pointer(false)
	Get(s, name, t)
}
