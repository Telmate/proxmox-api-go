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
		{name: "Invalid 03", input: IsoFile{Size: "something"}, err: errors.New(Error_IsoFile_File)},
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
		{name: "Invalid 03", input: QemuCdRom{Iso: &IsoFile{Size: "something"}}, err: errors.New(Error_IsoFile_File)},
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
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: format_Raw, Size: 100, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: format_Raw, Size: 50, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: format_Raw, Size: 33, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: format_Raw, Size: 99, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{
				Move: []qemuDiskMove{
					{Format: &format_Raw, Id: "ide0", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "sata1", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "scsi2", Storage: "NewStorage"},
					{Format: &format_Raw, Id: "virtio3", Storage: "NewStorage"},
				},
				Resize: []qemuDiskResize{
					{Id: "ide0", SizeInGigaBytes: 100},
					{Id: "sata1", SizeInGigaBytes: 50},
					{Id: "scsi2", SizeInGigaBytes: 33},
					{Id: "virtio3", SizeInGigaBytes: 99},
				},
			},
		},
		{name: "Disk NO CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: format_Raw, Size: 100, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Format: format_Raw, Size: 50, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: format_Raw, Size: 33, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: format_Raw, Size: 99, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk_X.Disk SAME",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk_X.Disk.Format CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Vmdk, Size: 32, Storage: "Test"}}},
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
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 90, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 80, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 50, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 33, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{Resize: []qemuDiskResize{
				{Id: "ide3", SizeInGigaBytes: 90},
				{Id: "sata4", SizeInGigaBytes: 80},
				{Id: "scsi5", SizeInGigaBytes: 50},
				{Id: "virtio6", SizeInGigaBytes: 33},
			}},
		},
		{name: "Disk.Size SMALLER",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 1, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 10, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 20, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 31, Storage: "Test"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
			},
			output: &qemuUpdateChanges{},
		},
		{name: "Disk.Storage CHANGE",
			storages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "NewStorage"}}},
				Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "NewStorage"}}},
				Scsi:   &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "NewStorage"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "NewStorage"}}},
			},
			currentStorages: QemuStorages{
				Ide:    &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				Scsi:   &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
				VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Size: 32, Storage: "Test"}}},
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
