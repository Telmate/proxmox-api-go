package proxmox

import (
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/stretchr/testify/require"
)

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
