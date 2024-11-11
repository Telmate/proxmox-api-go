package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type qemuUSB struct {
	enum    qemuUsbEnum
	Host    string
	Usb3    bool
	Mapping ResourceMappingUsbID
}

func (usb qemuUSB) String() (param string) {
	switch usb.enum {
	case qemuUsbEnumSpice:
		param = "spice"
	case qemuUsbEnumMapping:
		param = "mapping=" + usb.Mapping.String()
	case qemuUsbEnumDevice:
		param = "host=" + usb.Host
	case qemuUsbEnumPort:
		param = "host=" + usb.Host
	}
	if usb.Usb3 {
		param += ",usb3=1"
	}
	return
}

type qemuUsbEnum uint8

const (
	qemuUsbEnumSpice   qemuUsbEnum = 0
	qemuUsbEnumMapping qemuUsbEnum = 1
	qemuUsbEnumDevice  qemuUsbEnum = 2
	qemuUsbEnumPort    qemuUsbEnum = 3
)

type QemuUSBs map[QemuUsbID]QemuUSB

const QemuUSBsAmount = uint8(QemuUsbIDMaximum) + 1

func (QemuUSBs) mapToSDK(params map[string]interface{}) QemuUSBs {
	usbList := make(QemuUSBs)
	for i := QemuUsbID(0); i < QemuUsbID(QemuUSBsAmount); i++ {
		if v, isSet := params["usb"+i.String()]; isSet {
			usbList[i] = QemuUSB{}.mapToSDK(v.(string))
		}
	}
	if len(usbList) > 0 {
		return usbList
	}
	return nil
}

func (config QemuUSBs) mapToAPI(current QemuUSBs, params map[string]interface{}) string {
	var builder strings.Builder
	for i, e := range config {
		if v, isSet := current[i]; isSet {
			if e.Delete {
				builder.WriteString(",usb" + i.String())
				continue
			}
			params["usb"+i.String()] = e.mapToAPI(&v)
		} else {
			if e.Delete {
				continue
			}
			params["usb"+i.String()] = e.mapToAPI(nil)
		}
	}
	return builder.String()
}

func (config QemuUSBs) Validate(current QemuUSBs) (err error) {
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
			}
		} else {
			if err = e.validate(QemuUSB{}); err != nil {
				return
			}
		}
	}
	return nil
}

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

func (id QemuUsbID) String() string {
	return strconv.Itoa(int(id))
}

func (id QemuUsbID) Validate() error {
	if id > QemuUsbIDMaximum {
		return errors.New(QemuUsbID_Error_Invalid)
	}
	return nil
}

type QemuUSB struct {
	Delete  bool            `json:"delete,omitempty"`
	Device  *QemuUsbDevice  `json:"device,omitempty"`
	Mapping *QemuUsbMapping `json:"mapping,omitempty"`
	Port    *QemuUsbPort    `json:"port,omitempty"`
	Spice   *QemuUsbSpice   `json:"spice,omitempty"`
}

const (
	QemuUSB_Error_MutualExclusive string = "usb device, usb mapped, usb port, and usb spice are mutually exclusive"
	QemuUSB_Error_DeviceID        string = "usb device id is required during creation"
	QemuUSB_Error_MappingID       string = "usb mapping id is required during creation"
	QemuUSB_Error_PortID          string = "usb port id is required during creation"
)

func (config QemuUSB) mapToAPI(current *QemuUSB) string {
	var usedConfig qemuUSB
	if current != nil {
		usedConfig = current.mapToApiIntermediary(usedConfig)
	}
	return config.mapToApiIntermediary(usedConfig).String()
}

func (config QemuUSB) mapToApiIntermediary(usedConfig qemuUSB) qemuUSB {
	if config.Device != nil {
		usedConfig.enum = qemuUsbEnumDevice
		if config.Device.USB3 != nil {
			usedConfig.Usb3 = *config.Device.USB3
		}
		if config.Device.ID != nil {
			usedConfig.Host = (*config.Device.ID).String()
		}
	} else if config.Mapping != nil {
		usedConfig.enum = qemuUsbEnumMapping
		if config.Mapping.USB3 != nil {
			usedConfig.Usb3 = *config.Mapping.USB3
		}
		if config.Mapping.ID != nil {
			usedConfig.Mapping = *config.Mapping.ID
		}
	} else if config.Port != nil {
		usedConfig.enum = qemuUsbEnumPort
		if config.Port.USB3 != nil {
			usedConfig.Usb3 = *config.Port.USB3
		}
		if config.Port.ID != nil {
			usedConfig.Host = (*config.Port.ID).String()
		}
	} else if config.Spice != nil {
		usedConfig.enum = qemuUsbEnumSpice
		if config.Spice.USB3 {
			usedConfig.Usb3 = config.Spice.USB3
		}
	}
	return usedConfig
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
			return QemuUSB{Device: &QemuUsbDevice{ID: util.Pointer(UsbDeviceID(usbType[1])), USB3: &usb3}}
		}
		return QemuUSB{Port: &QemuUsbPort{ID: util.Pointer(UsbPortID(usbType[1])), USB3: &usb3}}
	case "mapping":
		return QemuUSB{Mapping: &QemuUsbMapping{ID: util.Pointer(ResourceMappingUsbID(usbType[1])), USB3: &usb3}}
	case "spice":
		return QemuUSB{Spice: &QemuUsbSpice{USB3: usb3}}
	}
	return QemuUSB{}
}

func (config QemuUSB) Validate(current *QemuUSB) (err error) {
	if current != nil {
		err = config.validate(*current)
	} else {
		err = config.validate(QemuUSB{})
	}
	return err
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
	ID   *UsbDeviceID `json:"id,omitempty"`
	USB3 *bool        `json:"usb3,omitempty"`
}

func (config QemuUsbDevice) Validate() error {
	if config.ID == nil {
		return nil
	}
	return config.ID.Validate()
}

type QemuUsbMapping struct {
	ID   *ResourceMappingUsbID `json:"id,omitempty"`
	USB3 *bool                 `json:"usb3,omitempty"`
}

func (config QemuUsbMapping) Validate() error {
	if config.ID == nil {
		return nil
	}
	return config.ID.Validate()
}

type QemuUsbPort struct {
	ID   *UsbPortID `json:"id,omitempty"`
	USB3 *bool      `json:"usb3,omitempty"`
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

func (id UsbDeviceID) String() string {
	return string(id)
}

type UsbPortID string // regex: \d+-\d+

const (
	UsbPortID_Error_Invalid string = "invalid usb port id"
)

func (id UsbPortID) String() string {
	return string(id)
}

func (id UsbPortID) Validate() error {
	idArray := strings.Split(string(id), "-")
	if len(idArray) != 2 {
		return errors.New(UsbPortID_Error_Invalid)
	}
	if _, err := strconv.Atoi(idArray[0]); err != nil {
		return errors.New(UsbPortID_Error_Invalid)
	}
	if _, err := strconv.Atoi(idArray[1]); err != nil {
		return errors.New(UsbPortID_Error_Invalid)
	}
	return nil
}
