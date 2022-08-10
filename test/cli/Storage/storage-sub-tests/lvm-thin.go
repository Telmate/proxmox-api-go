package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

var LVMThinFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "lvm-thin",
	LVMThin: &proxmox.ConfigStorageLVMThin{
		VGname:   "pve",
		Thinpool: "data",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
	},
}

var LVMThinEmpty = proxmox.ConfigStorage{
	Type: "lvm-thin",
	LVMThin: &proxmox.ConfigStorageLVMThin{
		VGname:   "pve",
		Thinpool: "data",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
	},
}

func LVMThinGetFull(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(LVMThinFull)
	s.ID = name
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func LVMThinGetEmpty(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(LVMThinEmpty)
	s.ID = name
	s.Content.DiskImage = proxmox.PointerBool(false)
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}
