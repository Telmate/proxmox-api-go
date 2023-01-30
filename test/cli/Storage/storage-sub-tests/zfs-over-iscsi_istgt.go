package storagesubtests

import (
	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var ZFSoverISCSI_IstgtFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     proxmox.PointerString("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "istgt",
		Istgt: &proxmox.ConfigStorageZFSoverISCSI_istgt{
			Writecache: true,
		},
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

var ZFSoverISCSI_IstgtEmpty = proxmox.ConfigStorage{
	Type: "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Target:        "test-target",
		ISCSIprovider: "istgt",
		Istgt:         &proxmox.ConfigStorageZFSoverISCSI_istgt{},
	},
}
