package storagesubtests

import (
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var ZFSoverISCSIFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     util.Pointer("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "iet",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: util.Pointer(true),
	},
}

var ZFSoverISCSIEmpty = proxmox.ConfigStorage{
	Type: "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Target:        "test-target",
		ISCSIprovider: "iet",
	},
}
