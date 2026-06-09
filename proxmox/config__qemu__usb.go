package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/body"
)

type QemuUSBs map[QemuUsbID]QemuUSB

const QemuUSBsAmount = uint8(QemuUsbIDMaximum) + 1

func (raw *rawConfigQemu) GetUSBs() QemuUSBs {
	usbList := make(QemuUSBs)
	for i := range QemuUsbID(QemuUSBsAmount) {
		if v, isSet := raw.a[qemuPrefixApiKeyUSB+i.String()]; isSet {
			usbList[i] = QemuUSB{}.mapToSDK(v.(string))
		}
	}
	if len(usbList) > 0 {
		return usbList
	}
	return nil
}

func (config QemuUSBs) mapToApiCreate(b *strings.Builder) {
	for i, e := range config {
		if e.Delete {
			continue
		}
		b.WriteString("&" + qemuPrefixApiKeyUSB)
		b.WriteString(i.String())
		b.WriteRune('=')
		e.mapToApiCreate(b)
	}
}

func (config QemuUSBs) mapToApiUpdate(current QemuUSBs, b, delete *strings.Builder) {
	for i, e := range config {
		if v, ok := current[i]; ok { // update / delete
			if e.Delete {
				delete.WriteString("," + qemuPrefixApiKeyUSB)
				delete.WriteString(i.String())
			} else {
				e.mapToApiUpdate(v, i, b)
			}
		} else { // create
			if !e.Delete {
				b.WriteString("&" + qemuPrefixApiKeyUSB)
				b.WriteString(i.String())
				b.WriteRune('=')
				e.mapToApiCreate(b)
			}
		}
	}
}

func (config QemuUSBs) Validate(current QemuUSBs) (err error) {
	if len(current) > 0 {
		return config.validateUpdate(current)
	}
	return config.validateCreate()
}

func (config QemuUSBs) validateCreate() (err error) {
	for i, e := range config {
		if err = i.Validate(); err != nil {
			return
		}
		if e.Delete {
			continue
		}
		if err = e.validate(QemuUSB{}); err != nil {
			return
		}
	}
	return nil
}

func (config QemuUSBs) validateUpdate(current QemuUSBs) (err error) {
	for i, e := range config {
		if err = i.Validate(); err != nil {
			return
		}
		if e.Delete {
			continue
		}
		if current != nil {
			if v, isSet := (current)[i]; isSet {
				if err = e.validate(v); err != nil {
					return
				}
			} else if err = e.validate(QemuUSB{}); err != nil {
				return
			}
		}
	}
	return nil
}

// Enum
//
//	const (
//		QemuUsbID0
//		...
//		QemuUsbID4
//	)
type QemuUsbID uint8

const (
	QemuUsbID_Error_Invalid string = "usb id must be in the range 0-4"

	QemuUsbIDMaximum = QemuUsbID4

	QemuUsbID0 QemuUsbID = 0
	QemuUsbID1 QemuUsbID = 1
	QemuUsbID2 QemuUsbID = 2
	QemuUsbID3 QemuUsbID = 3
	QemuUsbID4 QemuUsbID = 4
)

func (id QemuUsbID) String() string { return strconv.Itoa(int(id)) } // String is for fmt.Stringer.

func (id QemuUsbID) Validate() error {
	if id > QemuUsbIDMaximum {
		return errors.New(QemuUsbID_Error_Invalid)
	}
	return nil
}

type QemuUSB struct {
	Device  *QemuUsbDevice  `json:"device,omitempty"`
	Mapping *QemuUsbMapping `json:"mapping,omitempty"`
	Port    *QemuUsbPort    `json:"port,omitempty"`
	Spice   *QemuUsbSpice   `json:"spice,omitempty"`
	Delete  bool            `json:"delete,omitempty"`
}

const (
	QemuUSB_Error_MutualExclusive string = "usb device, usb mapped, usb port, and usb spice are mutually exclusive"
	QemuUSB_Error_DeviceID        string = "usb device id is required during creation"
	QemuUSB_Error_MappingID       string = "usb mapping id is required during creation"
	QemuUSB_Error_PortID          string = "usb port id is required during creation"
)

