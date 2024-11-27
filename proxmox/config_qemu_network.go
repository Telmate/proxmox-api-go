package proxmox

import (
	"errors"
	"net"
	"slices"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type QemuMTU struct {
	Inherit bool `json:"inherit,omitempty"`
	Value   MTU  `json:"value,omitempty"`
}

const QemuMTU_Error_Invalid string = "inherit and value are mutually exclusive"

// unsafe requires caller to check for nil
func (config *QemuMTU) mapToApiUnsafe(builder *strings.Builder) {
	if config.Inherit {
		builder.WriteString(",mtu=1")
		return
	}
	if config.Value != 0 {
		builder.WriteString(",mtu=" + strconv.Itoa(int(config.Value)))
	}
}

func (config QemuMTU) Validate() error {
	if config.Inherit {
		if config.Value != 0 {
			return errors.New(QemuMTU_Error_Invalid)
		}
		return nil
	}
	return config.Value.Validate()
}

const (
	QemuNetworkInterface_Error_BridgeRequired string = "bridge is required during creation"
	QemuNetworkInterface_Error_ModelRequired  string = "model is required during creation"
	QemuNetworkInterface_Error_MtuNoEffect    string = "mtu has no effect when model is not virtio"
)

// if we get more edge cases, we should give every model its own struct
type QemuNetworkInterface struct {
	Bridge        *string           `json:"bridge,omitempty"` // Required for creation
	Delete        bool              `json:"delete,omitempty"`
	Connected     *bool             `json:"connected,omitempty"`
	Firewall      *bool             `json:"firewall,omitempty"`
	MAC           *net.HardwareAddr `json:"mac,omitempty"`
	MTU           *QemuMTU          `json:"mtu,omitempty"`   // only when `Model == QemuNetworkModelVirtIO`
	Model         *QemuNetworkModel `json:"model,omitempty"` // Required for creation
	MultiQueue    *QemuNetworkQueue `json:"queue,omitempty"`
	RateLimitKBps *QemuNetworkRate  `json:"rate,omitempty"`
	NativeVlan    *Vlan             `json:"native_vlan,omitempty"`
	TaggedVlans   *Vlans            `json:"tagged_vlans,omitempty"`
	mac           string
}

func (config QemuNetworkInterface) mapToApi(current *QemuNetworkInterface) (settings string) {
	builder := strings.Builder{}
	var mac, model string
	if current != nil { // Update
		if config.Model != nil {
			model = config.Model.String()
		} else if current.Model != nil {
			model = current.Model.String()
		}
		builder.WriteString(model)
		if config.MAC != nil {
			mac = config.MAC.String() // Returns a lowercase MAC address
			if mac == strings.ToLower(current.mac) {
				mac = current.mac
			} else {
				mac = strings.ToUpper(mac)
			}
			builder.WriteString("=" + mac)
		} else if current.MAC != nil {
			if current.mac != "" {
				mac = current.mac
			} else {
				mac = strings.ToUpper(current.MAC.String())
			}
			builder.WriteString("=" + mac)
		}
		if config.Bridge != nil {
			builder.WriteString(",bridge=" + *config.Bridge)
		} else if current.Bridge != nil {
			builder.WriteString(",bridge=" + *current.Bridge)
		}
		if config.Firewall != nil {
			if *config.Firewall {
				builder.WriteString(",firewall=" + boolToIntString(*config.Firewall))
			}
		} else if current.Firewall != nil && *current.Firewall {
			builder.WriteString(",firewall=" + boolToIntString(*current.Firewall))
		}
		if config.Connected != nil {
			if !*config.Connected {
				builder.WriteString(",link_down=" + boolToIntString(!*config.Connected))
			}
		} else if current.Connected != nil && !*current.Connected {
			builder.WriteString(",link_down=" + boolToIntString(!*current.Connected))
		}
		if model == string(QemuNetworkModelVirtIO) {
			if config.MTU != nil {
				config.MTU.mapToApiUnsafe(&builder)
			} else if current.MTU != nil {
				current.MTU.mapToApiUnsafe(&builder)
			}
		}
		if config.MultiQueue != nil {
			if *config.MultiQueue != 0 {
				builder.WriteString(",queues=" + strconv.Itoa(int(*config.MultiQueue)))
			}
		} else if current.MultiQueue != nil && *current.MultiQueue != 0 {
			builder.WriteString(",queues=" + strconv.Itoa(int(*current.MultiQueue)))
		}
		if config.RateLimitKBps != nil {
			config.RateLimitKBps.mapToApiUnsafe(&builder)
		} else if current.RateLimitKBps != nil {
			current.RateLimitKBps.mapToApiUnsafe(&builder)
		}
		if config.NativeVlan != nil {
			if *config.NativeVlan != 0 {
				builder.WriteString(",tag=" + config.NativeVlan.String())
			}
		} else if current.NativeVlan != nil && *current.NativeVlan != 0 {
			builder.WriteString(",tag=" + current.NativeVlan.String())
		}
		if config.TaggedVlans != nil {
			vlans := config.TaggedVlans.mapToApiUnsafe()
			if vlans != "" {
				builder.WriteString(",trunks=" + vlans[1:])
			}
		} else if current.TaggedVlans != nil {
			vlans := current.TaggedVlans.mapToApiUnsafe()
			if vlans != "" {
				builder.WriteString(",trunks=" + vlans[1:])
			}
		}
		return builder.String()
	}
	// Create
	if config.Model != nil {
		model = config.Model.String()
		builder.WriteString(config.Model.String())
	}
	if config.MAC != nil {
		mac = config.MAC.String()
		if mac != "" {
			builder.WriteString("=" + strings.ToUpper(mac))
		}
	}
	if config.Bridge != nil {
		builder.WriteString(",bridge=" + *config.Bridge)
	}
	if config.Firewall != nil && *config.Firewall {
		builder.WriteString(",firewall=" + boolToIntString(*config.Firewall))
	}
	if config.Connected != nil && !*config.Connected {
		builder.WriteString(",link_down=" + boolToIntString(!*config.Connected))
	}
	if config.MTU != nil && model == string(QemuNetworkModelVirtIO) {
		config.MTU.mapToApiUnsafe(&builder)
	}
	if config.MultiQueue != nil && *config.MultiQueue != 0 {
		builder.WriteString(",queues=" + strconv.Itoa(int(*config.MultiQueue)))
	}
	if config.RateLimitKBps != nil {
		config.RateLimitKBps.mapToApiUnsafe(&builder)
	}
	if config.NativeVlan != nil && *config.NativeVlan != 0 {
		builder.WriteString(",tag=" + config.NativeVlan.String())
	}
	if config.TaggedVlans != nil {
		vlans := config.TaggedVlans.mapToApiUnsafe()
		if vlans != "" {
			builder.WriteString(",trunks=" + vlans[1:])
		}
	}
	return builder.String()
}

func (QemuNetworkInterface) mapToSDK(rawParams string) (config QemuNetworkInterface) {
	modelAndMac := strings.SplitN(rawParams, ",", 2)
	modelAndMac = strings.Split(modelAndMac[0], "=")
	var model QemuNetworkModel
	if len(modelAndMac) == 2 {
		model = QemuNetworkModel(modelAndMac[0])
		config.Model = &model
		mac, _ := net.ParseMAC(modelAndMac[1])
		config.mac = modelAndMac[1]
		config.MAC = &mac
	}
	params := splitStringOfSettings(rawParams)
	if v, isSet := params["bridge"]; isSet {
		config.Bridge = &v
	}
	if v, isSet := params["link_down"]; isSet {
		config.Connected = util.Pointer(v == "0")
	} else {
		config.Connected = util.Pointer(true)
	}
	if v, isSet := params["firewall"]; isSet {
		config.Firewall = util.Pointer(v == "1")
	} else {
		config.Firewall = util.Pointer(false)
	}
	if model == QemuNetworkModelVirtIO {
		if v, isSet := params["mtu"]; isSet {
			var mtu QemuMTU
			if v == "1" {
				mtu.Inherit = true
			} else {
				tmpMtu, _ := strconv.Atoi(v)
				mtu.Value = MTU(tmpMtu)
			}
			config.MTU = &mtu
		}
	}
	if v, isSet := params["queues"]; isSet {
		tmpQueue, _ := strconv.Atoi(v)
		config.MultiQueue = util.Pointer(QemuNetworkQueue(tmpQueue))
	}
	if v, isSet := params["rate"]; isSet {
		config.RateLimitKBps = QemuNetworkRate(0).mapToSDK(v)
	}
	if v, isSet := params["tag"]; isSet {
		tmpVlan, _ := strconv.Atoi(v)
		config.NativeVlan = util.Pointer(Vlan(tmpVlan))
	}
	if v, isSet := params["trunks"]; isSet {
		rawVlans := strings.Split(v, ";")
		vlans := make(Vlans, len(rawVlans))
		for i, e := range rawVlans {
			tmpVlan, _ := strconv.Atoi(e)
			vlans[i] = Vlan(tmpVlan)
		}
		config.TaggedVlans = &vlans
	} else {
		config.TaggedVlans = &Vlans{}
	}
	return
}

func (config QemuNetworkInterface) Validate(current *QemuNetworkInterface) error {
	if config.Delete {
		return nil
	}
	var model QemuNetworkModel
	if current != nil { // Update
		if current.Model != nil {
			model = *current.Model
		}
	} else { // Create
		if config.Bridge == nil {
			return errors.New(QemuNetworkInterface_Error_BridgeRequired)
		}
		if config.Model == nil {
			return errors.New(QemuNetworkInterface_Error_ModelRequired)
		}
	}
	// shared
	if config.Model != nil {
		if err := config.Model.Validate(); err != nil {
			return err
		}
		model = QemuNetworkModel((*config.Model).String())
	}
	if config.MTU != nil {
		if model != QemuNetworkModelVirtIO && (config.MTU.Inherit || config.MTU.Value != 0) {
			return errors.New(QemuNetworkInterface_Error_MtuNoEffect)
		}
		if err := config.MTU.Validate(); err != nil {
			return err
		}
	}
	if config.MultiQueue != nil {
		if err := config.MultiQueue.Validate(); err != nil {
			return err
		}
	}
	if config.RateLimitKBps != nil {
		if err := config.RateLimitKBps.Validate(); err != nil {
			return err
		}
	}
	if config.NativeVlan != nil {
		if err := config.NativeVlan.Validate(); err != nil {
			return err
		}
	}
	if config.TaggedVlans != nil {
		if err := config.TaggedVlans.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type QemuNetworkInterfaceID uint8

const (
	QemuNetworkInterfaceID_Error_Invalid string = "network interface ID must be in the range 0-31"

	QemuNetworkInterfaceID0  QemuNetworkInterfaceID = 0
	QemuNetworkInterfaceID1  QemuNetworkInterfaceID = 1
	QemuNetworkInterfaceID2  QemuNetworkInterfaceID = 2
	QemuNetworkInterfaceID3  QemuNetworkInterfaceID = 3
	QemuNetworkInterfaceID4  QemuNetworkInterfaceID = 4
	QemuNetworkInterfaceID5  QemuNetworkInterfaceID = 5
	QemuNetworkInterfaceID6  QemuNetworkInterfaceID = 6
	QemuNetworkInterfaceID7  QemuNetworkInterfaceID = 7
	QemuNetworkInterfaceID8  QemuNetworkInterfaceID = 8
	QemuNetworkInterfaceID9  QemuNetworkInterfaceID = 9
	QemuNetworkInterfaceID10 QemuNetworkInterfaceID = 10
	QemuNetworkInterfaceID11 QemuNetworkInterfaceID = 11
	QemuNetworkInterfaceID12 QemuNetworkInterfaceID = 12
	QemuNetworkInterfaceID13 QemuNetworkInterfaceID = 13
	QemuNetworkInterfaceID14 QemuNetworkInterfaceID = 14
	QemuNetworkInterfaceID15 QemuNetworkInterfaceID = 15
	QemuNetworkInterfaceID16 QemuNetworkInterfaceID = 16
	QemuNetworkInterfaceID17 QemuNetworkInterfaceID = 17
	QemuNetworkInterfaceID18 QemuNetworkInterfaceID = 18
	QemuNetworkInterfaceID19 QemuNetworkInterfaceID = 19
	QemuNetworkInterfaceID20 QemuNetworkInterfaceID = 20
	QemuNetworkInterfaceID21 QemuNetworkInterfaceID = 21
	QemuNetworkInterfaceID22 QemuNetworkInterfaceID = 22
	QemuNetworkInterfaceID23 QemuNetworkInterfaceID = 23
	QemuNetworkInterfaceID24 QemuNetworkInterfaceID = 24
	QemuNetworkInterfaceID25 QemuNetworkInterfaceID = 25
	QemuNetworkInterfaceID26 QemuNetworkInterfaceID = 26
	QemuNetworkInterfaceID27 QemuNetworkInterfaceID = 27
	QemuNetworkInterfaceID28 QemuNetworkInterfaceID = 28
	QemuNetworkInterfaceID29 QemuNetworkInterfaceID = 29
	QemuNetworkInterfaceID30 QemuNetworkInterfaceID = 30
	QemuNetworkInterfaceID31 QemuNetworkInterfaceID = 31

	QemuNetworkInterfaceIDMaximum QemuNetworkInterfaceID = QemuNetworkInterfaceID31
)

func (id QemuNetworkInterfaceID) String() string {
	return strconv.Itoa(int(id))
}

func (id QemuNetworkInterfaceID) Validate() error {
	if id > QemuNetworkInterfaceIDMaximum {
		return errors.New(QemuNetworkInterfaceID_Error_Invalid)
	}
	return nil
}

type QemuNetworkInterfaces map[QemuNetworkInterfaceID]QemuNetworkInterface

const QemuNetworkInterfacesAmount = uint8(QemuNetworkInterfaceIDMaximum) + 1

func (config QemuNetworkInterfaces) mapToAPI(current QemuNetworkInterfaces, params map[string]interface{}) (delete string) {
	for i, e := range config {
		if v, isSet := current[i]; isSet { // Update
			if e.Delete {
				delete += ",net" + i.String()
				continue
			}
			params["net"+i.String()] = e.mapToApi(&v)
		} else { // Create
			if e.Delete {
				continue
			}
			params["net"+i.String()] = e.mapToApi(nil)
		}
	}
	return
}

func (QemuNetworkInterfaces) mapToSDK(params map[string]interface{}) QemuNetworkInterfaces {
	interfaces := QemuNetworkInterfaces{}
	for i := uint8(0); i < QemuNetworkInterfacesAmount; i++ {
		if rawInterface, isSet := params["net"+strconv.Itoa(int(i))]; isSet {
			interfaces[QemuNetworkInterfaceID(i)] = QemuNetworkInterface{}.mapToSDK(rawInterface.(string))
		}
	}
	if len(interfaces) > 0 {
		return interfaces
	}
	return nil
}

func (config QemuNetworkInterfaces) Validate(current QemuNetworkInterfaces) error {
	for i, e := range config {
		if err := i.Validate(); err != nil {
			return err
		}
		var currentInterface *QemuNetworkInterface
		if v, isSet := current[i]; isSet {
			currentInterface = &v
		}
		if err := e.Validate(currentInterface); err != nil {
			return err
		}
	}
	return nil
}

type QemuNetworkModel string // enum

const (
	QemuNetworkModelE1000              QemuNetworkModel = "e1000"
	QemuNetworkModelE100082540em       QemuNetworkModel = "e1000-82540em"
	qemuNetworkModelE100082540em_Lower QemuNetworkModel = "e100082540em"
	QemuNetworkModelE100082544gc       QemuNetworkModel = "e1000-82544gc"
	qemuNetworkModelE100082544gc_Lower QemuNetworkModel = "e100082544gc"
	QemuNetworkModelE100082545em       QemuNetworkModel = "e1000-82545em"
	qemuNetworkModelE100082545em_Lower QemuNetworkModel = "e100082545em"
	QemuNetworkModelE1000e             QemuNetworkModel = "e1000e"
	QemuNetworkModelI82551             QemuNetworkModel = "i82551"
	QemuNetworkModelI82557b            QemuNetworkModel = "i82557b"
	QemuNetworkModelI82559er           QemuNetworkModel = "i82559er"
	QemuNetworkModelNe2kISA            QemuNetworkModel = "ne2k_isa"
	qemuNetworkModelNe2kISA_Lower      QemuNetworkModel = "ne2kisa"
	QemuNetworkModelNe2kPCI            QemuNetworkModel = "ne2k_pci"
	qemuNetworkModelNe2kPCI_Lower      QemuNetworkModel = "ne2kpci"
	QemuNetworkModelPcNet              QemuNetworkModel = "pcnet"
	QemuNetworkModelRtl8139            QemuNetworkModel = "rtl8139"
	QemuNetworkModelVirtIO             QemuNetworkModel = "virtio"
	QemuNetworkModelVmxNet3            QemuNetworkModel = "vmxnet3"
)

func (QemuNetworkModel) enumMap() map[QemuNetworkModel]QemuNetworkModel {
	return map[QemuNetworkModel]QemuNetworkModel{
		QemuNetworkModelE1000:              QemuNetworkModelE1000,
		qemuNetworkModelE100082540em_Lower: QemuNetworkModelE100082540em,
		qemuNetworkModelE100082544gc_Lower: QemuNetworkModelE100082544gc,
		qemuNetworkModelE100082545em_Lower: QemuNetworkModelE100082545em,
		QemuNetworkModelE1000e:             QemuNetworkModelE1000e,
		QemuNetworkModelI82551:             QemuNetworkModelI82551,
		QemuNetworkModelI82557b:            QemuNetworkModelI82557b,
		QemuNetworkModelI82559er:           QemuNetworkModelI82559er,
		qemuNetworkModelNe2kISA_Lower:      QemuNetworkModelNe2kISA,
		qemuNetworkModelNe2kPCI_Lower:      QemuNetworkModelNe2kPCI,
		QemuNetworkModelPcNet:              QemuNetworkModelPcNet,
		QemuNetworkModelRtl8139:            QemuNetworkModelRtl8139,
		QemuNetworkModelVirtIO:             QemuNetworkModelVirtIO,
		QemuNetworkModelVmxNet3:            QemuNetworkModelVmxNet3}
}

func (QemuNetworkModel) Error() error {
	models := QemuNetworkModel("").enumMap()
	modelsConverted := make([]string, len(models))
	var index int
	for _, e := range models {
		modelsConverted[index] = string(e)
		index++
	}
	slices.Sort(modelsConverted)
	return errors.New("qemuNetworkModel can only be one of the following values: " + strings.Join(modelsConverted, ", "))
}

// returns the model with proper dashes, underscores and capitalization
func (model QemuNetworkModel) String() string {
	models := QemuNetworkModel("").enumMap()
	if v, ok := models[QemuNetworkModel(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(string(model), "_", ""), "-", "")))]; ok {
		return string(v)
	}
	return ""
}

func (model QemuNetworkModel) Validate() error {
	if model.String() != "" {
		return nil
	}
	return QemuNetworkModel("").Error()
}

type QemuNetworkQueue uint8 // 0-64, 0 to disable

const (
	QemuNetworkQueueMaximum        QemuNetworkQueue = 64
	QemuNetworkQueue_Error_Invalid string           = "network queue must be in the range 0-64"
)

func (queue QemuNetworkQueue) Validate() error {
	if queue > QemuNetworkQueueMaximum {
		return errors.New(QemuNetworkQueue_Error_Invalid)
	}
	return nil
}

type QemuNetworkRate uint32 // 0-10240000

const (
	QemuNetworkRate_Error_Invalid string          = "network rate must be in the range 0-10240000"
	QemuNetworkRateMaximum        QemuNetworkRate = 10240000
)

// unsafe requires caller to check for nil
func (rate QemuNetworkRate) mapToApiUnsafe(builder *strings.Builder) {
	if rate == 0 {
		return
	}
	rawRate := strconv.Itoa(int(rate))
	length := len(rawRate)
	switch {
	case length > 3:
		// Insert a decimal point three places from the end
		if rate%1000 == 0 {
			builder.WriteString(",rate=" + rawRate[:length-3])
		} else {
			builder.WriteString(strings.TrimRight(",rate="+rawRate[:length-3]+"."+rawRate[length-3:], "0"))
		}
	case length > 0:
		// Prepend zeros to ensure decimal places
		prefixRate := "000" + rawRate
		builder.WriteString(strings.TrimRight(",rate=0."+prefixRate[length:], "0"))
	}
}

func (QemuNetworkRate) mapToSDK(rawRate string) *QemuNetworkRate {
	splitRate := strings.Split(rawRate, ".")
	var rate int
	switch len(splitRate) {
	case 1:
		if splitRate[0] != "0" {
			rate, _ = strconv.Atoi(splitRate[0] + "000")
		}
	case 2:
		// Pad the fractional part to ensure it has at least 3 digits
		fractional := splitRate[1] + "000"
		rate, _ = strconv.Atoi(splitRate[0] + fractional[:3])
	}
	return util.Pointer(QemuNetworkRate(rate))
}

func (rate QemuNetworkRate) Validate() error {
	if rate > QemuNetworkRateMaximum {
		return errors.New(QemuNetworkRate_Error_Invalid)
	}
	return nil
}
