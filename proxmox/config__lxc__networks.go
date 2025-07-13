package proxmox

import (
	"errors"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

// manual is programmed here
// https://github.com/proxmox/proxmox-ve-rs/blob/1811e0560cb11186aa94fe24605ce8bf7d05cc62/proxmox-ve-config/src/guest/vm.rs#L154

type LxcNetwork struct {
	Bridge        *string           `json:"bridge,omitempty"`    // Required for creation. Never nil when returned
	Connected     *bool             `json:"connected,omitempty"` // Never nil when returned
	Firewall      *bool             `json:"firewall,omitempty"`  // Never nil when returned
	IPv4          *LxcIPv4          `json:"ipv4,omitempty"`
	IPv6          *LxcIPv6          `json:"ipv6,omitempty"`
	MAC           *net.HardwareAddr `json:"mac,omitempty"` // Never nil when returned
	Mtu           *MTU              `json:"mtu,omitempty"`
	Name          *LxcNetworkName   `json:"name,omitempty"` // Required for creation. Never nil when returned
	NativeVlan    *Vlan             `json:"native_vlan,omitempty"`
	RateLimitKBps *GuestNetworkRate `json:"rate,omitempty"`
	TaggedVlans   *Vlans            `json:"tagged_vlans,omitempty"`
	Delete        bool              `json:"delete,omitempty"`
	mac           string
}

const (
	LxcNetwork_Error_BridgeRequired = "lxc network bridge is required for creation"
	LxcNetwork_Error_NameRequired   = "lxc network name is required for creation"
)

func (config LxcNetwork) mapToApiCreate() string {
	var settings string
	if config.Name != nil {
		settings += "name=" + config.Name.String()
	}
	if config.Bridge != nil {
		settings += ",bridge=" + *config.Bridge
	}
	if config.Connected != nil && !(*config.Connected) {
		settings += ",link_down=1"
	}
	if config.Firewall != nil && *config.Firewall {
		settings += ",firewall=1"
	}
	if config.IPv4 != nil {
		settings += config.IPv4.mapToApiCreate()
	}
	if config.IPv6 != nil {
		settings += config.IPv6.mapToApiCreate()
	}
	if config.MAC != nil {
		mac := config.MAC.String() // Returns a lowercase MAC address
		if mac != "" {
			if mac == strings.ToLower(config.mac) { // Preserve the original case, changing causes network interface reconnect
				mac = config.mac
			} else {
				mac = strings.ToUpper(mac)
			}
			settings += ",hwaddr=" + mac
		}
	}
	if config.Mtu != nil && *config.Mtu != 0 {
		settings += ",mtu=" + config.Mtu.String()
	}
	if config.NativeVlan != nil && *config.NativeVlan != 0 {
		settings += ",tag=" + config.NativeVlan.String()
	}
	if config.RateLimitKBps != nil {
		settings += config.RateLimitKBps.mapToAPI()
	}
	if config.TaggedVlans != nil {
		if v := config.TaggedVlans.string(); v != "" {
			settings += ",trunks=" + v
		}
	}
	return settings
}

func (config LxcNetwork) mapToApiUpdate(current LxcNetwork) string {
	var settings string
	if config.Name != nil {
		settings += "name=" + config.Name.String()
	} else if current.Name != nil {
		settings += "name=" + current.Name.String()
	}
	if config.Bridge != nil {
		settings += ",bridge=" + *config.Bridge
	} else if current.Bridge != nil {
		settings += ",bridge=" + *current.Bridge
	}
	if config.Connected != nil {
		if !*config.Connected {
			settings += ",link_down=1"
		}
	} else if current.Connected != nil && !(*current.Connected) {
		settings += ",link_down=1"
	}
	if config.Firewall != nil {
		if *config.Firewall {
			settings += ",firewall=1"
		}
	} else if current.Firewall != nil && *current.Firewall {
		settings += ",firewall=1"
	}
	if config.IPv4 != nil {
		if current.IPv4 != nil {
			settings += config.IPv4.mapToApiUpdate(*current.IPv4)
		} else {
			settings += config.IPv4.mapToApiCreate()
		}
	} else if current.IPv4 != nil {
		settings += current.IPv4.mapToApiCreate()
	}
	if config.IPv6 != nil {
		if current.IPv6 != nil {
			settings += config.IPv6.mapToApiUpdate(*current.IPv6)
		} else {
			settings += config.IPv6.mapToApiCreate()
		}
	} else if current.IPv6 != nil {
		settings += current.IPv6.mapToApiCreate()
	}
	if config.MAC != nil {
		mac := config.MAC.String() // Returns a lowercase MAC address
		if mac != "" {
			if mac == strings.ToLower(config.mac) { // Preserve the original case, changing causes network interface reconnect
				mac = config.mac
			} else {
				mac = strings.ToUpper(mac)
			}
			settings += ",hwaddr=" + mac
		}
	} else if current.MAC != nil {
		settings += ",hwaddr=" + current.mac
	}
	if config.Mtu != nil {
		if *config.Mtu != 0 {
			settings += ",mtu=" + config.Mtu.String()
		}
	} else if current.Mtu != nil && *current.Mtu != 0 {
		settings += ",mtu=" + current.Mtu.String()
	}
	if config.NativeVlan != nil {
		if *config.NativeVlan != 0 {
			settings += ",tag=" + config.NativeVlan.String()
		}
	} else if current.NativeVlan != nil && *current.NativeVlan != 0 {
		settings += ",tag=" + current.NativeVlan.String()
	}
	if config.RateLimitKBps != nil {
		settings += config.RateLimitKBps.mapToAPI()
	} else if current.RateLimitKBps != nil {
		settings += current.RateLimitKBps.mapToAPI()
	}
	if config.TaggedVlans != nil {
		if v := config.TaggedVlans.string(); v != "" {
			settings += ",trunks=" + v
		}
	} else if current.TaggedVlans != nil {
		if v := current.TaggedVlans.string(); v != "" {
			settings += ",trunks=" + v
		}
	}
	return settings
}

func (config LxcNetwork) Validate(current *LxcNetwork) error {
	if current == nil {
		return config.validateCreate()
	}
	return config.validate()
}

func (config LxcNetwork) validate() error {
	if config.Delete {
		return nil
	}
	if config.IPv4 != nil {
		if err := config.IPv4.Validate(); err != nil {
			return err
		}
	}
	if config.IPv6 != nil {
		if err := config.IPv6.Validate(); err != nil {
			return err
		}
	}
	if config.Mtu != nil {
		if err := config.Mtu.Validate(); err != nil {
			return err
		}
	}
	if config.Name != nil {
		if err := config.Name.Validate(); err != nil {
			return err
		}
	}
	if config.NativeVlan != nil {
		if err := config.NativeVlan.Validate(); err != nil {
			return err
		}
	}
	if config.RateLimitKBps != nil {
		if err := config.RateLimitKBps.Validate(); err != nil {
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

func (config LxcNetwork) validateCreate() error {
	if config.Delete {
		return nil // nothing to validate
	}
	if config.Bridge == nil || *config.Bridge == "" {
		return errors.New(LxcNetwork_Error_BridgeRequired)
	}
	if config.Name == nil {
		return errors.New(LxcNetwork_Error_NameRequired)
	}
	return config.validate()
}

type LxcNetworks map[LxcNetworkID]LxcNetwork

const (
	LxcNetworksAmount               = 16
	LxcNetworks_Error_Amount        = ""
	LxcNetworks_Error_DuplicateName = "lxc network name must be unique across all networks"
)

func (config LxcNetworks) mapToApiCreate(params map[string]any) {
	for id, network := range config {
		if network.Delete {
			continue // nothing to delete
		}
		params[lxcPrefixApiKeyNetwork+id.String()] = network.mapToApiCreate()
	}
}

func (config LxcNetworks) mapToApiUpdate(current LxcNetworks, params map[string]any) (delete string) {
	for id, network := range config {
		if v, isSet := current[id]; isSet {
			if network.Delete {
				delete = "," + lxcPrefixApiKeyNetwork + id.String()
				continue
			}
			newNetwork := network.mapToApiUpdate(v)
			if newNetwork != v.mapToApiCreate() {
				params[lxcPrefixApiKeyNetwork+id.String()] = newNetwork
			}
			continue
		}
		if network.Delete {
			continue // nothing to delete
		}
		params[lxcPrefixApiKeyNetwork+id.String()] = network.mapToApiCreate()
	}
	return delete
}

func (config LxcNetworks) Validate(current LxcNetworks) error {
	if len(config) > LxcNetworksAmount {
		return errors.New(LxcNetworks_Error_Amount)
	}

	transformedState := map[LxcNetworkID]LxcNetworkName{} // The transformed state of config being applied over current
	for k, v := range current {
		if v.Name != nil {
			transformedState[k] = *v.Name
		}
	}
	for k, v := range config {
		if v.Delete { // Remove all networks marked for deletion
			delete(transformedState, k)
		} else if v.Name != nil { // Add or Overwrite existing networks
			transformedState[k] = *v.Name
		}
	}
	uniqueNames := make(map[LxcNetworkName]struct{}, len(transformedState))
	for _, v := range transformedState {
		if _, duplicate := uniqueNames[v]; duplicate {
			return errors.New(LxcNetworks_Error_DuplicateName)
		}
		uniqueNames[v] = struct{}{}
	}

	var err error
	for id, network := range config {
		if err = id.Validate(); err != nil {
			return err
		}
		if _, isSet := current[id]; isSet {
			if err = network.validate(); err != nil {
				return err
			}
		} else {
			if err = network.validateCreate(); err != nil {
				return err
			}
		}
	}
	return nil
}

type LxcNetworkID uint8

const (
	LxcNetworkID0  LxcNetworkID = 0
	LxcNetworkID1  LxcNetworkID = 1
	LxcNetworkID2  LxcNetworkID = 2
	LxcNetworkID3  LxcNetworkID = 3
	LxcNetworkID4  LxcNetworkID = 4
	LxcNetworkID5  LxcNetworkID = 5
	LxcNetworkID6  LxcNetworkID = 6
	LxcNetworkID7  LxcNetworkID = 7
	LxcNetworkID8  LxcNetworkID = 8
	LxcNetworkID9  LxcNetworkID = 9
	LxcNetworkID10 LxcNetworkID = 10
	LxcNetworkID11 LxcNetworkID = 11
	LxcNetworkID12 LxcNetworkID = 12
	LxcNetworkID13 LxcNetworkID = 13
	LxcNetworkID14 LxcNetworkID = 14
	LxcNetworkID15 LxcNetworkID = 15
)

const (
	LxcNetworkIdMaximum        = 15
	LxcNetworkID_Error_Invalid = "lxc network id must be between 0 and 15"
)

func (id LxcNetworkID) String() string { return strconv.Itoa(int(id)) } // String is for fmt.Stringer.

func (id LxcNetworkID) Validate() error {
	if id > LxcNetworkIdMaximum {
		return errors.New(LxcNetworkID_Error_Invalid)
	}
	return nil
}

func (raw RawConfigLXC) Networks() LxcNetworks {
	nets := LxcNetworks{}
	for i := 0; i <= LxcNetworkIdMaximum; i++ {
		if v, isSet := raw[lxcPrefixApiKeyNetwork+strconv.Itoa(i)]; isSet {
			var bridge string
			var connected bool = true
			var firewall bool
			var name LxcNetworkName
			var mac net.HardwareAddr
			var macOriginal string
			settings := splitStringOfSettings(v.(string))
			if v, isSet := settings["bridge"]; isSet {
				bridge = v
			}
			if v, isSet := settings["link_down"]; isSet && v == "1" {
				connected = false
			}
			if v, isSet := settings["firewall"]; isSet && v == "1" {
				firewall = true
			}
			if v, isSet := settings["name"]; isSet {
				name = LxcNetworkName(v)
			}
			if v, isSet := settings["hwaddr"]; isSet {
				macOriginal = v // Store the original MAC address to preserve case
				mac, _ = net.ParseMAC(v)
			}
			network := LxcNetwork{
				Bridge:    &bridge,
				Connected: &connected,
				Firewall:  &firewall,
				MAC:       &mac,
				Name:      &name,
				mac:       macOriginal}
			var ipSet bool
			var ipv4 LxcIPv4
			if v, isSet := settings["ip"]; isSet {
				ipSet = true
				switch v {
				case "dhcp":
					ipv4.DHCP = true
				case "manual":
					ipv4.Manual = true
				default:
					ipv4.Address = util.Pointer(IPv4CIDR(v))
				}
			}
			if v, isSet := settings["gw"]; isSet {
				ipSet = true
				ipv4.Gateway = util.Pointer(IPv4Address(v))
			}
			if ipSet {
				network.IPv4 = &ipv4
			}
			ipSet = false // Reuse flag for IPv6 settings
			var ipv6 LxcIPv6
			if v, isSet := settings["ip6"]; isSet {
				ipSet = true
				switch v {
				case "dhcp":
					ipv6.DHCP = true
				case "auto":
					ipv6.SLAAC = true
				case "manual":
					ipv6.Manual = true
				default:
					ipv6.Address = util.Pointer(IPv6CIDR(v))
				}
			}
			if v, isSet := settings["gw6"]; isSet {
				ipSet = true
				ipv6.Gateway = util.Pointer(IPv6Address(v))
			}
			if ipSet {
				network.IPv6 = &ipv6
			}
			if v, isSet := settings["mtu"]; isSet {
				mtu, _ := strconv.Atoi(v)
				network.Mtu = util.Pointer(MTU(mtu))
			}
			if v, isSet := settings["tag"]; isSet {
				tag, _ := strconv.Atoi(v)
				network.NativeVlan = util.Pointer(Vlan(tag))
			}
			if v, isSet := settings["rate"]; isSet {
				network.RateLimitKBps = GuestNetworkRate(0).mapToSDK(v)
			}
			if v, isSet := settings["trunks"]; isSet {
				// Split the string by semicolon and convert to Vlans
				vlanStrings := strings.Split(v, ";")
				taggedVlans := make(Vlans, len(vlanStrings))
				for i, vlanStr := range vlanStrings {
					vlan, _ := strconv.Atoi(vlanStr)
					taggedVlans[i] = Vlan(vlan)
				}
				slices.Sort(taggedVlans)
				network.TaggedVlans = util.Pointer(taggedVlans)
			}
			nets[LxcNetworkID(i)] = network
		}
	}
	return nets
}

// max len 16, must be unique across all networks
// regex ^(?!\.\.)[a-zA-Z0-9_.-]{2,16}$
type LxcNetworkName string

var regexLxcNetworkName = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

const (
	LxcNetworkName_Error_Invalid       = `lxc network name must match regex: ^(?!\.\.)[a-zA-Z0-9_.-]{2,16}$`
	LxcNetworkName_Error_LengthMinimum = "lxc network name must be at least 2 characters long"
	LxcNetworkName_Error_LengthMaximum = "lxc network name must be at most 16 characters long"
)

func (name LxcNetworkName) String() string { return string(name) } // String is for fmt.Stringer.

func (name LxcNetworkName) Validate() error {
	if len(name) < 2 {
		return errors.New(LxcNetworkName_Error_LengthMinimum)
	}
	if len(name) > 16 {
		return errors.New(LxcNetworkName_Error_LengthMaximum)
	}
	if name == ".." {
		return errors.New(LxcNetworkName_Error_Invalid)
	}
	if !regexLxcNetworkName.Match([]byte(name)) {
		return errors.New(LxcNetworkName_Error_Invalid)
	}
	return nil
}

type LxcIPv4 struct {
	Address *IPv4CIDR    `json:"address,omitempty"`
	Gateway *IPv4Address `json:"gateway,omitempty"`
	DHCP    bool         `json:"dhcp,omitempty"`
	Manual  bool         `json:"manual,omitempty"`
}

const (
	LxcIPv4_Error_MutuallyExclusive        = "lxc IPv4 Manual and DHCP are mutually exclusive"
	LxcIPv4_Error_MutuallyExclusiveAddress = "lxc IPv4 Address and DHCP/Manual are mutually exclusive"
	LxcIPv4_Error_MutuallyExclusiveGateway = "lxc IPv4 Gateway and DHCP/Manual are mutually exclusive"
)

func (config LxcIPv4) combine(current LxcIPv4) LxcIPv4 {
	combined := LxcIPv4{
		Address: current.Address,
		DHCP:    config.DHCP,
		Gateway: current.Gateway,
		Manual:  config.Manual}
	if config.Address != nil {
		combined.Address = config.Address
	}
	if config.Gateway != nil {
		combined.Gateway = config.Gateway
	}
	return combined
}

func (config LxcIPv4) mapToApiCreate() string {
	if config.DHCP {
		return ",ip=dhcp"
	}
	if config.Manual {
		return ",ip=manual"
	}
	var settings string
	if config.Address != nil {
		if v := config.Address.String(); v != "" {
			settings += ",ip=" + v
		}
	}
	if config.Gateway != nil {
		if v := config.Gateway.String(); v != "" {
			return settings + ",gw=" + v
		}
	}
	return settings
}

func (config LxcIPv4) mapToApiUpdate(current LxcIPv4) string {
	combined := config.combine(current) // Combine the current and new config to preserve settings not being updated
	if combined.DHCP {
		return ",ip=dhcp"
	}
	if combined.Manual {
		return ",ip=manual"
	}
	var settings string
	if combined.Address != nil {
		if v := combined.Address.String(); v != "" {
			settings += ",ip=" + v
		}
	}
	if combined.Gateway != nil {
		if v := combined.Gateway.String(); v != "" {
			return settings + ",gw=" + v
		}
	}
	return settings
}

func (ipv4 LxcIPv4) Validate() error {
	var mutuallyExclusive bool
	if ipv4.DHCP {
		mutuallyExclusive = true
	}
	if ipv4.Manual {
		if mutuallyExclusive {
			return errors.New(LxcIPv4_Error_MutuallyExclusive)
		}
		mutuallyExclusive = true
	}
	if ipv4.Address != nil {
		if mutuallyExclusive {
			return errors.New(LxcIPv4_Error_MutuallyExclusiveAddress)
		}
		if err := ipv4.Address.Validate(); err != nil {
			return err
		}
	}
	if ipv4.Gateway != nil {
		if mutuallyExclusive {
			return errors.New(LxcIPv4_Error_MutuallyExclusiveGateway)
		}
		if err := ipv4.Gateway.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type LxcIPv6 struct {
	Address *IPv6CIDR    `json:"address,omitempty"`
	Gateway *IPv6Address `json:"gateway,omitempty"`
	DHCP    bool         `json:"dhcp,omitempty"`
	SLAAC   bool         `json:"slaac,omitempty"`
	Manual  bool         `json:"manual,omitempty"`
}

const (
	LxcIPv6_Error_MutuallyExclusive        = "lxc IPv6 DHCP/Manual/SLAAC are mutually exclusive"
	LxcIPv6_Error_MutuallyExclusiveAddress = "lxc IPv6 Address and DHCP/SLAAC/Manual are mutually exclusive"
	LxcIPv6_Error_MutuallyExclusiveGateway = "lxc IPv6 Gateway and DHCP/SLAAC/Manual are mutually exclusive"
)

func (config LxcIPv6) combine(current LxcIPv6) LxcIPv6 {
	combined := LxcIPv6{
		Address: current.Address,
		DHCP:    config.DHCP,
		Gateway: current.Gateway,
		Manual:  config.Manual,
		SLAAC:   config.SLAAC}
	if config.Address != nil {
		combined.Address = config.Address
	}
	if config.Gateway != nil {
		combined.Gateway = config.Gateway
	}
	return combined
}

func (config LxcIPv6) mapToApiCreate() string {
	if config.DHCP {
		return ",ip6=dhcp"
	}
	if config.SLAAC {
		return ",ip6=auto"
	}
	if config.Manual {
		return ",ip6=manual"
	}
	var settings string
	if config.Address != nil {
		if v := config.Address.String(); v != "" {
			settings += ",ip6=" + v
		}
	}
	if config.Gateway != nil {
		if v := config.Gateway.String(); v != "" {
			return settings + ",gw6=" + v
		}
	}
	return settings
}

func (config LxcIPv6) mapToApiUpdate(current LxcIPv6) string {
	combined := config.combine(current)
	if combined.DHCP {
		return ",ip6=dhcp"
	}
	if combined.Manual {
		return ",ip6=manual"
	}
	if combined.SLAAC {
		return ",ip6=auto"
	}
	var settings string
	if combined.Address != nil {
		if v := combined.Address.String(); v != "" {
			settings += ",ip6=" + v
		}
	}
	if combined.Gateway != nil {
		if v := combined.Gateway.String(); v != "" {
			return settings + ",gw6=" + v
		}
	}
	return settings
}

func (ipv6 LxcIPv6) Validate() error {
	var mutuallyExclusive bool
	if ipv6.DHCP {
		mutuallyExclusive = true
	}
	if ipv6.Manual {
		if mutuallyExclusive {
			return errors.New(LxcIPv6_Error_MutuallyExclusive)
		}
		mutuallyExclusive = true
	}
	if ipv6.SLAAC {
		if mutuallyExclusive {
			return errors.New(LxcIPv6_Error_MutuallyExclusive)
		}
		mutuallyExclusive = true
	}
	if ipv6.Address != nil {
		if mutuallyExclusive {
			return errors.New(LxcIPv6_Error_MutuallyExclusiveAddress)
		}
		if err := ipv6.Address.Validate(); err != nil {
			return err
		}
	}
	if ipv6.Gateway != nil {
		if mutuallyExclusive {
			return errors.New(LxcIPv6_Error_MutuallyExclusiveGateway)
		}
		if err := ipv6.Gateway.Validate(); err != nil {
			return err
		}
	}
	return nil
}
