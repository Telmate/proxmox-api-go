package proxmox

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

func Test_IsoFile_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input IsoFile
		err   error
	}{
		// Valid
		{name: "Valid 00", input: IsoFile{File: "anything", Storage: "something"}},
		// Invalid
		{name: "Invalid 00", input: IsoFile{}, err: errors.New(Error_IsoFile_File)},
		{name: "Invalid 01", input: IsoFile{File: "anything"}, err: errors.New(Error_IsoFile_Storage)},
		{name: "Invalid 02", input: IsoFile{Storage: "something"}, err: errors.New(Error_IsoFile_File)},
		{name: "Invalid 03", input: IsoFile{SizeInKibibytes: "something"}, err: errors.New(Error_IsoFile_File)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuCdRom_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuCdRom
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuCdRom{}},
		{name: "Valid 01", input: QemuCdRom{Iso: &IsoFile{File: "anything", Storage: "Something"}}},
		{name: "Valid 02", input: QemuCdRom{Passthrough: true}},
		// Invalid
		{name: "Invalid 00", input: QemuCdRom{Iso: &IsoFile{}}, err: errors.New(Error_IsoFile_File)},
		{name: "Invalid 01", input: QemuCdRom{Iso: &IsoFile{File: "anything"}}, err: errors.New(Error_IsoFile_Storage)},
		{name: "Invalid 02", input: QemuCdRom{Iso: &IsoFile{Storage: "something"}}, err: errors.New(Error_IsoFile_File)},
		{name: "Invalid 03", input: QemuCdRom{Iso: &IsoFile{SizeInKibibytes: "something"}}, err: errors.New(Error_IsoFile_File)},
		{name: "Invalid 04", input: QemuCdRom{Iso: &IsoFile{File: "anything", Storage: "something"}, Passthrough: true}, err: errors.New(Error_QemuCdRom_MutuallyExclusive)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuCloudInitDisk_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuCloudInitDisk
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuCloudInitDisk{Storage: "anything", Format: QemuDiskFormat_Raw}},
		// Invalid
		{name: "Invalid 00", input: QemuCloudInitDisk{}, err: QemuDiskFormat("").Error()},
		{name: "Invalid 01", input: QemuCloudInitDisk{Format: QemuDiskFormat_Raw}, err: errors.New(Error_QemuCloudInitDisk_Storage)},
		{name: "Invalid 02", input: QemuCloudInitDisk{Storage: "anything", Format: QemuDiskFormat("")}, err: QemuDiskFormat("").Error()},
		{name: "Invalid 03", input: QemuCloudInitDisk{Storage: "anything", Format: QemuDiskFormat("invalid")}, err: QemuDiskFormat("").Error()},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_qemuDisk_formatDisk(t *testing.T) {
	uintPtr := func(i uint) *uint { return &i }
	type localInput struct {
		vmID           uint
		linkedVmId     uint
		currentStorage string
		currentFormat  QemuDiskFormat
		syntax         diskSyntaxEnum
		disk           qemuDisk
	}
	tests := []struct {
		name   string
		input  localInput
		output string
	}{
		{name: "linked file",
			input: localInput{
				vmID:           100,
				linkedVmId:     110,
				currentStorage: "storage",
				currentFormat:  QemuDiskFormat_Qcow2,
				syntax:         diskSyntaxFile,
				disk: qemuDisk{
					Id:           6,
					Storage:      "storage",
					Format:       QemuDiskFormat_Qcow2,
					LinkedDiskId: uintPtr(1),
				},
			},
			output: "storage:110/base-110-disk-1.qcow2/100/vm-100-disk-6.qcow2",
		},
		{name: "linked volume",
			input: localInput{
				vmID:           100,
				linkedVmId:     110,
				currentStorage: "storage",
				currentFormat:  QemuDiskFormat_Raw,
				syntax:         diskSyntaxVolume,
				disk: qemuDisk{
					Id:           12,
					Storage:      "storage",
					Format:       QemuDiskFormat_Qcow,
					LinkedDiskId: uintPtr(8),
				},
			},
			output: "storage:base-110-disk-8/vm-100-disk-12",
		},
		{name: "normal file",
			input: localInput{
				vmID:           100,
				currentStorage: "storage",
				currentFormat:  QemuDiskFormat_Qcow2,
				syntax:         diskSyntaxFile,
				disk: qemuDisk{
					Id:      9,
					Storage: "storage",
					Format:  QemuDiskFormat_Qcow2,
				},
			},
			output: "storage:100/vm-100-disk-9.qcow2",
		},
		{name: "normal volume",
			input: localInput{
				vmID:           100,
				currentStorage: "storage",
				currentFormat:  QemuDiskFormat_Qcow2,
				syntax:         diskSyntaxVolume,
				disk: qemuDisk{
					Id:      45,
					Storage: "storage",
					Format:  QemuDiskFormat_Qed,
				},
			},
			output: "storage:vm-100-disk-45",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			result := test.input.disk.formatDisk(test.input.vmID, test.input.linkedVmId, test.input.currentStorage, test.input.currentFormat, test.input.syntax)
			require.Equal(t, test.output, result, test.name)
		})
	}
}

