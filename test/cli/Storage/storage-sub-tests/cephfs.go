package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
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
	cliTest.SetEnvironmentVariables()
	s := CloneJson(CephfsFull)
	s.ID = name
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func CephfsGetEmpty(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(CephfsEmpty)
	s.ID = name
	s.Content.Backup = proxmox.PointerBool(false)
	s.Content.Snippets = proxmox.PointerBool(false)
	s.Content.Template = proxmox.PointerBool(false)
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}
