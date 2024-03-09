package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var DirectoryFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "directory",
	Directory: &proxmox.ConfigStorageDirectory{
		Path:          "/test",
		Preallocation: util.Pointer("full"),
		Shared:        true,
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:    util.Pointer(true),
		Container: util.Pointer(true),
		DiskImage: util.Pointer(true),
		Iso:       util.Pointer(true),
		Snippets:  util.Pointer(true),
		Template:  util.Pointer(true),
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

var DirectoryEmpty = proxmox.ConfigStorage{
	Type: "directory",
	Directory: &proxmox.ConfigStorageDirectory{
		Path: "/test",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: util.Pointer(true),
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
	s.Directory.Preallocation = util.Pointer("metadata")
	s.Content.Backup = util.Pointer(false)
	s.Content.Container = util.Pointer(false)
	s.Content.DiskImage = util.Pointer(false)
	s.Content.Snippets = util.Pointer(false)
	s.Content.Template = util.Pointer(false)
	Get(s, name, t)
}
