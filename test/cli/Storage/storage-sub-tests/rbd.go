package storagesubtests

import (
	"testing"

	"github.com/perimeter-81/proxmox-api-go/proxmox"
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
	s := CloneJson(RBDFull)
	s.ID = name
	Get(s, name, t)
}

func RBDGetEmpty(name string, t *testing.T) {
	s := CloneJson(RBDEmpty)
	s.ID = name
	s.RBD.KRBD = false
	s.Content.DiskImage = proxmox.PointerBool(false)
	Get(s, name, t)
}
