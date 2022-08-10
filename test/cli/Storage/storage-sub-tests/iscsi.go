package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

var IscsiFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "iscsi",
	ISCSI: &proxmox.ConfigStorageISCSI{
		Portal: "10.20.1.1",
		Target: "target-volume",
	},
	Content: &proxmox.ConfigStorageContent{
		DiskImage: proxmox.PointerBool(true),
	},
}

var IscsiEmpty = proxmox.ConfigStorage{
	Type: "iscsi",
	ISCSI: &proxmox.ConfigStorageISCSI{
		Portal: "10.20.1.1",
		Target: "target-volume",
	},
	Content: &proxmox.ConfigStorageContent{},
}

func IscsiGetFull(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(IscsiFull)
	s.ID = name
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}

func IscsiGetEmpty(name string, t *testing.T) {
	cliTest.SetEnvironmentVariables()
	s := CloneJson(IscsiEmpty)
	s.ID = name
	s.Content.DiskImage = proxmox.PointerBool(false)
	Test := cliTest.Test{
		OutputJson: InlineMarshal(s),
		Args:       []string{"-i", "get", "storage", name},
	}
	Test.StandardTest(t)
}
