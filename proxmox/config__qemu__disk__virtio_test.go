package proxmox

import (
	"testing"
)

func testData_RawConfigQemu_Disks_VirtIO_Disk_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"virtio0": "test2:100/vm-100-disk-31.qcow2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"virtio1": "test2:100/vm-100-disk-31.qcow2,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=2,iops_wr=11,iops_wr_max=13,iops_wr_max_length=3,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,wwn=0x50015B3900F8EAD2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					AsyncIO: QemuDiskAsyncIO_IOuring,
					Backup:  false,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
							WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
						Iops: QemuDiskBandwidthIops{
							ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 2, Concurrent: 10},
							WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 3, Concurrent: 11}}},
					Cache:           QemuDiskCache_DirectSync,
					Discard:         true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              31,
					IOThread:        true,
					ReadOnly:        true,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 33554432,
					Storage:         "test2",
					WorldWideName:   "0x50015B3900F8EAD2",
					volumePath:      "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `ALL LinkedClone`,
				input: rawConfigQemu{a: map[string]any{
					"virtio1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-31.qcow2,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=2,iops_wr=11,iops_wr_max=13,iops_wr_max_length=3,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,wwn=0x500FA2D000C69587"}},
				output: testQemuBaseConfig_get(ConfigQemu{
					LinkedID: new(GuestID(110)),
					Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO: QemuDiskAsyncIO_IOuring,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 2, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 3, Concurrent: 11}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              31,
						IOThread:        true,
						LinkedDiskId:    new(GuestID(1)),
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test2",
						WorldWideName:   "0x500FA2D000C69587",
						volumePath:      "110/base-110-disk-1.qcow2/100/vm-100-disk-31.qcow2"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,aio=io_uring"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					AsyncIO:    QemuDiskAsyncIO_IOuring,
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"virtio3": "test2:100/vm-100-disk-31.qcow2,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     false,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"virtio4": "test2:100/vm-100-disk-31.qcow2,cache=directsync"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Cache:      QemuDiskCache_DirectSync,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"virtio5": "test2:100/vm-100-disk-31.qcow2,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Discard:    true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"virtio6": "test2:100/vm-100-disk-31.qcow2,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio7": "test2:100/vm-100-disk-31.qcow2,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"virtio7": "test2:100/vm-100-disk-31.qcow2,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"virtio8": "test2:100/vm-100-disk-31.qcow2,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio9": "test2:100/vm-100-disk-31.qcow2,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"virtio9": "test2:100/vm-100-disk-31.qcow2,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `iothread`,
				input: rawConfigQemu{a: map[string]any{
					"virtio10": "test2:100/vm-100-disk-31.qcow2,iothread=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					IOThread:   true,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"virtio11": "test2:100/vm-100-disk-31.qcow2,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup: true,
					Bandwidth: QemuDiskBandwidth{
						MBps: QemuDiskBandwidthMBps{
							ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio12": "test2:100/vm-100-disk-31.qcow2,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"virtio13": "test2:100/vm-100-disk-31.qcow2,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio14": "test2:100/vm-100-disk-31.qcow2,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Bandwidth:  QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"virtio15": "test2:100/vm-100-disk-31.qcow2,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  false,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `ro`,
				input: rawConfigQemu{a: map[string]any{
					"virtio0": "test2:100/vm-100-disk-31.qcow2,ro=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					ReadOnly:   true,
					Replicate:  true,
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"virtio1": "test2:100/vm-100-disk-31.qcow2,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:     true,
					Format:     QemuDiskFormat_Qcow2,
					Id:         31,
					Replicate:  true,
					Serial:     "disk-9763",
					Storage:    "test2",
					volumePath: "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              31,
					Replicate:       true,
					SizeInKibibytes: 33554432,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              31,
					Replicate:       true,
					SizeInKibibytes: 32,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              31,
					Replicate:       true,
					SizeInKibibytes: 32768,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,size=5T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:          true,
					Format:          QemuDiskFormat_Qcow2,
					Id:              31,
					Replicate:       true,
					SizeInKibibytes: 5368709120,
					Storage:         "test2",
					volumePath:      "100/vm-100-disk-31.qcow2"}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "test2:100/vm-100-disk-31.qcow2,wwn=0x500D3FAB00B4E672"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
					Backup:        true,
					Format:        QemuDiskFormat_Qcow2,
					Id:            31,
					Replicate:     true,
					Storage:       "test2",
					WorldWideName: "0x500D3FAB00B4E672",
					volumePath:    "100/vm-100-disk-31.qcow2"}}}}})}}
	}
}

