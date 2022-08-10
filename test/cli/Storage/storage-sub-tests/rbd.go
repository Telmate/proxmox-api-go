package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

var RBDFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "rbd",
	RBD: &proxmox.ConfigStorageRBD{
		Pool:      "test-pool",
		Monitors:  []string{"10.20.1.1", "10.20.1.2", "10.20.1.3"},
		Username:  "rbd-username",
		Namespace: "ceph-namespace",
		KRBD:      true,
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
		DiskImage: proxmox.PointerBool(true),
	},
}

var RBDEmpty = proxmox.ConfigStorage{
	Type: "rbd",
	RBD: &proxmox.ConfigStorageRBD{
		Pool:      "test-pool",
		Monitors:  []string{"10.20.1.3"},
		Username:  "rbd-username",
		Namespace: "ceph-namespace",
	},
	Content: &proxmox.ConfigStorageContent{
		Container: proxmox.PointerBool(true),
	},
}

func RBDGetFull(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(RBDFull)
	s.ID = name
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func RBDGetEmpty(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(RBDEmpty)
	s.ID = name
	s.RBD.KRBD = false
	s.Content.DiskImage = proxmox.PointerBool(false)
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}
