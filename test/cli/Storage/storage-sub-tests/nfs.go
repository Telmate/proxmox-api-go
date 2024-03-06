package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var NFSFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "nfs",
	NFS: &proxmox.ConfigStorageNFS{
		Server:        "10.20.1.1",
		Export:        "/exports",
		Preallocation: util.Pointer("full"),
		Version:       util.Pointer("4"),
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

var NFSEmpty = proxmox.ConfigStorage{
	Type: "nfs",
	NFS: &proxmox.ConfigStorageNFS{
		Server: "10.20.1.1",
		Export: "/exports",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: util.Pointer(true),
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
	s.NFS.Preallocation = util.Pointer("metadata")
	s.Content.Backup = util.Pointer(false)
	s.Content.Container = util.Pointer(false)
	s.Content.Snippets = util.Pointer(false)
	s.Content.Iso = util.Pointer(false)
	s.Content.Template = util.Pointer(false)
	Get(s, name, t)
}
