package proxmox

import (
	"testing"
)

func testData_RawConfigQemu_Disks_Scsi_CdRom_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `none`,
				input: rawConfigQemu{a: map[string]any{
					"scsi30": "none,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_30: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}})},
			{name: `passthrough`,
				input: rawConfigQemu{a: map[string]any{
					"scsi29": "cdrom,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_29: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
			{name: `iso`,
				input: rawConfigQemu{a: map[string]any{
					"scsi28": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_28: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
					File:            "debian-11.0.0-amd64-netinst.iso",
					Storage:         "local",
					SizeInKibibytes: "377M"}}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Scsi_CloudInit_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `file`,
				input: rawConfigQemu{a: map[string]any{
					"scsi0": "Test:100/vm-100-cloudinit.raw,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})},
			{name: `lvm`,
				input: rawConfigQemu{a: map[string]any{
					"scsi23": "Test:vm-100-cloudinit,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_23: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Scsi_Disk_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"scsi0": "test:100/vm-100-disk-2.qcow2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"scsi1": "test:100/vm-100-disk-2.qcow2,aio=threads,backup=0,cache=writeback,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,ssd=1,wwn=0x500AF18700E9CD25"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{
					AsyncIO: QemuDiskAsyncIO_Threads,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
					Cache:           QemuDiskCache_WriteBack,
					Discard:         true,
					EmulateSSD:      true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              2,
					IOThread:        true,
					ReadOnly:        true,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 33554432,
					Storage:         "test",
					WorldWideName:   "0x500AF18700E9CD25",
					volumePath:      "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `ALL LinkedClone`,
				input: rawConfigQemu{a: map[string]any{
					"scsi1": "test:110/base-110-disk-1.qcow2/100/vm-100-disk-2.qcow2,aio=threads,backup=0,cache=writeback,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,ssd=1,wwn=0x500879DC00F3BE6A"}},
				output: testQemuBaseConfig_get(ConfigQemu{
					LinkedID: new(GuestID(110)),
					Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_WriteBack,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              2,
						IOThread:        true,
						LinkedDiskId:    new(GuestID(1)),
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test",
						WorldWideName:   "0x500879DC00F3BE6A",
						volumePath:      "110/base-110-disk-1.qcow2/100/vm-100-disk-2.qcow2"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"scsi2": "test:100/vm-100-disk-2.qcow2,aio=threads"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{
					AsyncIO:    QemuDiskAsyncIO_Threads,
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"scsi3": "test:100/vm-100-disk-2.qcow2,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     false,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"scsi4": "test:100/vm-100-disk-2.qcow2,cache=writeback"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Cache:      QemuDiskCache_WriteBack,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"scsi5": "test:100/vm-100-disk-2.qcow2,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Discard:    true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi6": "test:100/vm-100-disk-2.qcow2,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi7": "test:100/vm-100-disk-2.qcow2,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"scsi7": "test:100/vm-100-disk-2.qcow2,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"scsi8": "test:100/vm-100-disk-2.qcow2,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi9": "test:100/vm-100-disk-2.qcow2,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"scsi9": "test:100/vm-100-disk-2.qcow2,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `iothread`,
				input: rawConfigQemu{a: map[string]any{
					"scsi10": "test:100/vm-100-disk-2.qcow2,iothread=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					IOThread:   true,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi11": "test:100/vm-100-disk-2.qcow2,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi12": "test:100/vm-100-disk-2.qcow2,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"scsi13": "test:100/vm-100-disk-2.qcow2,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi14": "test:100/vm-100-disk-2.qcow2,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"scsi15": "test:100/vm-100-disk-2.qcow2,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  false,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `ro`,
				input: rawConfigQemu{a: map[string]any{
					"scsi16": "test:100/vm-100-disk-2.qcow2,ro=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					ReadOnly:   true,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"scsi17": "test:100/vm-100-disk-2.qcow2,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Serial:     "disk-9763",
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "test:100/vm-100-disk-2.qcow2,size=32G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              2,
					Replicate:       true,
					SizeInKibibytes: 33554432,
					Storage:         "test",
					volumePath:      "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "test:100/vm-100-disk-2.qcow2,size=32K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              2,
					Replicate:       true,
					SizeInKibibytes: 32,
					Storage:         "test",
					volumePath:      "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "test:100/vm-100-disk-2.qcow2,size=32M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              2,
					Replicate:       true,
					SizeInKibibytes: 32768,
					Storage:         "test",
					volumePath:      "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "test:100/vm-100-disk-2.qcow2,size=4T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              2,
					Replicate:       true,
					SizeInKibibytes: 4294967296,
					Storage:         "test",
					volumePath:      "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi19": "test:100/vm-100-disk-2.qcow2,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:     true,
					EmulateSSD: true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         2,
					Replicate:  true,
					Storage:    "test",
					volumePath: "100/vm-100-disk-2.qcow2"}}}}})},
			{name: `WorldWideName`,
				input: rawConfigQemu{a: map[string]any{
					"scsi19": "test:100/vm-100-disk-2.qcow2,wwn=0x500E265400A1F3D7"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
					Backup:        true,
					Format:        QemuDiskFormat_Qcow2,
					Id:            2,
					Replicate:     true,
					Storage:       "test",
					WorldWideName: "0x500E265400A1F3D7",
					volumePath:    "100/vm-100-disk-2.qcow2"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Scsi_Passthrough_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"scsi0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `All`,
				input: rawConfigQemu{a: map[string]any{
					"scsi1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=1G,ssd=1,wwn=500CB15600D8FE32"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					AsyncIO: QemuDiskAsyncIO_Threads,
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
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					IOThread:        true,
					ReadOnly:        true,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 1048576,
					WorldWideName:   "500CB15600D8FE32"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"scsi2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					AsyncIO:   QemuDiskAsyncIO_Threads,
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"scsi3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    false,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"scsi4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=none"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Cache:     QemuDiskCache_None,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"scsi5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Discard:   true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi6": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"scsi7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"scsi8": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"scsi9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iothread`,
				input: rawConfigQemu{a: map[string]any{
					"scsi10": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iothread=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					IOThread:  true,
					Replicate: true}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi11": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi12": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"scsi13": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"scsi14": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"scsi15": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: false}}}}})},
			{name: `ro`,
				input: rawConfigQemu{a: map[string]any{
					"scsi16": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ro=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					ReadOnly:  true,
					Replicate: true}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"scsi17": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true,
					Serial:    "disk-9763"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1048576}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1024}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1073741824}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"scsi19": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:     true,
					EmulateSSD: true,
					File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:  true}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"scsi19": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5009A4FC00B7C613"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
					Backup:        true,
					File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:     true,
					WorldWideName: "0x5009A4FC00B7C613"}}}}})}}
	}
}

func Test_RawConfigQemu_GetDisks_Scsi_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_CdRom_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Scsi_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_CdRom_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Scsi_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_CloudInit_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Scsi_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_CloudInit_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Scsi_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_Disk_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Scsi_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_Disk_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Scsi_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_Passthrough_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Scsi_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Scsi_Passthrough_Get().Test(t)
}