func Test_qemuDisk_mapToStruct(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		linkedVmId uint
		output     uint
	}{
		{name: "Don't Update LinkedVmId file",
			input:      "storage:100/vm-100-disk-0.qcow2",
			linkedVmId: 110,
			output:     110,
		},
		{name: "Don't Update LinkedVmId volume",
			input:      "storage:vm-100-disk-45",
			linkedVmId: 110,
			output:     110,
		},
		{name: "Update LinkedVmId file",
			input:  "storage:110/base-110-disk-1.qcow2/100/vm-100-disk-0.qcow2",
			output: 110,
		},
		{name: "Update LinkedVmId volume",
			input:  "storage:base-110-disk-8/vm-100-disk-12",
			output: 110,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			linkedVmId := uint(test.linkedVmId)
			qemuDisk{}.mapToStruct(test.input, nil, &linkedVmId)
			require.Equal(t, test.output, linkedVmId, test.name)
		})
	}
}

func Test_qemuDisk_parseDisk(t *testing.T) {
	uintPtr := func(i uint) *uint { return &i }
	tests := []struct {
		name   string
		input  string
		output qemuDisk
	}{
		{name: "linked file",
			input: "storage:110/base-110-disk-1.qcow2/100/vm-100-disk-6.qcow2",
			output: qemuDisk{
				Id:           6,
				Storage:      "storage",
				Format:       QemuDiskFormat_Qcow2,
				LinkedDiskId: uintPtr(1),
				fileSyntax:   diskSyntaxFile,
			},
		},
		{name: "linked volume",
			input: "storage:base-110-disk-8/vm-100-disk-12",
			output: qemuDisk{
				Id:           12,
				Storage:      "storage",
				Format:       QemuDiskFormat_Raw,
				LinkedDiskId: uintPtr(8),
				fileSyntax:   diskSyntaxVolume,
			},
		},
		{name: "normal file",
			input: "storage:100/vm-100-disk-9.qcow2",
			output: qemuDisk{
				Id:         9,
				Storage:    "storage",
				Format:     QemuDiskFormat_Qcow2,
				fileSyntax: diskSyntaxFile,
			},
		},
		{name: "normal volume",
			input: "storage:vm-100-disk-45",
			output: qemuDisk{
				Id:         45,
				Storage:    "storage",
				Format:     QemuDiskFormat_Raw,
				fileSyntax: diskSyntaxVolume,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			linkedVmId := uint(0)
			disk := qemuDisk{}
			disk.Id, disk.Storage, disk.Format, disk.LinkedDiskId, disk.fileSyntax = qemuDisk{}.parseDisk(test.input, &linkedVmId)
			require.Equal(t, test.output, disk, test.name)
		})
	}
}

