package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var NFSFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "nfs",
	NFS: &proxmox.ConfigStorageNFS{
		Server:        "10.20.1.1",
		Export:        "/exports",
		Preallocation: proxmox.PointerString("full"),
		Version:       proxmox.PointerString("4"),
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

var NFSEmpty = proxmox.ConfigStorage{
	Type: "nfs",
	NFS: &proxmox.ConfigStorageNFS{
		Server: "10.20.1.1",
		Export: "/exports",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

func NFSGetFull(name string, t *testing.T) {
	s := CloneJson(NFSFull)
	s.ID = name
	Get(s, name, t)
}

func NFSGetEmpty(name string, t *testing.T) {
	s := CloneJson(NFSEmpty)
	s.ID = name
	s.NFS.Preallocation = proxmox.PointerString("metadata")
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.Container = proxmox.PointerBool(false)
	s.Content.Snippets = proxmox.PointerBool(false)
	s.Content.Iso = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Get(s, name, t)
}
