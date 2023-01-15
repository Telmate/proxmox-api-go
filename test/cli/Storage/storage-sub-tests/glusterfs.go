package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var GlusterfsFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "glusterfs",
	GlusterFS: &proxmox.ConfigStorageGlusterFS{
		Server1:       "10.20.1.1",
		Server2:       "10.20.1.2",
		Preallocation: proxmox.PointerString("full"),
		Volume:        "test",
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:    proxmox.PointerBool(true),
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

var GlusterfsEmpty = proxmox.ConfigStorage{
	Type: "glusterfs",
	GlusterFS: &proxmox.ConfigStorageGlusterFS{
		Server1: "10.20.1.3",
		Volume:  "test",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: proxmox.PointerBool(true),
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
	s.GlusterFS.Preallocation = proxmox.PointerString("metadata")
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.DiskImage = proxmox.PointerBool(false)
	s.Content.Snippets = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Get(s, name, t)
}
