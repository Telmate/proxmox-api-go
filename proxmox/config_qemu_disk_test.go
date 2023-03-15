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
		{name: "Valid 01", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{}}},
		{name: "Valid 02", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}}},
		{name: "Valid 03", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0}}}},
		{name: "Valid 04", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 1}}}},
		{name: "Valid 05", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}}},
		{name: "Valid 06", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}}},
		{name: "Valid 07", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}}},
		{name: "Valid 08", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0}}}},
		{name: "Valid 09", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 1}}}},
		{name: "Valid 10", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}}},
		{name: "Valid 11", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}}},
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
		{name: "Invalid 00", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Burst)},
		{name: "Invalid 01", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Burst)},
		{name: "Invalid 03", input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)},
		{name: "Invalid 04", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)},
		{name: "Invalid 05", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)},
		{name: "Invalid 06", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)},
		{name: "Invalid 07", input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)},
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

func Test_QemuDiskBandwidthData_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthData
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthData{}},
		{name: "Valid 01", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}},
		{name: "Valid 02", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0}}},
		{name: "Valid 03", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 1}}},
		{name: "Valid 04", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}},
		{name: "Valid 05", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}},
		{name: "Valid 06", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}},
		{name: "Valid 07", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0}}},
		{name: "Valid 08", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 1}}},
		{name: "Valid 09", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}},
		{name: "Valid 10", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Burst)},
		{name: "Invalid 01", input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Burst)},
		{name: "Invalid 03", input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)},
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

func Test_QemuDiskBandwidthDataLimit_Validate(t *testing.T) {
	testData := []struct {
		name  string
		input QemuDiskBandwidthDataLimit
		err   error
	}{
		// Valid
		{name: "Valid 00", input: QemuDiskBandwidthDataLimit{}},
		{name: "Valid 01", input: QemuDiskBandwidthDataLimit{Burst: 0}},
		{name: "Valid 02", input: QemuDiskBandwidthDataLimit{Burst: 1}},
		{name: "Valid 03", input: QemuDiskBandwidthDataLimit{Concurrent: 0}},
		{name: "Valid 04", input: QemuDiskBandwidthDataLimit{Concurrent: 1}},
		// Invalid
		{name: "Invalid 00", input: QemuDiskBandwidthDataLimit{Burst: 0.99}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Burst)},
		{name: "Invalid 01", input: QemuDiskBandwidthDataLimit{Concurrent: 0.99}, err: errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)},
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
		{name: "Invalid 00", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)},
		{name: "Invalid 01", input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)},
		{name: "Invalid 02", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)},
		{name: "Invalid 03", input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)},
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
		{name: "Invalid 00", input: QemuDiskBandwidthIopsLimit{Burst: 9}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)},
		{name: "Invalid 01", input: QemuDiskBandwidthIopsLimit{Concurrent: 9}, err: errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)},
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
	testData := []struct {
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
