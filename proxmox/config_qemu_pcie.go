package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type QemuPciDevices map[QemuPciID]QemuPci

const QemuPciDevicesAmount = uint8(QemuPciIDMaximum) + 1

func (config QemuPciDevices) mapToAPI(current QemuPciDevices, params map[string]interface{}) string {
	var builder strings.Builder
	for i, e := range config {
		if v, isSet := current[i]; isSet {
			if e.Delete {
				builder.WriteString(",hostpci" + i.String())
				continue
			}
			params["hostpci"+i.String()] = e.mapToAPI(&v)
		} else {
			if e.Delete {
				continue
			}
			params["hostpci"+i.String()] = e.mapToAPI(nil)
		}
	}
	return builder.String()
}

func (QemuPciDevices) mapToSDK(params map[string]interface{}) QemuPciDevices {
	pciList := make(QemuPciDevices)
	for i := QemuPciID(0); i < QemuPciID(QemuPciDevicesAmount); i++ {
		if v, isSet := params["hostpci"+i.String()]; isSet {
			pciList[i] = QemuPci{}.mapToSDK(v.(string))
		}
	}
	if len(pciList) > 0 {
		return pciList
	}
	return nil
}

func (config QemuPciDevices) Validate(current QemuPciDevices) (err error) {
	for i, e := range config {
		if err = i.Validate(); err != nil {
			return
		}
		if e.Delete {
			continue
		}
		if current != nil {
			if v, isSet := (current)[i]; isSet {
				if err = e.Validate(v); err != nil {
					return
				}
			}
		} else {
			if err = e.Validate(QemuPci{}); err != nil {
				return
			}
		}
	}
	return nil
}

type QemuPciID uint8

const (
	QemuPciID_Error_Invalid string = "pci id must be in the range 0-15"

	QemuPciIDMaximum = QemuPciID15

	QemuPciID0  QemuPciID = 0
	QemuPciID1  QemuPciID = 1
	QemuPciID2  QemuPciID = 2
	QemuPciID3  QemuPciID = 3
	QemuPciID4  QemuPciID = 4
	QemuPciID5  QemuPciID = 5
	QemuPciID6  QemuPciID = 6
	QemuPciID7  QemuPciID = 7
	QemuPciID8  QemuPciID = 8
	QemuPciID9  QemuPciID = 9
	QemuPciID10 QemuPciID = 10
	QemuPciID11 QemuPciID = 11
	QemuPciID12 QemuPciID = 12
	QemuPciID13 QemuPciID = 13
	QemuPciID14 QemuPciID = 14
	QemuPciID15 QemuPciID = 15
)

func (id QemuPciID) String() string {
	return strconv.Itoa(int(id))
}

func (id QemuPciID) Validate() error {
	if id > QemuPciIDMaximum {
		return errors.New(QemuPciID_Error_Invalid)
	}
	return nil
}

type QemuPci struct {
	Delete  bool            `json:"delete,omitempty"`
	Mapping *QemuPciMapping `json:"mapping,omitempty"`
	Raw     *QemuPciRaw     `json:"raw,omitempty"`
}

const (
	QemuPci_Error_MutualExclusive string = "mapping and raw are mutually exclusive"
	QemuPci_Error_MappedID        string = "mapped id is required during creation"
	QemuPci_Error_RawID           string = "raw id is required during creation"
)

func (config QemuPci) mapToAPI(current *QemuPci) string {
	var usedConfig qemuPci
	if current != nil {
		usedConfig = current.mapToApiIntermediary(qemuPci{})
	}
	return config.mapToApiIntermediary(usedConfig).String()
}

