package storagesubtests

import (
	"github.com/perimeter-81/proxmox-api-go/proxmox"
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
		Port:        proxmox.PointerInt(8007),
	},
	Content: &proxmox.ConfigStorageContent{
		Backup: proxmox.PointerBool(true),
	},
	BackupRetention: &proxmox.ConfigStorageBackupRetention{
		Last:    proxmox.PointerInt(6),
		Hourly:  proxmox.PointerInt(5),
		Daily:   proxmox.PointerInt(4),
		Monthly: proxmox.PointerInt(3),
		Weekly:  proxmox.PointerInt(2),
		Yearly:  proxmox.PointerInt(1),
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
