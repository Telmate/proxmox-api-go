package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/array"
)

type BootOrder []BootDevice

const (
	BootOrder_Error_Duplicate string = "duplicate entry in boot order"
)

func (config BootOrder) mapToApi() string { return array.StringSV(config, semicolon) }

func (config BootOrder) mapToApiUpdate(current BootOrder, b *strings.Builder) {
	v := config.mapToApi()
	if v != current.mapToApi() {
		b.WriteString("&" + qemuApiKeyBoot + "=")
		if len(config) > 0 {
			b.WriteString("order" + equal)
			b.WriteString(v)
		}
	}
}

func (config BootOrder) Validate() error {
	var err error
	for i := range config {
		if err = config[i].Validate(); err != nil {
			return err
		}
		for ii := range len(config) - i - 1 {
			if config[ii+i+1] == config[i] {
				return errors.New(BootOrder_Error_Duplicate)
			}
		}
	}
	return nil
}

// All IDs of the following types are valid and can be converted:
//
//	type (
//		QemuIdeID
//		QemuNetworkInterfaceID
//		QemuPciID
//		QemuSataID
//		QemuScsiID
//		QemuUsbID
//		QemuVirtIoID
//	)
//	func <type>.BootDevice()
type BootDevice string

const (
	BootDevice_Error_Invalid string = "invalid boot device"
)

type bootDevice interface {
	BootDevice() BootDevice
}

var (
	_ bootDevice = QemuIdeID(0)              // 3,1 4
	_ bootDevice = QemuNetworkInterfaceID(0) // 3,2 4-5
	_ bootDevice = QemuPciID(0)              // 7,2 8-9
	_ bootDevice = QemuSataID(0)             // 4,1 5
	_ bootDevice = QemuScsiID(0)             // 4,2 5-6
	_ bootDevice = QemuUsbID(0)              // 3,1 4
	_ bootDevice = QemuVirtIoID(0)           // 6,2 7-8
)

func (boot BootDevice) Validate() error {
	r := errors.New(BootDevice_Error_Invalid)
	switch len(boot) {
	case 4: // ide, net, usb
		v, err := strconv.ParseUint(string(boot[3]), 10, 64)
		if err != nil {
			return r
		}
		switch boot[0:3] {
		case qemuPrefixApiKeyDiskIde:
			if v <= uint64(QemuIdeIDMaximum) {
				return nil
			}
		case qemuPrefixApiKeyNetwork:
			if v <= uint64(QemuNetworkInterfaceIDMaximum) {
				return nil
			}
		case qemuPrefixApiKeyUSB:
			if v <= uint64(QemuUsbIDMaximum) {
				return nil
			}
		}
	case 5: // net, sata, scsi
		if boot[0:3] == qemuPrefixApiKeyNetwork {
			v, err := strconv.ParseUint(string(boot[3:5]), 10, 64)
			if err != nil {
				return r
			}
			if err = QemuNetworkInterfaceID(v).Validate(); err != nil {
				return r
			}
			return nil
		}
		v, err := strconv.ParseUint(string(boot[4]), 10, 64)
		if err != nil {
			return r
		}
		switch boot[0:4] {
		case qemuPrefixApiKeyDiskSata:
			if v <= uint64(QemuSataIDMaximum) {
				return nil
			}
		case qemuPrefixApiKeyDiskSCSI:
			if v <= uint64(QemuScsiIDMaximum) {
				return nil
			}
		}
	case 6: // scsi
		if boot[0:4] == qemuPrefixApiKeyDiskSCSI {
			v, err := strconv.ParseUint(string(boot[4:6]), 10, 64)
			if err != nil {
				return r
			}
			if v <= uint64(QemuScsiIDMaximum) {
				return nil
			}
		}
	case 7: // virtio
		if boot[0:6] == qemuPrefixApiKeyDiskVirtIO {
			v, err := strconv.ParseUint(string(boot[6]), 10, 64)
			if err != nil {
				return r
			}
			if v <= uint64(QemuVirtIoIDMaximum) {
				return nil
			}
		}
	case 8: // virtio, hostpci,
		if boot[0:6] == qemuPrefixApiKeyDiskVirtIO {
			v, err := strconv.ParseUint(string(boot[6:8]), 10, 64)
			if err != nil {
				return r
			}
			if v <= uint64(QemuVirtIoIDMaximum) {
				return nil
			}
			return r
		}
		if boot[0:7] == qemuPrefixApiKeyPCI {
			v, err := strconv.ParseUint(string(boot[7]), 10, 64)
			if err != nil {
				return r
			}
			if v <= uint64(QemuPciIDMaximum) {
				return nil
			}
		}
	case 9: // hostpci
		if boot[0:7] == qemuPrefixApiKeyPCI {
			v, err := strconv.ParseUint(string(boot[7:9]), 10, 64)
			if err != nil {
				return r
			}
			if v <= uint64(QemuPciIDMaximum) {
				return nil
			}
		}
	}
	return r
}

func (boot BootDevice) String() string { return string(boot) } // String is for fmt.Stringer.

func (raw *rawConfigQemu) GetBoot() BootOrder {
	if v, isSet := raw.a[qemuApiKeyBoot]; isSet {
		vv := v.(string)
		if vv == " " {
			return nil
		}
		vv = strings.TrimPrefix(vv, "order=")
		options := strings.Split(vv, ";") // We have to preserve the order
		order := make([]BootDevice, len(options))
		for i := range options {
			order[i] = BootDevice(options[i])
		}
		return order
	}
	return nil
}