func (QemuPci) mapToSDK(raw string) QemuPci {
	rawSettings := strings.SplitN(raw, ",", 2)
	var mappingID *ResourceMappingPciID
	var settings map[string]string
	if strings.IndexByte(rawSettings[0], '=') == -1 {
		if len(rawSettings) > 1 {
			settings = splitStringOfSettings(rawSettings[1])
		}
	} else {
		settings = splitStringOfSettings(raw)
		if v, isSet := settings["mapping"]; isSet {
			mappingID = util.Pointer(ResourceMappingPciID(v))
		}
	}
	var pcie, primaryGPU, romBar = false, false, true
	var vendorID *PciVendorID
	var deviceID *PciDeviceID
	var subVendorID *PciSubVendorID
	var subDeviceID *PciSubDeviceID
	if v, isSet := settings["pcie"]; isSet {
		pcie = v == "1"
	}
	if v, isSet := settings["x-vga"]; isSet {
		primaryGPU = v == "1"
	}
	if v, isSet := settings["rombar"]; isSet {
		romBar = v == "1"
	}
	if v, isSet := settings["vendor-id"]; isSet {
		vendorID = util.Pointer(PciVendorID(v))
	}
	if v, isSet := settings["device-id"]; isSet {
		deviceID = util.Pointer(PciDeviceID(v))
	}
	if v, isSet := settings["sub-vendor-id"]; isSet {
		subVendorID = util.Pointer(PciSubVendorID(v))
	}
	if v, isSet := settings["sub-device-id"]; isSet {
		subDeviceID = util.Pointer(PciSubDeviceID(v))
	}
	if mappingID != nil {
		return QemuPci{Mapping: &QemuPciMapping{
			ID:          mappingID,
			PCIe:        &pcie,
			PrimaryGPU:  &primaryGPU,
			ROMbar:      &romBar,
			VendorID:    vendorID,
			DeviceID:    deviceID,
			SubVendorID: subVendorID,
			SubDeviceID: subDeviceID,
		}}
	}
	return QemuPci{Raw: &QemuPciRaw{
		ID:          util.Pointer(PciID(rawSettings[0])),
		PCIe:        &pcie,
		PrimaryGPU:  &primaryGPU,
		ROMbar:      &romBar,
		VendorID:    vendorID,
		DeviceID:    deviceID,
		SubVendorID: subVendorID,
		SubDeviceID: subDeviceID,
	}}
}

func (config QemuPci) mapToApiIntermediary(usedConfig qemuPci) qemuPci {
	if config.Mapping != nil {
		usedConfig.enum = qemuPCciEnumMapping
		if config.Mapping.ID != nil {
			usedConfig.mappingID = *config.Mapping.ID
		}
		if config.Mapping.PCIe != nil {
			usedConfig.pCIe = *config.Mapping.PCIe
		}
		if config.Mapping.PrimaryGPU != nil {
			usedConfig.primaryGPU = *config.Mapping.PrimaryGPU
		}
		if config.Mapping.ROMbar != nil {
			usedConfig.romBar = *config.Mapping.ROMbar
		}
		if config.Mapping.VendorID != nil {
			usedConfig.vendorID = *config.Mapping.VendorID
		}
		if config.Mapping.DeviceID != nil {
			usedConfig.deviceID = *config.Mapping.DeviceID
		}
		if config.Mapping.SubVendorID != nil {
			usedConfig.subVendorID = *config.Mapping.SubVendorID
		}
		if config.Mapping.SubDeviceID != nil {
			usedConfig.subDeviceID = *config.Mapping.SubDeviceID
		}
		return usedConfig
	}
	if config.Raw != nil {
		usedConfig.enum = qemuPciEnumRaw
		if config.Raw.ID != nil {
			usedConfig.rawID = *config.Raw.ID
		}
		if config.Raw.PCIe != nil {
			usedConfig.pCIe = *config.Raw.PCIe
		}
		if config.Raw.PrimaryGPU != nil {
			usedConfig.primaryGPU = *config.Raw.PrimaryGPU
		}
		if config.Raw.ROMbar != nil {
			usedConfig.romBar = *config.Raw.ROMbar
		}
		if config.Raw.VendorID != nil {
			usedConfig.vendorID = *config.Raw.VendorID
		}
		if config.Raw.DeviceID != nil {
			usedConfig.deviceID = *config.Raw.DeviceID
		}
		if config.Raw.SubVendorID != nil {
			usedConfig.subVendorID = *config.Raw.SubVendorID
		}
		if config.Raw.SubDeviceID != nil {
			usedConfig.subDeviceID = *config.Raw.SubDeviceID
		}
	}
	return usedConfig
}

func (config QemuPci) Validate(current QemuPci) error {
	if config.Delete {
		return nil
	}
	var mutualExclusivity uint8
	if config.Mapping != nil {
		if config.Mapping.ID != nil {
			if err := config.Mapping.ID.Validate(); err != nil {
				return err
			}
		} else if current.Mapping == nil || current.Mapping.ID == nil {
			return errors.New(QemuPci_Error_MappedID)
		}
		if config.Mapping.DeviceID != nil {
			if err := config.Mapping.DeviceID.Validate(); err != nil {
				return err
			}
		}
		if config.Mapping.SubDeviceID != nil {
			if err := config.Mapping.SubDeviceID.Validate(); err != nil {
				return err
			}
		}
		if config.Mapping.SubVendorID != nil {
			if err := config.Mapping.SubVendorID.Validate(); err != nil {
				return err
			}
		}
		if config.Mapping.VendorID != nil {
			if err := config.Mapping.VendorID.Validate(); err != nil {
				return err
			}
		}
		mutualExclusivity++
	}
	if config.Raw != nil {
		if config.Raw.ID != nil {
			if err := config.Raw.ID.Validate(); err != nil {
				return err
			}
		} else if current.Raw == nil || current.Raw.ID == nil {
			return errors.New(QemuPci_Error_RawID)
		}
		if config.Raw.DeviceID != nil {
			if err := config.Raw.DeviceID.Validate(); err != nil {
				return err
			}
		}
		if config.Raw.SubDeviceID != nil {
			if err := config.Raw.SubDeviceID.Validate(); err != nil {
				return err
			}
		}
		if config.Raw.SubVendorID != nil {
			if err := config.Raw.SubVendorID.Validate(); err != nil {
				return err
			}
		}
		if config.Raw.VendorID != nil {
			if err := config.Raw.VendorID.Validate(); err != nil {
				return err
			}
		}
		mutualExclusivity++
	}
	if mutualExclusivity > 1 {
		return errors.New(QemuPci_Error_MutualExclusive)
	}
	return nil
}

