package proxmox

import (
	"testing"
)

func testData_RawConfigQemu_Disks_Ide_CdRom_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `none`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "none,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}})},
			{name: `passthrough`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "cdrom,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
			{name: `iso`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
					File:            "debian-11.0.0-amd64-netinst.iso",
					Storage:         "local",
					SizeInKibibytes: "377M"}}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Ide_CloudInit_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `file`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "Test:100/vm-100-cloudinit.raw,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})},
			{name: `lvm`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "Test:vm-100-cloudinit,media=cdrom"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{
					Format:  QemuDiskFormat_Raw,
					Storage: "Test"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Ide_Disk_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,aio=io_uring,backup=0,cache=writethrough,discard=on,iops_rd=12,iops_rd_max=13,iops_rd_max_length=4,iops_wr=15,iops_wr_max=14,iops_wr_max_length=5,mbps_rd=1.46,mbps_rd_max=3.57,mbps_wr=2.68,mbps_wr_max=4.55,replicate=0,serial=disk-9763,size=1032G,ssd=1,wwn=0x500F753600A987E1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					AsyncIO: QemuDiskAsyncIO_IOuring,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.57, Concurrent: 1.46},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55, Concurrent: 2.68}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 12},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14, BurstDuration: 5, Concurrent: 15}}},
					Cache:           QemuDiskCache_WriteThrough,
					Discard:         true,
					EmulateSSD:      true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              53,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 1082130432,
					Storage:         "test2",
					WorldWideName:   "0x500F753600A987E1",
					volumePath:      "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `ALL LinkedClone`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-53.qcow2,aio=io_uring,backup=0,cache=writethrough,discard=on,iops_rd=12,iops_rd_max=13,iops_rd_max_length=4,iops_wr=15,iops_wr_max=14,iops_wr_max_length=5,mbps_rd=1.46,mbps_rd_max=3.57,mbps_wr=2.68,mbps_wr_max=4.55,replicate=0,serial=disk-9763,size=1032G,ssd=1,wwn=0x500679CE00B1DAF4"}},
				output: testQemuBaseConfig_get(ConfigQemu{
					LinkedID: new(GuestID(110)),
					Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO: QemuDiskAsyncIO_IOuring,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.57, Concurrent: 1.46},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55, Concurrent: 2.68}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 12},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14, BurstDuration: 5, Concurrent: 15}}},
						Cache:           QemuDiskCache_WriteThrough,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              53,
						LinkedDiskId:    new(GuestID(1)),
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1082130432,
						Storage:         "test2",
						WorldWideName:   "0x500679CE00B1DAF4",
						volumePath:      "110/base-110-disk-1.qcow2/100/vm-100-disk-53.qcow2"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "test2:100/vm-100-disk-53.qcow2,aio=io_uring"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
					AsyncIO:    QemuDiskAsyncIO_IOuring,
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "test2:100/vm-100-disk-53.qcow2,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     false,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,cache=writethrough"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Cache:      QemuDiskCache_WriteThrough,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Discard:    true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "test2:100/vm-100-disk-53.qcow2,iops_rd=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 12}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "test2:100/vm-100-disk-53.qcow2,iops_rd_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "test2:100/vm-100-disk-53.qcow2,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,iops_wr=15"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 15}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,iops_wr_max=14"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "test2:100/vm-100-disk-53.qcow2,mbps_rd=1.46"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.46}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "test2:100/vm-100-disk-53.qcow2,mbps_rd_max=3.57"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.57}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,mbps_wr=2.68"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.68}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,mbps_wr_max=4.55"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "test2:100/vm-100-disk-53.qcow2,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  false,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "test2:100/vm-100-disk-53.qcow2,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Serial:     "disk-9763",
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              53,
					Replicate:       true,
					SizeInKibibytes: 1082130432,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              53,
					Replicate:       true,
					SizeInKibibytes: 1032,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              53,
					Replicate:       true,
					SizeInKibibytes: 1056768,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "test2:100/vm-100-disk-53.qcow2,size=2T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              53,
					Replicate:       true,
					SizeInKibibytes: 2147483648,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:     true,
					EmulateSSD: true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         53,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-53.qcow2"}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "test2:100/vm-100-disk-53.qcow2,wwn=0x500DB82100C6FA59"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
					Backup:        true,
					Format:        QemuDiskFormat_Qcow2,
					Id:            53,
					Replicate:     true,
					Storage:       "test2",
					WorldWideName: "0x500DB82100C6FA59",
					volumePath:    "100/vm-100-disk-53.qcow2"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_Ide_Passthrough_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `All`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,backup=0,cache=unsafe,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=1G,ssd=1,wwn=0x500CBE4300D978A6"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					AsyncIO: QemuDiskAsyncIO_Threads,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
					Cache:           QemuDiskCache_Unsafe,
					Discard:         true,
					EmulateSSD:      true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 1048576,
					WorldWideName:   "0x500CBE4300D978A6"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					AsyncIO:   QemuDiskAsyncIO_Threads,
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    false,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=unsafe"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Cache:     QemuDiskCache_Unsafe,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Discard:   true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: false}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true,
					Serial:    "disk-9763"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 10485760}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 10}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 10240}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 10737418240}}}}})},
			{name: `ssd`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:     true,
					EmulateSSD: true,
					File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:  true}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5005AC1200F643B8"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
					Backup:        true,
					File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:     true,
					WorldWideName: "0x5005AC1200F643B8"}}}}})}}
	}
}

func Test_RawConfigQemu_GetDisks_Ide_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_CdRom_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Ide_CdRom(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_CdRom_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Ide_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_CloudInit_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Ide_CloudInit(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_CloudInit_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Ide_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_Disk_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Ide_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_Disk_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_Ide_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_Passthrough_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_Ide_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_Ide_Passthrough_Get().Test(t)
}
