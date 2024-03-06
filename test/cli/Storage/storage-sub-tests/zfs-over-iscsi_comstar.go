package storagesubtests

import (
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var ZFSoverISCSI_ComstarFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     util.Pointer("8k"),
		Target:        "test-target",
		Thinprovision: true,
		ISCSIprovider: "comstar",
		Comstar: &proxmox.ConfigStorageZFSoverISCSI_Comstar{
			TargetGroup: "t-group",
			HostGroup:   "h-group",
			Writecache:  true,
		},
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: util.Pointer(true),
	},
}

var ZFSoverISCSI_ComstarEmpty = proxmox.ConfigStorage{
	Type: "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Target:        "test-target",
		ISCSIprovider: "comstar",
		Comstar:       &proxmox.ConfigStorageZFSoverISCSI_Comstar{},
	},
}
