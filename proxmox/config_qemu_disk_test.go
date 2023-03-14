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

func Test_QemuDiskAsyncIO_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskAsyncIO
		err   bool
	}{
		// Valid
		{input: QemuDiskAsyncIO_Native},
		{input: QemuDiskAsyncIO_Threads},
		{input: QemuDiskAsyncIO_IOuring},
		// Invalid
		{input: "bla", err: true},
		{input: "invalid value", err: true},
		{input: "!@#$", err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskBandwidth_Validate(t *testing.T) {
	float0 := float32(0)
	float0a := float32(0.99)
	float1 := float32(1)
	uint0 := uint(0)
	uint9 := uint(9)
	uint10 := uint(10)
	testData := []struct {
		input QemuDiskBandwidth
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidth{}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float1}}}},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float1}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint10}}}},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint10}}}},
		// Invalid
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float0}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float0a}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0a}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float0}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float0a}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0}}}, err: true},
		{input: QemuDiskBandwidth{Data: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0a}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint0}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint0}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint0}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint9}}}, err: true},
		{input: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint0}}}, err: true},
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
	float0 := float32(0)
	float0a := float32(0.99)
	float1 := float32(1)
	testData := []struct {
		input QemuDiskBandwidthData
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthData{}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float1}}},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float1}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float1}}},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float1}}},
		// Invalid
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float0}}, err: true},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Burst: &float0a}}, err: true},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0}}, err: true},
		{input: QemuDiskBandwidthData{ReadLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0a}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float0}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Burst: &float0a}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0}}, err: true},
		{input: QemuDiskBandwidthData{WriteLimit: QemuDiskBandwidthDataLimit{Concurrent: &float0a}}, err: true},
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
	float0 := float32(0)
	float0a := float32(0.99)
	float1 := float32(1)
	testData := []struct {
		input QemuDiskBandwidthDataLimit
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthDataLimit{}},
		{input: QemuDiskBandwidthDataLimit{Burst: &float1}},
		{input: QemuDiskBandwidthDataLimit{Concurrent: &float1}},
		// Invalid
		{input: QemuDiskBandwidthDataLimit{Burst: &float0}, err: true},
		{input: QemuDiskBandwidthDataLimit{Burst: &float0a}, err: true},
		{input: QemuDiskBandwidthDataLimit{Concurrent: &float0}, err: true},
		{input: QemuDiskBandwidthDataLimit{Concurrent: &float0a}, err: true},
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
	uint0 := uint(0)
	uint9 := uint(9)
	uint10 := uint(10)
	testData := []struct {
		input QemuDiskBandwidthIops
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthIops{}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint10}}},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint10}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint10}}},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint10}}},
		// Invalid
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint0}}, err: true},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: &uint9}}, err: true},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint0}}, err: true},
		{input: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint9}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint0}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: &uint9}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint0}}, err: true},
		{input: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: &uint9}}, err: true},
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
	uint0 := uint(0)
	uint9 := uint(9)
	uint10 := uint(10)
	testData := []struct {
		input QemuDiskBandwidthIopsLimit
		err   bool
	}{
		// Valid
		{input: QemuDiskBandwidthIopsLimit{}},
		{input: QemuDiskBandwidthIopsLimit{Burst: &uint10}},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: &uint10}},
		// Invalid
		{input: QemuDiskBandwidthIopsLimit{Burst: &uint0}, err: true},
		{input: QemuDiskBandwidthIopsLimit{Burst: &uint9}, err: true},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: &uint0}, err: true},
		{input: QemuDiskBandwidthIopsLimit{Concurrent: &uint9}, err: true},
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
		err   bool
	}{
		// Valid
		{input: QemuDiskCache_None},
		{input: QemuDiskCache_WriteThrough},
		{input: QemuDiskCache_WriteBack},
		{input: QemuDiskCache_Unsafe},
		{input: QemuDiskCache_DirectSync},
		// Invalid
		{input: "bla", err: true},
		{input: "invalid value", err: true},
		{input: "!@#$", err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
		} else {
			require.NoError(t, e.input.Validate())
		}
	}
}

func Test_QemuDiskFormat_Validate(t *testing.T) {
	testData := []struct {
		input QemuDiskFormat
		err   bool
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
		{input: "bla", err: true},
		{input: "invalid value", err: true},
		{input: "!@#$", err: true},
	}
	for _, e := range testData {
		if e.err {
			require.Error(t, e.input.Validate())
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
