package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
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
	s := CloneJson(IscsiFull)
	s.ID = name
	Get(s, name, t)
}

func IscsiGetEmpty(name string, t *testing.T) {
	s := CloneJson(IscsiEmpty)
	s.ID = name
	s.Content.DiskImage = proxmox.PointerBool(false)
	Get(s, name, t)
}