func testData_RawConfigQemu_Disks_VirtIO_Passthrough_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: ``,
				input: rawConfigQemu{a: map[string]any{
					"virtio0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `ALL`,
				input: rawConfigQemu{a: map[string]any{
					"virtio1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=native,backup=0,cache=unsafe,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=1G,wwn=0x500B6ED600F1C945"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					AsyncIO: QemuDiskAsyncIO_Native,
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
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					IOThread:        true,
					ReadOnly:        true,
					Replicate:       false,
					Serial:          "disk-9763",
					SizeInKibibytes: 1048576,
					WorldWideName:   "0x500B6ED600F1C945"}}}}})},
			{name: `aio`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=native"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					AsyncIO:   QemuDiskAsyncIO_Native,
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `backup`,
				input: rawConfigQemu{a: map[string]any{
					"virtio3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    false,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `cache`,
				input: rawConfigQemu{a: map[string]any{
					"virtio4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=unsafe"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Cache:     QemuDiskCache_Unsafe,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `discard`,
				input: rawConfigQemu{a: map[string]any{
					"virtio5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Discard:   true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd`,
				input: rawConfigQemu{a: map[string]any{
					"virtio6": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_rd_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"virtio7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr`,
				input: rawConfigQemu{a: map[string]any{
					"virtio8": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iops_wr_max_length`,
				input: rawConfigQemu{a: map[string]any{
					"virtio9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `iothread`,
				input: rawConfigQemu{a: map[string]any{
					"virtio10": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iothread=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					IOThread:  true,
					Replicate: true}}}}})},
			{name: `mbps_rd`,
				input: rawConfigQemu{a: map[string]any{
					"virtio11": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_rd_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio12": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr`,
				input: rawConfigQemu{a: map[string]any{
					"virtio13": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `mbps_wr_max`,
				input: rawConfigQemu{a: map[string]any{
					"virtio14": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true}}}}})},
			{name: `replicate`,
				input: rawConfigQemu{a: map[string]any{
					"virtio15": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: false}}}}})},
			{name: `ro`,
				input: rawConfigQemu{a: map[string]any{
					"virtio0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ro=1"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					ReadOnly:  true,
					Replicate: true}}}}})},
			{name: `serial`,
				input: rawConfigQemu{a: map[string]any{
					"virtio1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:    true,
					File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate: true,
					Serial:    "disk-9763"}}}}})},
			{name: `size G`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1G"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1048576}}}}})},
			{name: `size K`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1K"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1}}}}})},
			{name: `size M`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1M"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1024}}}}})},
			{name: `size T`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1T"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:          true,
					File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:       true,
					SizeInKibibytes: 1073741824}}}}})},
			{name: `wwn`,
				input: rawConfigQemu{a: map[string]any{
					"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5008FA6500D9C8B3"}},
				output: testQemuBaseConfig_get(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
					Backup:        true,
					File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
					Replicate:     true,
					WorldWideName: "0x5008FA6500D9C8B3"}}}}})}}
	}
}

func Test_RawConfigQemu_GetDisks_VirtIO_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_VirtIO_Disk_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_VirtIO_Disk(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_VirtIO_Disk_Get().Test(t)
}

func Test_RawConfigQemu_GetDisks_VirtIO_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_VirtIO_Passthrough_Get().Inject(t, getQemuDisksInject)
}

func Test_RawConfigQemu_Disks_Get_VirtIO_Passthrough(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Disks_VirtIO_Passthrough_Get().Test(t)
}
