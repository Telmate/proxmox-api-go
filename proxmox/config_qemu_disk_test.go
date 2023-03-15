package proxmox

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

func Test_IsoFile_Validate(t *testing.T) {
	testData := []struct {
		input IsoFile
		err   bool
	}{
		// Valid
		{input: IsoFile{File: "anything", Storage: "something"}},
		// Invalid
		{input: IsoFile{}, err: true},
		{input: IsoFile{File: "anything"}, err: true},
		{input: IsoFile{Storage: "something"}, err: true},
		{input: IsoFile{Size: "something"}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuCdRom_Validate(t *testing.T) {
	testData := []struct {
		input QemuCdRom
		err   bool
	}{
		// Valid
		{input: QemuCdRom{}},
		{input: QemuCdRom{Iso: &IsoFile{File: "anything", Storage: "Something"}}},
		{input: QemuCdRom{Passthrough: true}},
		// Invalid
		{input: QemuCdRom{Iso: &IsoFile{}}, err: true},
		{input: QemuCdRom{Iso: &IsoFile{File: "anything"}}, err: true},
		{input: QemuCdRom{Iso: &IsoFile{Storage: "something"}}, err: true},
		{input: QemuCdRom{Iso: &IsoFile{Size: "something"}}, err: true},
		{input: QemuCdRom{Iso: &IsoFile{File: "anything", Storage: "something"}, Passthrough: true}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuCloudInitDisk_Validate(t *testing.T) {
	formatRaw := QemuDiskFormat_Raw
	formatEmpty := QemuDiskFormat("")
	formatInvalid := QemuDiskFormat("invalid")
	testData := []struct {
		input QemuCloudInitDisk
		err   bool
	}{
		// Valid
		{input: QemuCloudInitDisk{Storage: "anything", Format: formatRaw}},
		// Invalid
		{input: QemuCloudInitDisk{}, err: true},
		{input: QemuCloudInitDisk{Format: formatRaw}, err: true},
		{input: QemuCloudInitDisk{Storage: "anything", Format: formatEmpty}, err: true},
		{input: QemuCloudInitDisk{Storage: "anything", Format: formatInvalid}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskAsyncIO_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskAsyncIO
		err   error
	}{
		// Valid
		{input: ""},
		{input: QemuDiskAsyncIO_Native},
		{input: QemuDiskAsyncIO_Threads},
		{input: QemuDiskAsyncIO_IOuring},
		// Invalid
		{input: "bla", err: QemuDiskAsyncIO("").Error()},
		{input: "invalid value", err: QemuDiskAsyncIO("").Error()},
		{input: "!@#$", err: QemuDiskAsyncIO("").Error()},
	}
	for _, e := range testData {
		if e.err != nil {
			require.Equal(t, e.input.Validate(), e.err)
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidth_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskBandwidth
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidth{}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}}},
		// Invalid
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidthData_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskBandwidthData
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthData{}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 1}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 1}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 1}}},
		// Invalid
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}, err: true},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: 0.99}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: 0.99}}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidthDataLimit_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskBandwidthDataLimit
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthDataLimit{}},
		{input: QemuDiskBandwidthDataLimit{Burst: 0}},
		{input: QemuDiskBandwidthDataLimit{Burst: 1}},
		{input: QemuDiskBandwidthDataLimit{Concurrent: 0}},
		{input: QemuDiskBandwidthDataLimit{Concurrent: 1}},
		// Invalid
		{input: QemuDiskBandwidthDataLimit{Burst: 0.99}, err: true},
		{input: QemuDiskBandwidthDataLimit{Concurrent: 0.99}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidthIops_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskBandwidthIops
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthIops{}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 0}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 10}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 0}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
		// Invalid
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: true},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 9}}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidthIopsLimit_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskBandwidthIopsLimit
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthIopsLimit{}},
		{input: QemuDiskBandwidthIopsLimit{Burst: 0}},
		{input: QemuDiskBandwidthIopsLimit{Burst: 10}},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: 0}},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: 10}},
		// Invalid
		{input: QemuDiskBandwidthIopsLimit{Burst: 9}, err: true},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: 9}, err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskCache_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskCache
		err   error
	}{
		// Valid
		{input: ""},
		{input: QemuDiskCache_None},
		{input: QemuDiskCache_WriteThrough},
		{input: QemuDiskCache_WriteBack},
		{input: QemuDiskCache_Unsafe},
		{input: QemuDiskCache_DirectSync},
		// Invalid
		{input: "bla", err: QemuDiskCache("").Error()},
		{input: "invalid value", err: QemuDiskCache("").Error()},
		{input: "!@#$", err: QemuDiskCache("").Error()},
	}
	for _, e := range testData {
		if e.err != nil {
			require.Equal(t, e.input.Validate(), e.err)
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskFormat_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskFormat
		err   error
	}{
		// Valid
		{input: QemuDiskFormat_Cow},
		{input: QemuDiskFormat_Cloop},
		{input: QemuDiskFormat_Qcow},
		{input: QemuDiskFormat_Qcow2},
		{input: QemuDiskFormat_Qed},
		{input: QemuDiskFormat_Vmdk},
		{input: QemuDiskFormat_Raw},
		// Invalid
		{input: "bla", err: QemuDiskFormat("").Error()},
		{input: "invalid value", err: QemuDiskFormat("").Error()},
		{input: "!@#$", err: QemuDiskFormat("").Error()},
	}
	for _, e := range testData {
		if e.err != nil {
			require.Equal(t, e.input.Validate(), e.err)
		} else {
			require.NoError(t, e.input.Validate())
		}
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
	for _, e := range testRunes.legal {
		require.NoError(t, QemuDiskSerial(e).Validate())
	}
	for _, e := range testRunes.illegal {
		require.Error(t, QemuDiskSerial(e).Validate())
	}
}
