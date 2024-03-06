package storagesubtests

import (
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var PBSFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "pbs",
	PBS: &proxmox.ConfigStoragePBS{
		Server:      "10.20.1.1",
		Datastore:   "proxmox",
		Username:    "root@pam",
		Fingerprint: "B7:BC:55:10:CC:1C:63:7B:5E:5F:B7:85:81:6A:77:3D:BB:39:4B:68:33:7B:1B:11:7C:A5:AB:43:CC:F7:78:CF",
		Port:        util.Pointer(8007),
	},
	Content: &proxmox.ConfigStorageContent{
		Backup: util.Pointer(true),
	},
	BackupRetention: &proxmox.ConfigStorageBackupRetention{
		Last:    util.Pointer(6),
		Hourly:  util.Pointer(5),
		Daily:   util.Pointer(4),
		Monthly: util.Pointer(3),
		Weekly:  util.Pointer(2),
		Yearly:  util.Pointer(1),
	},
}

var PBSEmpty = proxmox.ConfigStorage{
	Type: "pbs",
	PBS: &proxmox.ConfigStoragePBS{
		Server:    "10.20.1.1",
		Datastore: "proxmox",
		Username:  "root@pam",
	},
}