// TODO add [,legacy-igd=<1|0>]
// TODO add [,mdev=<string>]
// TODO add [,romfile=<string>]
type qemuPci struct {
	enum        qemuPciEnum
	mappingID   ResourceMappingPciID // [,mapping=<mapping-id>]
	rawID       PciID                // [[host=]<HOSTPCIID[;HOSTPCIID2...]>]
	pCIe        bool                 // [,pcie=<1|0>] // only in key when true
	primaryGPU  bool                 // [,x-vga=<1|0>] // only in key when true
	romBar      bool                 // [,rombar=<1|0>] // only in key when false
	vendorID    PciVendorID          // [,vendor-id=<hex id>]
	deviceID    PciDeviceID          // [,device-id=<hex id>]
	subVendorID PciSubVendorID       // [,sub-vendor-id=<hex id>]
	subDeviceID PciSubDeviceID       // [,sub-device-id=<hex id>]
}

const (
	qemuPCciEnumMapping qemuPciEnum = true
	qemuPciEnumRaw      qemuPciEnum = false

	qemuPci_Error_Number  string = " must be a hexadecimal number"
	qemuPci_Error_Maximum string = " must be in the range 0x0000-0xFFFF"
)

func (config qemuPci) String() string {
	var builder strings.Builder
	if config.pCIe {
		builder.WriteString(",pcie=1")
	}
	if config.primaryGPU {
		builder.WriteString(",x-vga=1")
	}
	if !config.romBar {
		builder.WriteString(",rombar=0")
	}
	if config.vendorID != "" {
		builder.WriteString(",vendor-id=" + config.vendorID.String())
	}
	if config.deviceID != "" {
		builder.WriteString(",device-id=" + config.deviceID.String())
	}
	if config.subVendorID != "" {
		builder.WriteString(",sub-vendor-id=" + config.subVendorID.String())
	}
	if config.subDeviceID != "" {
		builder.WriteString(",sub-device-id=" + config.subDeviceID.String())
	}
	var settings string
	switch config.enum {
	case qemuPCciEnumMapping:
		settings = "mapping=" + string(config.mappingID)
	case qemuPciEnumRaw:
		settings = config.rawID.String()
	}
	return settings + builder.String()
}

type qemuPciEnum bool

type QemuPciMapping struct {
	DeviceID    *PciDeviceID          `json:"device_id,omitempty"`
	ID          *ResourceMappingPciID `json:"id,omitempty"`
	PCIe        *bool                 `json:"pcie,omitempty"`
	PrimaryGPU  *bool                 `json:"gpu,omitempty"`
	ROMbar      *bool                 `json:"rombar,omitempty"`
	SubDeviceID *PciSubDeviceID       `json:"sub_device_id,omitempty"`
	SubVendorID *PciSubVendorID       `json:"sub_vendor_id,omitempty"`
	VendorID    *PciVendorID          `json:"vendor_id,omitempty"`
}

type QemuPciRaw struct {
	DeviceID    *PciDeviceID    `json:"device_id,omitempty"`
	ID          *PciID          `json:"id,omitempty"`
	PCIe        *bool           `json:"pcie,omitempty"`
	PrimaryGPU  *bool           `json:"gpu,omitempty"`
	ROMbar      *bool           `json:"rombar,omitempty"`
	SubDeviceID *PciSubDeviceID `json:"sub_device_id,omitempty"`
	SubVendorID *PciSubVendorID `json:"sub_vendor_id,omitempty"`
	VendorID    *PciVendorID    `json:"vendor_id,omitempty"`
}

// Hexadecimal, range 0x0000-0xFFFF, prefixed is optional
// Set to empty string to remove the device id
type PciDeviceID string

const PciDeviceID_Error_Invalid string = "device id" + qemuPci_Error_Maximum

func (id PciDeviceID) String() string {
	if id == "" {
		return ""
	}
	return ensurePrefix(hexPrefix, strings.ToLower(string(id)))
}

