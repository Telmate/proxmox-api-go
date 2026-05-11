package api

import (
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/body"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func ReducedConfig(id pveSDK.GuestID, node pveSDK.NodeName, name pveSDK.GuestName) (set pveSDK.ConfigQemu, expected *pveSDK.ConfigQemu) {
	set = pveSDK.ConfigQemu{
		CPU:         &pveSDK.QemuCPU{Cores: new(pveSDK.QemuCpuCores(1))},
		Description: new(""),
		Disks: &pveSDK.QemuStorages{
			Ide: &pveSDK.QemuIdeDisks{
				Disk_0: &pveSDK.QemuIdeStorage{Delete: true},
				Disk_1: &pveSDK.QemuIdeStorage{Delete: true},
			},
			Sata: &pveSDK.QemuSataDisks{
				Disk_0: &pveSDK.QemuSataStorage{Delete: true},
				Disk_1: &pveSDK.QemuSataStorage{Delete: true},
			},
			Scsi: &pveSDK.QemuScsiDisks{
				Disk_0: &pveSDK.QemuScsiStorage{Delete: true},
				Disk_1: &pveSDK.QemuScsiStorage{Delete: true},
			},
			VirtIO: &pveSDK.QemuVirtIODisks{
				Disk_0: &pveSDK.QemuVirtIOStorage{Delete: true},
				Disk_1: &pveSDK.QemuVirtIOStorage{Delete: true},
			},
		},
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
		Description: new(body.Alphanumeric + body.Symbols),
		Disks: &pveSDK.QemuStorages{
			Ide: &pveSDK.QemuIdeDisks{
				Disk_0: &pveSDK.QemuIdeStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuIdeStorage{Disk: &pveSDK.QemuIdeDisk{
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			Sata: &pveSDK.QemuSataDisks{
				Disk_0: &pveSDK.QemuSataStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuSataStorage{Disk: &pveSDK.QemuSataDisk{
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			Scsi: &pveSDK.QemuScsiDisks{
				Disk_0: &pveSDK.QemuScsiStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuScsiStorage{Disk: &pveSDK.QemuScsiDisk{
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			VirtIO: &pveSDK.QemuVirtIODisks{
				Disk_0: &pveSDK.QemuVirtIOStorage{Disk: &pveSDK.QemuVirtIODisk{
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
		},
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
		Description: new(body.Alphanumeric + body.Symbols),
		Disks: &pveSDK.QemuStorages{
			Ide: &pveSDK.QemuIdeDisks{
				Disk_0: &pveSDK.QemuIdeStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuIdeStorage{Disk: &pveSDK.QemuIdeDisk{
					Id:              1,
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			Sata: &pveSDK.QemuSataDisks{
				Disk_0: &pveSDK.QemuSataStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuSataStorage{Disk: &pveSDK.QemuSataDisk{
					Id:              2,
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			Scsi: &pveSDK.QemuScsiDisks{
				Disk_0: &pveSDK.QemuScsiStorage{CdRom: &pveSDK.QemuCdRom{}},
				Disk_1: &pveSDK.QemuScsiStorage{Disk: &pveSDK.QemuScsiDisk{
					Id:              3,
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
			VirtIO: &pveSDK.QemuVirtIODisks{
				Disk_0: &pveSDK.QemuVirtIOStorage{Disk: &pveSDK.QemuVirtIODisk{
					Id:              4,
					Format:          pveSDK.QemuDiskFormat_Raw,
					SizeInKibibytes: 12345,
					Storage:         test.GuestStorage,
				}}},
		},
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