func (config QemuUSB) mapToApiCreate(b *strings.Builder) {
	if config.Device != nil {
		b.WriteString("host" + equal)
		if config.Device.ID != nil {
			b.WriteString(body.Escape(config.Device.ID.String()))
		}
		if config.Device.USB3 != nil && *config.Device.USB3 {
			b.WriteString(comma + "usb3" + equal + "1")
		}
	} else if config.Mapping != nil {
		b.WriteString("mapping" + equal)
		if config.Mapping.ID != nil {
			b.WriteString(config.Mapping.ID.String())
		}
		if config.Mapping.USB3 != nil && *config.Mapping.USB3 {
			b.WriteString(comma + "usb3" + equal + "1")
		}
	} else if config.Port != nil {
		b.WriteString("host" + equal)
		if config.Port.ID != nil {
			b.WriteString(config.Port.ID.String())
		}
		if config.Port.USB3 != nil && *config.Port.USB3 {
			b.WriteString(comma + "usb3" + equal + "1")
		}
	} else if config.Spice != nil {
		b.WriteString("spice")
		if config.Spice.USB3 {
			b.WriteString(comma + "usb3" + equal + "1")
		}
	}
}

func (config QemuUSB) mapToApiUpdate(current QemuUSB, id QemuUsbID, builder *strings.Builder) {
	var b strings.Builder
	current.mapToApiCreate(&b)
	currentVal := b.String()
	b = strings.Builder{}
	if config.Device != nil {
		b.WriteString("host" + equal)
		if current.Device != nil {
			if config.Device.ID != nil {
				b.WriteString(body.Escape(config.Device.ID.String()))
			} else {
				b.WriteString(body.Escape(current.Device.ID.String()))
			}
			if config.Device.USB3 != nil {
				if *config.Device.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			} else {
				if *current.Device.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			}
		} else {
			if config.Device.ID != nil {
				b.WriteString(body.Escape(config.Device.ID.String()))
			}
			if config.Device.USB3 != nil && *config.Device.USB3 {
				b.WriteString(comma + "usb3" + equal + "1")
			}
		}
	} else if config.Mapping != nil {
		b.WriteString("mapping" + equal)
		if current.Mapping != nil {
			if config.Mapping.ID != nil {
				b.WriteString(config.Mapping.ID.String())
			} else {
				b.WriteString(current.Mapping.ID.String())
			}
			if config.Mapping.USB3 != nil {
				if *config.Mapping.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			} else {
				if *current.Mapping.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			}
		} else {
			if config.Mapping.ID != nil {
				b.WriteString(config.Mapping.ID.String())
			}
			if config.Mapping.USB3 != nil && *config.Mapping.USB3 {
				b.WriteString(comma + "usb3" + equal + "1")
			}
		}
	} else if config.Port != nil {
		b.WriteString("host" + equal)
		if current.Port != nil {
			if config.Port.ID != nil {
				b.WriteString(config.Port.ID.String())
			} else {
				b.WriteString(current.Port.ID.String())
			}
			if config.Port.USB3 != nil {
				if *config.Port.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			} else {
				if *current.Port.USB3 {
					b.WriteString(comma + "usb3" + equal + "1")
				}
			}
		} else {
			if config.Port.ID != nil {
				b.WriteString(config.Port.ID.String())
			}
			if config.Port.USB3 != nil && *config.Port.USB3 {
				b.WriteString(comma + "usb3" + equal + "1")
			}
		}
	} else if config.Spice != nil {
		b.WriteString("spice")
		if config.Spice.USB3 {
			b.WriteString(comma + "usb3" + equal + "1")
		}
	} else {
		return
	}
	newVal := b.String()
	if newVal != currentVal {
		builder.WriteString("&" + qemuPrefixApiKeyUSB)
		builder.WriteString(id.String())
		builder.WriteRune('=')
		builder.WriteString(newVal)
	}
}

func (QemuUSB) mapToSDK(rawUSB string) QemuUSB {
	var usb3 bool
	splitUSB := strings.Split(rawUSB, ",")
	if len(splitUSB) == 2 {
		usb3 = splitUSB[1] == "usb3=1"
	}
	usbType := strings.Split(splitUSB[0], "=")
	switch usbType[0] {
	case "host":
		if strings.Contains(usbType[1], ":") {
			return QemuUSB{Device: &QemuUsbDevice{ID: new(UsbDeviceID(usbType[1])), USB3: &usb3}}
		}
		return QemuUSB{Port: &QemuUsbPort{ID: new(UsbPortID(usbType[1])), USB3: &usb3}}
	case "mapping":
		return QemuUSB{Mapping: &QemuUsbMapping{ID: new(ResourceMappingUsbID(usbType[1])), USB3: &usb3}}
	case "spice":
		return QemuUSB{Spice: &QemuUsbSpice{USB3: usb3}}
	}
	return QemuUSB{}
}

