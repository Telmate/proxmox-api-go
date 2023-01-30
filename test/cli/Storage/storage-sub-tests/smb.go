package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var SMBFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "smb",
	SMB: &proxmox.ConfigStorageSMB{
		Username:      "b.wayne",
		Share:         "NetworkShare",
		Preallocation: proxmox.PointerString("full"),
		Domain:        "organization.pve",
		Server:        "10.20.1.1",
		Version:       proxmox.PointerString("3"),
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:    proxmox.PointerBool(true),
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
		Iso:       proxmox.PointerBool(true),
		Snippets:  proxmox.PointerBool(true),
		Template:  proxmox.PointerBool(true),
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

var SMBEmpty = proxmox.ConfigStorage{
	Type: "smb",
	SMB: &proxmox.ConfigStorageSMB{
		Username: "b.wayne",
		Share:    "NetworkShare",
		Domain:   "organization.pve",
		Server:   "10.20.1.1",
	},
	Content: &proxmox.ConfigStorageContent{
		Snippets: proxmox.PointerBool(true),
	},
}

func SMBGetFull(name string, t *testing.T) {
	s := CloneJson(SMBFull)
	s.ID = name
	Get(s, name, t)
}

func SMBGetEmpty(name string, t *testing.T) {
	s := CloneJson(SMBEmpty)
	s.ID = name
	s.SMB.Preallocation = proxmox.PointerString("metadata")
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.Container = proxmox.PointerBool(false)
	s.Content.DiskImage = proxmox.PointerBool(false)
	s.Content.Iso = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Get(s, name, t)
}
