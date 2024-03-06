package storagesubtests

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

var SMBFull = proxmox.ConfigStorage{
	Enable: true,
	Nodes:  []string{"pve"},
	Type:   "smb",
	SMB: &proxmox.ConfigStorageSMB{
		Username:      "b.wayne",
		Share:         "NetworkShare",
		Preallocation: util.Pointer("full"),
		Domain:        "organization.pve",
		Server:        "10.20.1.1",
		Version:       util.Pointer("3"),
	},
	Content: &proxmox.ConfigStorageContent{
		Backup:    util.Pointer(true),
		Container: util.Pointer(true),
		DiskImage: util.Pointer(true),
		Iso:       util.Pointer(true),
		Snippets:  util.Pointer(true),
		Template:  util.Pointer(true),
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

var SMBEmpty = proxmox.ConfigStorage{
	Type: "smb",
	SMB: &proxmox.ConfigStorageSMB{
		Username: "b.wayne",
		Share:    "NetworkShare",
		Domain:   "organization.pve",
		Server:   "10.20.1.1",
	},
	Content: &proxmox.ConfigStorageContent{
		Snippets: util.Pointer(true),
	},
}

func SMBGetFull(name string, t *testing.T) {
	s := CloneJson(SMBFull)
	s.ID = name
	Get(s, name, t)
}

func SMBGetEmpty(name string, t *testing.T) {
	s := CloneJson(SMBEmpty)
	s.ID = name
	s.SMB.Preallocation = util.Pointer("metadata")
	s.Content.Backup = util.Pointer(false)
	s.Content.Container = util.Pointer(false)
	s.Content.DiskImage = util.Pointer(false)
	s.Content.Iso = util.Pointer(false)
	s.Content.Template = util.Pointer(false)
	Get(s, name, t)
}
