package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

var ZFSFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "zfs",
	ZFS: &proxmox.ConfigStorageZFS{
		Pool:          "test-pool",
		Blocksize:     proxmox.PointerString("4k"),
		Thinprovision: true,
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
	},
}

var ZFSEmpty = proxmox.ConfigStorage{
	Type: "zfs",
	ZFS: &proxmox.ConfigStorageZFS{
		Pool: "test-pool",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

func ZFSGetFull(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(ZFSFull)
	s.ID = name
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func ZFSGetEmpty(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(ZFSEmpty)
	s.ID = name
	s.ZFS.Blocksize = proxmox.PointerString("8k")
	s.Content.Container = proxmox.PointerBool(false)
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}
