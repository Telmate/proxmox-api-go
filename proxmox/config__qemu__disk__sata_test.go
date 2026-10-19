package proxmox

import (
	"testing"
)

func testData_RawConfigQemu_Disks_Sata_CdRom_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `none`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "none,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{CdRom: &QemuCdRom{}}}}})},
			{name: `passthrough`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "cdrom,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
			{name: `iso`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
					File:            "debian-11.0.0-amd64-netinst.iso",
					Storage:         "local",
					SizeInKibibytes: "377M"}}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Sata_CloudInit_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `file`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "Test:100/vm-100-cloudinit.raw,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})},
			{name: `lvm`,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "Test:vm-100-cloudinit,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Sata_Disk_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "test2:100/vm-100-disk-47.qcow2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "test2:100/vm-100-disk-47.qcow2,aio=native,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=32G,ssd=1,wwn=0x500DFA8900E3C641"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
					AsyncIO: QemuDiskAsyncIO_Native,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
					Cache:           QemuDiskCache_None,
					Discard:         true,
					EmulateSSD:      true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              47,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 33554432,
					Storage:         "test2",
					WorldWideName:   "0x500DFA8900E3C641",
					volumePath:      "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `ALL LinkedClone`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-47.qcow2,aio=native,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=32G,ssd=1,wwn=0x5003B97600A8F2D4"}},
				output: testQemuBaseConfig_get(ConfigQemu{
					LinkedID: new(GuestID(110)),
					Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_None,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              47,
						LinkedDiskId:    new(GuestID(1)),
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test2",
						WorldWideName:   "0x5003B97600A8F2D4",
						volumePath:      "110/base-110-disk-1.qcow2/100/vm-100-disk-47.qcow2"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "test2:100/vm-100-disk-47.qcow2,aio=native"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
					AsyncIO:    QemuDiskAsyncIO_Native,
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "test2:100/vm-100-disk-47.qcow2,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     false,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,cache=none"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Cache:      QemuDiskCache_None,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "test2:100/vm-100-disk-47.qcow2,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Discard:    true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "test2:100/vm-100-disk-47.qcow2,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "test2:100/vm-100-disk-47.qcow2,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "test2:100/vm-100-disk-47.qcow2,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "test2:100/vm-100-disk-47.qcow2,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "test2:100/vm-100-disk-47.qcow2,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "test2:100/vm-100-disk-47.qcow2,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "test2:100/vm-100-disk-47.qcow2,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "test2:100/vm-100-disk-47.qcow2,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup: true,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "test2:100/vm-100-disk-47.qcow2,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "test2:100/vm-100-disk-47.qcow2,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  false,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "test2:100/vm-100-disk-47.qcow2,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Serial:     "disk-9763",
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,size=32G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              47,
					Replicate:       true,
					SizeInKibibytes: 33554432,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,size=32K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              47,
					Replicate:       true,
					SizeInKibibytes: 32,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,size=32M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              47,
					Replicate:       true,
					SizeInKibibytes: 32768,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "test2:100/vm-100-disk-47.qcow2,size=3T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              47,
					Replicate:       true,
					SizeInKibibytes: 3221225472,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "test2:100/vm-100-disk-47.qcow2,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:     true,
					EmulateSSD: true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         47,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-47.qcow2"}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "test2:100/vm-100-disk-47.qcow2,wwn=0x5001E48A00D567C9"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
					Backup:        true,
					Format:        QemuDiskFormat_Qcow2,
					Id:            47,
					Replicate:     true,
					Storage:       "test2",
					WorldWideName: "0x5001E48A00D567C9",
					volumePath:    "100/vm-100-disk-47.qcow2"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Sata_Passthrough_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `All`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=5,iops_wr=11,iops_wr_max=13,iops_wr_max_length=4,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=1G,ssd=1,wwn=500E9FBC00F2A6D3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					AsyncIO: QemuDiskAsyncIO_IOuring,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 5, Concurrent: 10},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 11}}},
					Cache:           QemuDiskCache_DirectSync,
					Discard:         true,
					EmulateSSD:      true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 1048576,
					WorldWideName:   "500E9FBC00F2A6D3"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=io_uring"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					AsyncIO:   QemuDiskAsyncIO_IOuring,
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    false,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=directsync"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Cache:     QemuDiskCache_DirectSync,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Discard:   true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"sata0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: false}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true,
					Serial:    "disk-9763"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=3G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 3145728}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=5789K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 5789}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=45M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 46080}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=7T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 7516192768}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:     true,
					EmulateSSD: true,
					File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:  true}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5004D2EF00C8B57A"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
					Backup:        true,
					File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:     true,
					WorldWideName: "0x5004D2EF00C8B57A"}}}}})}}
	}
}

func Test_RawConfigQemu_GetDisks_Sata_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_CdRom_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Sata_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_CdRom_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Sata_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_CloudInit_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Sata_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_CloudInit_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Sata_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_Disk_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Sata_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_Disk_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Sata_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_Passthrough_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Sata_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Sata_Passthrough_Get().Test(t)
}
