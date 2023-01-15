package storagesubtests

import (
	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var ZFSoverISCSI_ComstarFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs-over-iscsi",
	ZFSoverISCSI: &proxmox.ConfigStorageZFSoverISCSI{
		Portal:        "test-portal",
		Pool:          "test-pool",
		Blocksize:     proxmox.PointerString("8k"),
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
		DiskImage: proxmox.PointerBool(true),
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