func Test_QemuDiskAsyncIO_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskAsyncIO
		err   error
	}{
		// Valid
		{name: "Valid 00", input: ""},
		{name: "Valid 01", input: QemuDiskAsyncIO_Native},
		{name: "Valid 02", input: QemuDiskAsyncIO_Threads},
		{name: "Valid 03", input: QemuDiskAsyncIO_IOuring},
		// Invalid
		{name: "Invalid 00", input: "bla", err: QemuDiskAsyncIO("").Error()},
		{name: "Invalid 01", input: "invalid value", err: QemuDiskAsyncIO("").Error()},
		{name: "Invalid 02", input: "!@#$", err: QemuDiskAsyncIO("").Error()},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidth_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidth
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidth{}},
		{name: "Valid 01", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}},
		{name: "Valid 02", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}},
		{name: "Valid 03", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0}}}},
		{name: "Valid 04", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 1}}}},
		{name: "Valid 05", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0}}}},
		{name: "Valid 06", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1}}}},
		{name: "Valid 07", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}},
		{name: "Valid 08", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0}}}},
		{name: "Valid 09", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 1}}}},
		{name: "Valid 10", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0}}}},
		{name: "Valid 11", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1}}}},
		{name: "Valid 12", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}},
		{name: "Valid 13", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}},
		{name: "Valid 14", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}}},
		{name: "Valid 15", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}}},
		{name: "Valid 16", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}}},
		{name: "Valid 17", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}}},
		{name: "Valid 18", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}},
		{name: "Valid 19", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}}},
		{name: "Valid 20", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}}},
		{name: "Valid 21", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}}},
		{name: "Valid 22", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
		{name: "Invalid 01", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
		{name: "Invalid 03", input: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
		{name: "Invalid 04", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
		{name: "Invalid 05", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
		{name: "Invalid 06", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
		{name: "Invalid 07", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthIops_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthIops
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthIops{}},
		{name: "Valid 01", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}},
		{name: "Valid 02", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{name: "Valid 03", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{name: "Valid 04", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}},
		{name: "Valid 05", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
		{name: "Valid 06", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}},
		{name: "Valid 07", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}},
		{name: "Valid 08", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{name: "Valid 09", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}},
		{name: "Valid 10", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
		{name: "Invalid 01", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
		{name: "Invalid 03", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthIopsLimit_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthIopsLimit
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthIopsLimit{}},
		{name: "Valid 01", input: QemuDiskBandwidthIopsLimit{Burst: 0}},
		{name: "Valid 02", input: QemuDiskBandwidthIopsLimit{Burst: 10}},
		{name: "Valid 03", input: QemuDiskBandwidthIopsLimit{Concurrent: 0}},
		{name: "Valid 04", input: QemuDiskBandwidthIopsLimit{Concurrent: 10}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthIopsLimit{Burst: 9}, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
		{name: "Invalid 01", input: QemuDiskBandwidthIopsLimit{Concurrent: 9}, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthIopsLimitBurst_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthIopsLimitBurst
		err   error
	}{
		// Valid
		{name: "Valid 03", input: 0},
		{name: "Valid 04", input: 10},
		// Invalid
		{name: "Invalid 01", input: 9, err: errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthIopsLimitConcurrent_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthIopsLimitConcurrent
		err   error
	}{
		// Valid
		{name: "Valid 03", input: 0},
		{name: "Valid 04", input: 10},
		// Invalid
		{name: "Invalid 01", input: 9, err: errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthMBps_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthMBps
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthMBps{}},
		{name: "Valid 01", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}},
		{name: "Valid 02", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0}}},
		{name: "Valid 03", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 1}}},
		{name: "Valid 04", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0}}},
		{name: "Valid 05", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1}}},
		{name: "Valid 06", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}},
		{name: "Valid 07", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0}}},
		{name: "Valid 08", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 1}}},
		{name: "Valid 09", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0}}},
		{name: "Valid 10", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1}}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
		{name: "Invalid 01", input: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
		{name: "Invalid 03", input: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthMBpsLimit_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthMBpsLimit
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthMBpsLimit{}},
		{name: "Valid 01", input: QemuDiskBandwidthMBpsLimit{Burst: 0}},
		{name: "Valid 02", input: QemuDiskBandwidthMBpsLimit{Burst: 1}},
		{name: "Valid 03", input: QemuDiskBandwidthMBpsLimit{Concurrent: 0}},
		{name: "Valid 04", input: QemuDiskBandwidthMBpsLimit{Concurrent: 1}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthMBpsLimit{Burst: 0.99}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
		{name: "Invalid 01", input: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthMBpsLimitBurst_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthMBpsLimitBurst
		err   error
	}{
		// Valid
		{name: "Valid 01", input: 0},
		{name: "Valid 02", input: 1},
		// Invalid
		{name: "Invalid 00", input: 0.99, err: errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskBandwidthMBpsLimitConcurrent_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthMBpsLimitConcurrent
		err   error
	}{
		// Valid
		{name: "Valid 01", input: 0},
		{name: "Valid 02", input: 1},
		// Invalid
		{name: "Invalid 00", input: 0.99, err: errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskCache_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskCache
		err   error
	}{
		// Valid
		{name: "Valid 00", input: ""},
		{name: "Valid 01", input: QemuDiskCache_None},
		{name: "Valid 02", input: QemuDiskCache_WriteThrough},
		{name: "Valid 03", input: QemuDiskCache_WriteBack},
		{name: "Valid 04", input: QemuDiskCache_Unsafe},
		{name: "Valid 05", input: QemuDiskCache_DirectSync},
		// Invalid
		{name: "Invalid 00", input: "bla", err: QemuDiskCache("").Error()},
		{name: "Invalid 01", input: "invalid value", err: QemuDiskCache("").Error()},
		{name: "Invalid 02", input: "!@#$", err: QemuDiskCache("").Error()},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskFormat_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input QemuDiskFormat
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskFormat_Cow},
		{name: "Valid 01", input: QemuDiskFormat_Cloop},
		{name: "Valid 02", input: QemuDiskFormat_Qcow},
		{name: "Valid 03", input: QemuDiskFormat_Qcow2},
		{name: "Valid 04", input: QemuDiskFormat_Qed},
		{name: "Valid 05", input: QemuDiskFormat_Vmdk},
		{name: "Valid 06", input: QemuDiskFormat_Raw},
		// Invalid
		{name: "Invalid 00", input: "bla", err: QemuDiskFormat("").Error()},
		{name: "Invalid 01", input: "invalid value", err: QemuDiskFormat("").Error()},
		{name: "Invalid 02", input: "!@#$", err: QemuDiskFormat("").Error()},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskId_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskId
		err   error
	}{
		// Invalid
		{name: "Invalid 00", input: "ide4", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 01", input: "ide01", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 02", input: "ide10", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 03", input: "sata6", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 04", input: "sata01", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 05", input: "sata10", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 06", input: "scsi31", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 07", input: "scsi01", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 08", input: "scsi100", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 09", input: "virtio16", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 10", input: "virtio01", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 11", input: "virtio100", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 12", input: "bla", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 13", input: "invalid value", err: errors.New(ERROR_QemuDiskId_Invalid)},
		{name: "Invalid 14", input: "!@#$", err: errors.New(ERROR_QemuDiskId_Invalid)},
		// Valid
		{name: "Valid 01", input: "ide0"},
		{name: "Valid 02", input: "ide1"},
		{name: "Valid 03", input: "ide2"},
		{name: "Valid 04", input: "ide3"},
		{name: "Valid 05", input: "sata0"},
		{name: "Valid 06", input: "sata1"},
		{name: "Valid 07", input: "sata2"},
		{name: "Valid 08", input: "sata3"},
		{name: "Valid 09", input: "sata4"},
		{name: "Valid 10", input: "sata5"},
		{name: "Valid 11", input: "scsi0"},
		{name: "Valid 12", input: "scsi1"},
		{name: "Valid 13", input: "scsi2"},
		{name: "Valid 14", input: "scsi3"},
		{name: "Valid 15", input: "scsi4"},
		{name: "Valid 16", input: "scsi5"},
		{name: "Valid 17", input: "scsi6"},
		{name: "Valid 18", input: "scsi7"},
		{name: "Valid 19", input: "scsi8"},
		{name: "Valid 20", input: "scsi9"},
		{name: "Valid 21", input: "scsi10"},
		{name: "Valid 22", input: "scsi11"},
		{name: "Valid 23", input: "scsi12"},
		{name: "Valid 24", input: "scsi13"},
		{name: "Valid 25", input: "scsi14"},
		{name: "Valid 26", input: "scsi15"},
		{name: "Valid 27", input: "scsi16"},
		{name: "Valid 28", input: "scsi17"},
		{name: "Valid 29", input: "scsi18"},
		{name: "Valid 30", input: "scsi19"},
		{name: "Valid 31", input: "scsi20"},
		{name: "Valid 32", input: "scsi21"},
		{name: "Valid 33", input: "scsi22"},
		{name: "Valid 34", input: "scsi23"},
		{name: "Valid 35", input: "scsi24"},
		{name: "Valid 36", input: "scsi25"},
		{name: "Valid 37", input: "scsi26"},
		{name: "Valid 38", input: "scsi27"},
		{name: "Valid 39", input: "scsi28"},
		{name: "Valid 40", input: "scsi29"},
		{name: "Valid 41", input: "scsi30"},
		{name: "Valid 42", input: "virtio0"},
		{name: "Valid 43", input: "virtio1"},
		{name: "Valid 44", input: "virtio2"},
		{name: "Valid 45", input: "virtio3"},
		{name: "Valid 46", input: "virtio4"},
		{name: "Valid 47", input: "virtio5"},
		{name: "Valid 48", input: "virtio6"},
		{name: "Valid 49", input: "virtio7"},
		{name: "Valid 50", input: "virtio8"},
		{name: "Valid 51", input: "virtio9"},
		{name: "Valid 52", input: "virtio10"},
		{name: "Valid 53", input: "virtio11"},
		{name: "Valid 54", input: "virtio12"},
		{name: "Valid 55", input: "virtio13"},
		{name: "Valid 56", input: "virtio14"},
		{name: "Valid 57", input: "virtio15"},
	}
	for _, test := range testData {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuDiskSerial_Validate(t *testing.T) {
	testRunes := struct {
		legal   []string
		illegal []string
	}{
		legal:   test_data_qemu.QemuDiskSerial_Legal(),
		illegal: test_data_qemu.QemuDiskSerial_Illegal(),
	}
	for i, test := range testRunes.legal {
		name := fmt.Sprintf("legal%03d", i)
		t.Run(name, func(*testing.T) {
			require.NoError(t, QemuDiskSerial(test).Validate(), name)
		})
	}
	for i, test := range testRunes.illegal {
		name := fmt.Sprintf("illegal%03d", i)
		t.Run(name, func(*testing.T) {
			require.Error(t, QemuDiskSerial(test).Validate(), name)
		})
	}
}

func Test_qemuDiskShort_mapToApiValues(t *testing.T) {
	format_Raw := QemuDiskFormat_Raw
	format_Qcow2 := QemuDiskFormat_Qcow2
	tests := []struct {
		name   string
		delete bool
		input  qemuDiskMove
		output map[string]interface{}
	}{
		{name: "ALL",
			delete: true,
			input: qemuDiskMove{
				Format:  &format_Raw,
				Id:      "ide0",
				Storage: "test0",
			},
			output: map[string]interface{}{
				"disk":    "ide0",
				"storage": "test0",
				"delete":  "1",
				"format":  "raw",
			},
		},
		{name: "Format nil",
			delete: true,
			input: qemuDiskMove{
				Id:      "sata4",
				Storage: "aaa0",
			},
			output: map[string]interface{}{
				"disk":    "sata4",
				"storage": "aaa0",
				"delete":  "1",
			},
		},
		{name: "Delete false",
			input: qemuDiskMove{
				Format:  &format_Qcow2,
				Id:      "scsi10",
				Storage: "test0",
			},
			output: map[string]interface{}{
				"format":  "qcow2",
				"disk":    "scsi10",
				"storage": "test0",
			},
		},
		{name: "MINIMAL",
			input: qemuDiskMove{
				Id:      "virtio13",
				Storage: "Test0",
			},
			output: map[string]interface{}{
				"disk":    "virtio13",
				"storage": "Test0",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.mapToApiValues(test.delete), test.name)
		})
	}
}

func Test_qemuDiskShort_Validate(t *testing.T) {
	format_Raw := QemuDiskFormat_Raw
	format_Invalid := QemuDiskFormat("invalid")
	format_Empty := QemuDiskFormat("")
	tests := []struct {
		name  string
		input qemuDiskMove
		err   error
	}{
		// TODO Add cases when Storage has a custom type
		// Invalid
		{name: "Invalid 00", input: qemuDiskMove{Format: &format_Invalid},
			err: QemuDiskFormat("").Error(),
		},
		{name: "Invalid 01", input: qemuDiskMove{Format: &format_Empty},
			err: QemuDiskFormat("").Error(),
		},
		{name: "Invalid 02", input: qemuDiskMove{Id: "invalid"},
			err: errors.New(ERROR_QemuDiskId_Invalid),
		},
		// Valid
		{name: "Valid 00", input: qemuDiskMove{
			Format: &format_Raw,
			Id:     "ide0",
		}},
		{name: "Valid 01", input: qemuDiskMove{
			Id: "ide0",
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			if test.err != nil {
				require.Equal(t, test.input.Validate(), test.err, test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_QemuStorages_cloudInitRemove(t *testing.T) {
	type testInput struct {
		currentStorages QemuStorages
		newStorages     QemuStorages
	}
	tests := []struct {
		name   string
		input  testInput
		output string
	}{
		{name: "Different Slot, Different Type",
			input: testInput{
				currentStorages: QemuStorages{
					Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
				newStorages: QemuStorages{
					Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
			},
			output: "sata2",
		},
		{name: "Different Slot, Same Type",
			input: testInput{
				currentStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
				newStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
			},
			output: "ide1",
		},
		{name: "Same Slot",
			input: testInput{
				currentStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
				newStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
			},
			output: "",
		},
		{name: "Same Slot, CloudInit Disk",
			input: testInput{
				currentStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
				newStorages: QemuStorages{
					Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
			},
			output: "",
		},
		{name: "Same Slot, Disk CloudInit",
			input: testInput{
				currentStorages: QemuStorages{
					Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
				newStorages: QemuStorages{
					Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}}}},
			},
			output: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.newStorages.cloudInitRemove(test.input.currentStorages), test.name)
		})
	}
}

func Test_QemuStorages_markDiskChanges(t *testing.T) {
	format_Raw := QemuDiskFormat_Raw
	tests := []struct {
		name            string
		storages        QemuStorages
		currentStorages QemuStorages
		output          *qemuUpdateChanges
	}{
		{name: "Disk CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: format_Raw, SizeInKibibytes: 100, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: format_Raw, SizeInKibibytes: 50, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: format_Raw, SizeInKibibytes: 33, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: format_Raw, SizeInKibibytes: 99, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{
				Move: []qemuDiskMove{
					{Format: &format_Raw, Id: "ide0", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "sata1", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "scsi2", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "virtio3", Storage: "NewStorage"},
				},
				Resize: []qemuDiskResize{
					{Id: "ide0", SizeInKibibytes: 100},
					{Id: "sata1", SizeInKibibytes: 50},
					{Id: "scsi2", SizeInKibibytes: 33},
					{Id: "virtio3", SizeInKibibytes: 99},
				},
			},
		},
		{name: "Disk NO CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: format_Raw, SizeInKibibytes: 100, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: format_Raw, SizeInKibibytes: 50, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: format_Raw, SizeInKibibytes: 33, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: format_Raw, SizeInKibibytes: 99, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk_X.Disk SAME",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk_X.Disk.Format CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{Move: []qemuDiskMove{
				{Format: &format_Raw, Id: "ide2", Storage: "Test"},
				{Format: &format_Raw, Id: "sata3", Storage: "Test"},
				{Format: &format_Raw, Id: "scsi4", Storage: "Test"},
				{Format: &format_Raw, Id: "virtio5", Storage: "Test"},
			}},
		},
		{name: "Disk.Size BIGGER",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 90, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 80, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 50, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 33, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{Resize: []qemuDiskResize{
				{Id: "ide3", SizeInKibibytes: 90},
				{Id: "sata4", SizeInKibibytes: 80},
				{Id: "scsi5", SizeInKibibytes: 50},
				{Id: "virtio6", SizeInKibibytes: 33},
			}},
		},
		{name: "Disk.Size SMALLER",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 1, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 10, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 20, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 31, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk.Storage CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, SizeInKibibytes: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{Move: []qemuDiskMove{
				{Id: "ide1", Storage: "NewStorage"},
				{Id: "sata0", Storage: "NewStorage"},
				{Id: "scsi7", Storage: "NewStorage"},
				{Id: "virtio8", Storage: "NewStorage"},
			}},
		},
		{name: "nil",
			output: &qemuUpdateChanges{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.storages.markDiskChanges(test.currentStorages), test.name)
		})
	}
}

func Test_QemuStorages_selectInitialResize(t *testing.T) {
	type testInput struct {
		currentStorages *QemuStorages
		newStorages     QemuStorages
	}
	tests := []struct {
		name   string
		input  testInput
		output []qemuDiskResize
	}{
		{name: "Disks Resize Down Gibibytes",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 10485761}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 1048577}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 70254593}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 872415233}}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 10485760}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 1048576}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 70254592}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 872415232}}},
			}},
		},
		{name: "Disks Resize Down Kibibytes",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 94384}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 75}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835893}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695729}}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 943}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 7}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide1", SizeInKibibytes: 943},
				{Id: "sata2", SizeInKibibytes: 7},
				{Id: "scsi3", SizeInKibibytes: 8584654835},
				{Id: "virtio4", SizeInKibibytes: 19695},
			},
		},
		{name: "Disks Resize Up",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 943}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 7}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695}}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 94384}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 75}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835893}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695729}}},
			}},
		},
		{name: "Disks Same",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 94384}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 75}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835893}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695729}}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 94384}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 75}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835893}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695729}}},
			}},
		},
		{name: "Don't resize cause whole x gibibyte",
			input: testInput{currentStorages: nil, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 1048576}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 10485760}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 47185920}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 480247808}}},
			}},
		},
		{name: "newStorages empty",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 94384}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 75}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 8584654835893}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 19695729}}},
			}, newStorages: QemuStorages{}},
		},
		{name: "No current disks 1 x kibibyte",
			input: testInput{currentStorages: nil, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 7867}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 985947483}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 577564}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 323}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide0", SizeInKibibytes: 7867},
				{Id: "sata1", SizeInKibibytes: 985947483},
				{Id: "scsi2", SizeInKibibytes: 577564},
				{Id: "virtio3", SizeInKibibytes: 323},
			},
		},
		{name: "No current disks 2 x kibibyte",
			input: testInput{currentStorages: &QemuStorages{}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 7867}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 985947483}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 577564}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 323}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide1", SizeInKibibytes: 7867},
				{Id: "sata2", SizeInKibibytes: 985947483},
				{Id: "scsi3", SizeInKibibytes: 577564},
				{Id: "virtio4", SizeInKibibytes: 323},
			},
		},
		{name: "No current disks 3 x kibibyte",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{},
				Sata:   &QemuSataDisks{},
				Scsi:   &QemuScsiDisks{},
				VirtIO: &QemuVirtIODisks{},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 7867}}},
				Sata:   &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 985947483}}},
				Scsi:   &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 577564}}},
				VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 323}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide2", SizeInKibibytes: 7867},
				{Id: "sata3", SizeInKibibytes: 985947483},
				{Id: "scsi4", SizeInKibibytes: 577564},
				{Id: "virtio5", SizeInKibibytes: 323},
			},
		},
		{name: "No current disks 4 x kibibyte",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 7867}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 985947483}}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 577564}}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 323}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide3", SizeInKibibytes: 7867},
				{Id: "sata4", SizeInKibibytes: 985947483},
				{Id: "scsi5", SizeInKibibytes: 577564},
				{Id: "virtio6", SizeInKibibytes: 323},
			},
		},
		{name: "No current disks 5 x kibibyte",
			input: testInput{currentStorages: &QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}},
			}, newStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 7867}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 985947483}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 577564}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 323}}},
			}},
			output: []qemuDiskResize{
				{Id: "ide0", SizeInKibibytes: 7867},
				{Id: "sata5", SizeInKibibytes: 985947483},
				{Id: "scsi6", SizeInKibibytes: 577564},
				{Id: "virtio7", SizeInKibibytes: 323},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.newStorages.selectInitialResize(test.input.currentStorages), test.name)
		})
	}
}

func Test_QemuWorldWideName_Validate(t *testing.T) {
	testRunes := struct {
		legal   []string
		illegal []string
	}{
		legal:   test_data_qemu.QemuWorldWideName_Legal(),
		illegal: test_data_qemu.QemuWorldWideName_Illegal(),
	}
	for _, test := range testRunes.legal {
		name := "legal:" + test
		t.Run(name, func(*testing.T) {
			require.NoError(t, QemuWorldWideName(test).Validate(), name)
		})
	}
	for _, test := range testRunes.illegal {
		name := "illegal:" + test
		t.Run(name, func(*testing.T) {
			require.Error(t, QemuWorldWideName(test).Validate(), name)
		})
	}
}
