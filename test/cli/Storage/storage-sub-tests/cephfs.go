package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
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
		Backup:   util.Pointer(true),
		Iso:      util.Pointer(true),
		Snippets: util.Pointer(true),
		Template: util.Pointer(true),
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

var CephfsEmpty = proxmox.ConfigStorage{
	Type: "cephfs",
	CephFS: &proxmox.ConfigStorageCephFS{
		Monitors: []string{"10.20.1.1"},
		Username: "test-ceph-user",
		FSname:   "test-fs-name",
	},
	Content: &proxmox.ConfigStorageContent{
		Iso: util.Pointer(true),
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
	s.Content.Backup = util.Pointer(false)
	s.Content.Snippets = util.Pointer(false)
	s.Content.Template = util.Pointer(false)
	Get(s, name, t)
}
