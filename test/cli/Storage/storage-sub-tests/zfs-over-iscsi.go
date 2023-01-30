package storagesubtests

import (
	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var ZFSoverISCSIFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     proxmox.PointerString("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "iet",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
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
