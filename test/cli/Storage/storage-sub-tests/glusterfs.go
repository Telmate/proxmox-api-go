package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var GlusterfsFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "glusterfs",
	GlusterFS: &proxmox.ConfigStorageGlusterFS{
		Server1:       "10.20.1.1",
		Server2:       "10.20.1.2",
		Preallocation: util.Pointer("full"),
		Volume:        "test",
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:    util.Pointer(true),
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

var GlusterfsEmpty = proxmox.ConfigStorage{
	Type: "glusterfs",
	GlusterFS: &proxmox.ConfigStorageGlusterFS{
		Server1: "10.20.1.3",
		Volume:  "test",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: util.Pointer(true),
	},
}

func GlusterfsGetFull(name string, t *testing.T) {
	s := CloneJson(GlusterfsFull)
	s.ID = name
	Get(s, name, t)
}

func GlusterfsGetEmpty(name string, t *testing.T) {
	s := CloneJson(GlusterfsEmpty)
	s.ID = name
	s.GlusterFS.Preallocation = util.Pointer("metadata")
	s.Content.Backup = util.Pointer(false)
	s.Content.DiskImage = util.Pointer(false)
	s.Content.Snippets = util.Pointer(false)
	s.Content.Template = util.Pointer(false)
	Get(s, name, t)
}
