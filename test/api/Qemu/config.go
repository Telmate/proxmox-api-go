package api

import (
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func ReducedConfig(id pveSDK.GuestID, node pveSDK.NodeName, name pveSDK.GuestName) (set pveSDK.ConfigQemu, expected *pveSDK.ConfigQemu) {
	set = pveSDK.ConfigQemu{
		CPU:             &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description:     new(""),
		EfiDisk:         &pveSDK.EfiDisk{Delete: true},
		ID:              &id,
		Memory:          &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:            &name,
		Node:            &node,
		StartAtNodeBoot: new(false),
		Tablet:          new(false),
		Tags:            new(pveSDK.Tags{}),
	}
	expected = &pveSDK.ConfigQemu{
		Bios:            "seabios",
		Boot:            " ",
		CPU:             &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description:     new(""),
		Hotplug:         "network,disk,usb",
		ID:              &id,
		Memory:          &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:            &name,
		Node:            &node,
		Protection:      new(false),
		QemuDisks:       pveSDK.QemuDevices{},
		QemuKVM:         new(true),
		QemuOs:          "other",
		QemuUnusedDisks: pveSDK.QemuDevices{},
		QemuVga:         pveSDK.QemuDevice{},
		Scsihw:          "lsi",
		StartAtNodeBoot: new(false),
		Tablet:          new(false),
		Tags:            new(pveSDK.Tags),
	}
	return
}

func MinimumConfig(id pveSDK.GuestID, node pveSDK.NodeName, name pveSDK.GuestName) (set pveSDK.ConfigQemu, expected *pveSDK.ConfigQemu) {
	set = pveSDK.ConfigQemu{
		CPU:    &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		ID:     &id,
		Memory: &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:   &name,
		Node:   &node,
	}
	expected = &pveSDK.ConfigQemu{
		Bios:            "seabios",
		Boot:            " ",
		CPU:             &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description:     new(""),
		Hotplug:         "network,disk,usb",
		ID:              &id,
		Memory:          &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:            &name,
		Node:            &node,
		Protection:      new(false),
		QemuDisks:       pveSDK.QemuDevices{},
		QemuKVM:         new(true),
		QemuOs:          "other",
		QemuUnusedDisks: pveSDK.QemuDevices{},
		QemuVga:         pveSDK.QemuDevice{},
		Scsihw:          "lsi",
		StartAtNodeBoot: new(false),
		Tablet:          new(true),
		Tags:            new(pveSDK.Tags),
	}
	return
}

func MaximumConfig(id pveSDK.GuestID, node pveSDK.NodeName, name pveSDK.GuestName) (set pveSDK.ConfigQemu, expected *pveSDK.ConfigQemu) {
	set = pveSDK.ConfigQemu{
		CPU:         &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description: new(""),
		EfiDisk: &pveSDK.EfiDisk{
			Size:            1024,
			Format:          new(pveSDK.QemuDiskFormat("raw")),
			PreEnrolledKeys: new(true),
			Storage:         new(pveSDK.StorageName(test.GuestStorage)),
		},
		ID:              &id,
		Memory:          &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:            &name,
		Node:            &node,
		StartAtNodeBoot: new(true),
		Tablet:          new(true),
		Tags:            new(pveSDK.Tags{"Debian", "test", pveSDK.Tag(name)}),
	}
	expected = &pveSDK.ConfigQemu{
		Bios:        "seabios",
		Boot:        " ",
		CPU:         &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description: new(""),
		EfiDisk: &pveSDK.EfiDisk{
			Size:            1024,
			Format:          new(pveSDK.QemuDiskFormat("raw")),
			PreEnrolledKeys: new(true),
			Storage:         new(pveSDK.StorageName(test.GuestStorage)),
		},
		Hotplug:         "network,disk,usb",
		ID:              &id,
		Memory:          &pveSDK.QemuMemory{CapacityMiB: new(pveSDK.QemuMemoryCapacity(16))},
		Name:            &name,
		Node:            &node,
		Protection:      new(false),
		QemuDisks:       pveSDK.QemuDevices{},
		QemuKVM:         new(true),
		QemuOs:          "other",
		QemuUnusedDisks: pveSDK.QemuDevices{},
		QemuVga:         pveSDK.QemuDevice{},
		Scsihw:          "lsi",
		StartAtNodeBoot: new(true),
		Tablet:          new(true),
		Tags:            new(pveSDK.Tags{"debian", "test", pveSDK.Tag(strings.ToLower(name.String()))}),
	}
	return
}
