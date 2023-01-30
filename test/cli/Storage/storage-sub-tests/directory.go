package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var DirectoryFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "directory",
	Directory: &proxmox.ConfigStorageDirectory{
		Path:          "/test",
		Preallocation: proxmox.PointerString("full"),
		Shared:        true,
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

var DirectoryEmpty = proxmox.ConfigStorage{
	Type: "directory",
	Directory: &proxmox.ConfigStorageDirectory{
		Path: "/test",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: proxmox.PointerBool(true),
	},
}

func DirectoryGetFull(name string, t *testing.T) {
	s := CloneJson(DirectoryFull)
	s.ID = name
	Get(s, name, t)
}

func DirectoryGetEmpty(name string, t *testing.T) {
	s := CloneJson(DirectoryEmpty)
	s.ID = name
	s.Directory.Preallocation = proxmox.PointerString("metadata")
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.Container = proxmox.PointerBool(false)
	s.Content.DiskImage = proxmox.PointerBool(false)
	s.Content.Snippets = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Get(s, name, t)
}