func (id PciDeviceID) Validate() error {
	if id == "" {
		return nil
	}
	if _, err := strconv.ParseUint(strings.TrimPrefix(string(id), "0x"), 16, 16); err != nil {
		return errors.New(PciDeviceID_Error_Invalid)
	}
	return nil
}

// Hexadecimal, range 0x0000-0xFFFF, prefixed is optional
// Set to empty string to remove the sub device id
type PciSubDeviceID string

const PciSubDeviceID_Error_Invalid string = "sub device id" + qemuPci_Error_Maximum

func (id PciSubDeviceID) String() string {
	if id == "" {
		return ""
	}
	return ensurePrefix(hexPrefix, strings.ToLower(string(id)))
}

func (id PciSubDeviceID) Validate() error {
	if id == "" {
		return nil
	}
	if _, err := strconv.ParseUint(strings.TrimPrefix(string(id), "0x"), 16, 16); err != nil {
		return errors.New(PciSubDeviceID_Error_Invalid)
	}
	return nil
}

// Hexadecimal, range 0x0000-0xFFFF, prefixed is optional
// Set to empty string to remove the sub vendor id
type PciSubVendorID string

const PciSubVendorID_Error_Invalid string = "sub vendor id" + qemuPci_Error_Maximum

func (id PciSubVendorID) String() string {
	if id == "" {
		return ""
	}
	return ensurePrefix(hexPrefix, strings.ToLower(string(id)))
}

func (id PciSubVendorID) Validate() error {
	if id == "" {
		return nil
	}
	if _, err := strconv.ParseUint(strings.TrimPrefix(string(id), "0x"), 16, 16); err != nil {
		return errors.New(PciSubVendorID_Error_Invalid)
	}
	return nil
}

// Hexadecimal, range 0x0000-0xFFFF, prefixed is optional
// Set to empty string to remove the vendor id
type PciVendorID string

const PciVendorID_Error_Invalid string = "vendor id" + qemuPci_Error_Maximum

func (id PciVendorID) String() string {
	if id == "" {
		return ""
	}
	return ensurePrefix(hexPrefix, strings.ToLower(string(id)))
}

func (id PciVendorID) Validate() error {
	if id == "" {
		return nil
	}
	if _, err := strconv.ParseUint(strings.TrimPrefix(string(id), "0x"), 16, 16); err != nil {
		return errors.New(PciVendorID_Error_Invalid)
	}
	return nil
}

// 0000:00:00.1
type PciID string

const (
	PciID_Error_InvalidBus      string = "pci id invalid bus identifier"
	PciID_Error_InvalidDomain   string = "pci id invalid domain identifier"
	PciID_Error_LengthBus       string = "pci id bus identifier should be 2 characters long"
	PciID_Error_LengthDomain    string = "pci id domain identifier should be 4 characters long"
	PciID_Error_MissingBus      string = "pci id missing bus identifier"
	PciID_Error_MissingDevice   string = "pci id missing device identifier"
	PciID_Error_LengthDevice    string = "pci id device identifier should be 2 characters long"
	PciID_Error_InvalidDevice   string = "pci id invalid device identifier"
	PciID_Error_InvalidFunction string = "pci id invalid function identifier"
	PciID_Error_MaximumFunction string = "pci id function identifier should be in the range 0-7"
)

func (id PciID) String() string {
	return string(id)
}

func (id PciID) Validate() error {
	idParts := strings.Split(string(id), ":")
	if len(idParts) < 3 {
		if len(idParts) < 2 {
			return errors.New(PciID_Error_MissingBus)
		}
		return errors.New(PciID_Error_MissingDevice)
	}
	if len(idParts[0]) != 4 {
		return errors.New(PciID_Error_LengthDomain)
	}
	if _, err := strconv.ParseUint(idParts[0], 16, 16); err != nil {
		return errors.New(PciID_Error_InvalidDomain)
	}
	if len(idParts[1]) != 2 {
		return errors.New(PciID_Error_LengthBus)
	}
	if _, err := strconv.ParseUint(idParts[1], 16, 16); err != nil {
		return errors.New(PciID_Error_InvalidBus)
	}
	deviceAndFunction := strings.Split(idParts[2], ".")
	if len(deviceAndFunction[0]) != 2 {
		return errors.New(PciID_Error_LengthDevice)
	}
	if _, err := strconv.ParseUint(deviceAndFunction[0], 16, 16); err != nil {
		return errors.New(PciID_Error_InvalidDevice)
	}
	if len(deviceAndFunction) == 2 {
		convertedID, err := strconv.ParseUint(deviceAndFunction[1], 10, 0)
		if err != nil {
			return errors.New(PciID_Error_InvalidFunction)
		}
		if convertedID > 7 {
			return errors.New(PciID_Error_MaximumFunction)
		}
	}
	return nil
}
