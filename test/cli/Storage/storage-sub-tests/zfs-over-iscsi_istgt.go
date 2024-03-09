package storagesubtests

import (
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var ZFSoverISCSI_IstgtFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     util.Pointer("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "istgt",
		Istgt: &proxmox.ConfigStorageZFSoverISCSI_istgt{
			Writecache: true,
		},
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: util.Pointer(true),
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
