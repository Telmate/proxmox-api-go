package storagesubtests

import (
	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var ZFSoverISCSI_LioFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     proxmox.PointerString("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "lio",
		LIO: &proxmox.ConfigStorageZFSoverISCSI_LIO{
			TargetPortalGroup: "t-group",
		},
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

var ZFSoverISCSI_LioEmpty = proxmox.ConfigStorage{
	Type: "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Target:        "test-target",
		ISCSIprovider: "lio",
		LIO: &proxmox.ConfigStorageZFSoverISCSI_LIO{
			TargetPortalGroup: "t-group",
		},
	},
}
