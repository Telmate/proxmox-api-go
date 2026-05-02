package api

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func MinimumConfig(id pveSDK.GuestID, node pveSDK.NodeName, storage pveSDK.StorageName, privilidge *bool, name pveSDK.GuestName) (set, expected pveSDK.ConfigLXC) {
	set = pveSDK.ConfigLXC{
		ID: new(id),
		BootMount: &pveSDK.LxcBootMount{
			SizeInKibibytes: new(pveSDK.LxcMountSize(128 * 1024)),
			Storage:         new(string(storage))},
		CreateOptions: &pveSDK.LxcCreateOptions{
			OsTemplate: &pveSDK.LxcTemplate{
				Storage: test.TemplateStorage,
				File:    test.DownloadedLXCTemplate,
			}},
		Name:       new(name),
		Node:       new(node),
		Privileged: privilidge,
	}
	var expectedPrivileged bool = false
	if privilidge != nil {
		expectedPrivileged = *privilidge
	}
	var expectedQuota *bool
	if expectedPrivileged {
		expectedQuota = new(false)
	}
	expected = pveSDK.ConfigLXC{
		Architecture: "amd64",
		BootMount: &pveSDK.LxcBootMount{
			ACL:             new(pveSDK.TriBoolNone),
			Replicate:       new(true),
			SizeInKibibytes: new(pveSDK.LxcMountSize(131072)),
			Storage:         new("local-zfs"),
			Quota:           expectedQuota,
		},
		ID:              new(id),
		Memory:          new(pveSDK.LxcMemory(512)),
		Name:            new(name),
		Networks:        pveSDK.LxcNetworks{},
		Node:            new(node),
		OperatingSystem: "alpine",
		Privileged:      new(expectedPrivileged),
		Protection:      new(false),
		StartAtNodeBoot: new(false),
		Swap:            new(pveSDK.LxcSwap(512)),
		Tags:            new(pveSDK.Tags),
	}
	return
}