func (config QemuUSB) Validate(current *QemuUSB) error {
	if current != nil {
		return config.validate(*current)
	}
	return config.validate(QemuUSB{})
}

func (config QemuUSB) validate(current QemuUSB) error {
	if config.Delete {
		return nil
	}
	var mutualExclusivity uint8
	if config.Device != nil {
		if config.Device.ID != nil {
			if err := config.Device.ID.Validate(); err != nil {
				return err
			}
		} else if current.Device == nil || current.Device.ID == nil {
			return errors.New(QemuUSB_Error_DeviceID)
		}
		mutualExclusivity++
	}
	if config.Mapping != nil {
		if config.Mapping.ID != nil {
			if err := config.Mapping.ID.Validate(); err != nil {
				return err
			}
		} else if current.Mapping == nil || current.Mapping.ID == nil {
			return errors.New(QemuUSB_Error_MappingID)
		}
		mutualExclusivity++
	}
	if config.Port != nil {
		if config.Port.ID != nil {
			if err := config.Port.ID.Validate(); err != nil {
				return err
			}
		} else if current.Port == nil || current.Port.ID == nil {
			return errors.New(QemuUSB_Error_PortID)
		}
		mutualExclusivity++
	}
	if config.Spice != nil {
		mutualExclusivity++
	}
	if mutualExclusivity > 1 {
		return errors.New(QemuUSB_Error_MutualExclusive)
	}
	return nil
}

type QemuUsbDevice struct {
	ID   *UsbDeviceID `json:"id,omitempty"`   // never nil when returned
	USB3 *bool        `json:"usb3,omitempty"` // never nil when returned
}

func (config QemuUsbDevice) Validate() error {
	if config.ID == nil {
		return nil
	}
	return config.ID.Validate()
}

type QemuUsbMapping struct {
	ID   *ResourceMappingUsbID `json:"id,omitempty"`   // never nil when returned
	USB3 *bool                 `json:"usb3,omitempty"` // never nil when returned
}

func (config QemuUsbMapping) Validate() error {
	if config.ID == nil {
		return nil
	}
	return config.ID.Validate()
}

type QemuUsbPort struct {
	ID   *UsbPortID `json:"id,omitempty"`   // never nil when returned
	USB3 *bool      `json:"usb3,omitempty"` // never nil when returned
}

func (config QemuUsbPort) Validate() error {
	if config.ID == nil {
		return nil
	}
	return config.ID.Validate()
}

type QemuUsbSpice struct {
	USB3 bool `json:"usb3"`
}

type UsbDeviceID string

const (
	UsbDeviceID_Error_Invalid   string = "invalid usb device-id"
	UsbDeviceID_Error_VendorID  string = "usb vendor-id isn't valid hexadecimal"
	UsbDeviceID_Error_ProductID string = "usb product-id isn't valid hexadecimal"
)

func (id UsbDeviceID) Validate() error {
	rawID := strings.Split(string(id), ":")
	if len(rawID) != 2 {
		return errors.New(UsbDeviceID_Error_Invalid)
	}
	if _, err := strconv.ParseUint(rawID[0], 16, 16); err != nil {
		return errors.New(UsbDeviceID_Error_VendorID)
	}
	if _, err := strconv.ParseUint(rawID[1], 16, 16); err != nil {
		return errors.New(UsbDeviceID_Error_ProductID)
	}
	return nil
}

func (id UsbDeviceID) String() string { return string(id) } // String is for fmt.Stringer.

type UsbPortID string // regex: \d+-\d+

const UsbPortID_Error_Invalid string = "invalid usb port id. Expected expression of the form '<bus>-<port>(.<port>)*' where bus and port are integers"

func (id UsbPortID) String() string { return string(id) } // String is for fmt.Stringer.

func (id UsbPortID) Validate() error {
	idArray := strings.Split(string(id), "-")
	if len(idArray) != 2 {
		return errors.New(UsbPortID_Error_Invalid)
	}
	if _, err := strconv.Atoi(idArray[0]); err != nil {
		return errors.New(UsbPortID_Error_Invalid)
	}
	parts := strings.Split(idArray[1], ".")
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return errors.New(UsbPortID_Error_Invalid)
		}
	}
	return nil
}
