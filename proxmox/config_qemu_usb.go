package proxmox

import (
	"strconv"
	"strings"
)

type qemuUSB struct {
	Type    qemuUsbType
	Host    string
	Usb3    bool
	Mapping ResourceMappingUsbID
}

func (usb qemuUSB) String() (param string) {
	switch usb.Type {
	case qemuUsbTypeSpice:
		param = "spice"
	case qemuUsbTypeMapping:
		param = "mapping=" + usb.Mapping.String()
	case qemuUsbTypeDevice:
		param = "host=" + usb.Host
	case qemuUsbTypePort:
		param = "host=" + usb.Host
	}
	if usb.Usb3 {
		param += ",usb3=1"
	}
	return
}

type qemuUsbType uint8

const (
	qemuUsbTypeSpice   qemuUsbType = 0
	qemuUsbTypeMapping qemuUsbType = 1
	qemuUsbTypeDevice  qemuUsbType = 2
	qemuUsbTypePort    qemuUsbType = 3
)

type QemuUSBs map[QemuUsbID]QemuUSB

func (config QemuUSBs) mapToAPI(current QemuUSBs, params map[string]interface{}) string {
	var builder strings.Builder
	for i, e := range config {
		if v, isSet := current[i]; isSet {
			if e.Delete {
				builder.WriteString(",usb" + strconv.Itoa(int(i)))
				continue
			}
			params["usb"+strconv.Itoa(int(i))] = e.mapToAPI(&v)
		} else {
			if e.Delete {
				continue
			}
			params["usb"+strconv.Itoa(int(i))] = e.mapToAPI(nil)
		}
	}
	return builder.String()
}

type QemuUsbID uint8

const (
	QemuUsbID0 QemuUsbID = 0
	QemuUsbID1 QemuUsbID = 1
	QemuUsbID2 QemuUsbID = 2
	QemuUsbID3 QemuUsbID = 3
	QemuUsbID4 QemuUsbID = 4
)

type QemuUSB struct {
	Delete  bool            `json:"delete,omitempty"`
	Device  *QemuUsbDevice  `json:"device,omitempty"`
	Mapping *QemuUsbMapping `json:"mapping,omitempty"`
	Port    *QemuUsbPort    `json:"port,omitempty"`
	Spice   *QemuUsbSpice   `json:"spice,omitempty"`
}

func (config QemuUSB) mapToAPI(current *QemuUSB) string {
	var usb qemuUSB
	if current != nil {
		if current.Device != nil {
			if current.Device.ID != nil {
				usb.Host = (*current.Device.ID).String()
			}
			if current.Device.USB3 != nil {
				usb.Usb3 = *current.Device.USB3
			}
		} else if current.Mapping != nil {
			if current.Mapping.ID != nil {
				usb.Mapping = *current.Mapping.ID
			}
			if current.Mapping.USB3 != nil {
				usb.Usb3 = *current.Mapping.USB3
			}
		} else if current.Port != nil {
			if current.Port.ID != nil {
				usb.Host = string(*current.Port.ID)
			}
			if current.Port.USB3 != nil {
				usb.Usb3 = *current.Port.USB3
			}
		} else if current.Spice != nil {
			usb.Usb3 = current.Spice.USB3
		}
	}
	if config.Device != nil {
		usb.Type = qemuUsbTypeDevice
		if config.Device.USB3 != nil {
			usb.Usb3 = *config.Device.USB3
		}
		if config.Device.ID != nil {
			usb.Host = (*config.Device.ID).String()
		}
	} else if config.Mapping != nil {
		usb.Type = qemuUsbTypeMapping
		if config.Mapping.USB3 != nil {
			usb.Usb3 = *config.Mapping.USB3
		}
		if config.Mapping.ID != nil {
			usb.Mapping = *config.Mapping.ID
		}
	} else if config.Port != nil {
		usb.Type = qemuUsbTypePort
		if config.Port.USB3 != nil {
			usb.Usb3 = *config.Port.USB3
		}
		if config.Port.ID != nil {
			usb.Host = (*config.Port.ID).String()
		}
	} else if config.Spice != nil {
		usb.Type = qemuUsbTypeSpice
		if config.Spice.USB3 {
			usb.Usb3 = config.Spice.USB3
		}
	}
	return usb.String()
}

type QemuUsbDevice struct {
	ID   *UsbDeviceID `json:"id,omitempty"`
	USB3 *bool        `json:"usb3,omitempty"`
}

type QemuUsbMapping struct {
	ID   *ResourceMappingUsbID `json:"id,omitempty"`
	USB3 *bool                 `json:"usb3,omitempty"`
}

type QemuUsbPort struct {
	ID   *UsbPortID `json:"id,omitempty"`
	USB3 *bool      `json:"usb3,omitempty"`
}

type QemuUsbSpice struct {
	USB3 bool `json:"usb3"`
}

type UsbDeviceID string

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
