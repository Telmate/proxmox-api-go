package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
)

var CephfsFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "cephfs",
	CephFS: &proxmox.ConfigStorageCephFS{
		Monitors: []string{"10.20.1.1", "10.20.1.2"},
		Username: "test-ceph-user",
		FSname:   "test-fs-name",
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:   proxmox.PointerBool(true),
		Iso:      proxmox.PointerBool(true),
		Snippets: proxmox.PointerBool(true),
		Template: proxmox.PointerBool(true),
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

var CephfsEmpty = proxmox.ConfigStorage{
	Type: "cephfs",
	CephFS: &proxmox.ConfigStorageCephFS{
		Monitors: []string{"10.20.1.1"},
		Username: "test-ceph-user",
		FSname:   "test-fs-name",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: proxmox.PointerBool(true),
	},
}

func CephfsGetFull(name string, t *testing.T) {
	s := CloneJson(CephfsFull)
	s.ID = name
	Get(s, name, t)
}

func CephfsGetEmpty(name string, t *testing.T) {
	s := CloneJson(CephfsEmpty)
	s.ID = name
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.Snippets = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Get(s, name, t)
}
