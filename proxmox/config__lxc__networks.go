package proxmox

import (
	"errors"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/body"
)

// manual is programmed here
// https://github.com/proxmox/proxmox-ve-rs/blob/1811e0560cb11186aa94fe24605ce8bf7d05cc62/proxmox-ve-config/src/guest/vm.rs#L154

type LxcNetwork struct {
	Bridge        *string           `json:"bridge,omitempty"`       // Required for creation. Never nil when returned
	Connected     *bool             `json:"connected,omitempty"`    // Never nil when returned
	Firewall      *bool             `json:"firewall,omitempty"`     // Never nil when returned
	HostManaged   *bool             `json:"host_managed,omitempty"` // Only avalible in PVE 9.1 and up. Never nil from PVE 9.1 and up
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
	LxcNetwork_Error_HostManaged    = "lxc network host managed only availible from PVE 9.1 and up"
	LxcNetwork_Error_BridgeRequired = "lxc network bridge is required for creation"
	LxcNetwork_Error_NameRequired   = "lxc network name is required for creation"
)

func (config LxcNetwork) mapToApiCreate(b *strings.Builder) {
	if config.Name != nil {
		b.WriteString("name" + equal)
		b.WriteString(config.Name.String())
	}
	if config.Bridge != nil {
		b.WriteString(comma + "bridge" + equal)
		b.WriteString(*config.Bridge)
	}
	if config.Connected != nil && !(*config.Connected) {
		b.WriteString(comma + "link_down" + equal + "1")
	}
	if config.Firewall != nil && *config.Firewall {
		b.WriteString(comma + "firewall" + equal + "1")
	}
	if config.HostManaged != nil && *config.HostManaged {
		b.WriteString(comma + "host-managed" + equal + "1")
	}
	if config.IPv4 != nil {
		config.IPv4.mapToApiCreate(b)
	}
	if config.IPv6 != nil {
		config.IPv6.mapToApiCreate(b)
	}
	if config.MAC != nil {
		if mac := config.MAC.String(); mac != "" { // Returns a lowercase MAC address
			b.WriteString(comma + "hwaddr" + equal)
			b.WriteString(body.Escape(strings.ToUpper(mac)))
		}
	}
	if config.Mtu != nil && *config.Mtu != 0 {
		b.WriteString(comma + "mtu" + equal)
		b.WriteString(config.Mtu.String())
	}
	if config.NativeVlan != nil && *config.NativeVlan != 0 {
		b.WriteString(comma + "tag" + equal)
		b.WriteString(config.NativeVlan.String())
	}
	if config.RateLimitKBps != nil {
		config.RateLimitKBps.mapToAPI(b)
	}
	if config.TaggedVlans != nil {
		if v := config.TaggedVlans.string(); v != "" {
			b.WriteString(comma + "trunks" + equal)
			b.WriteString(v)
		}
	}
}

func (config LxcNetwork) mapToApiUpdate(current *LxcNetwork) string {
	var b strings.Builder
	bPtr := &b
	b.WriteString("name" + equal)
	if config.Name != nil {
		b.WriteString(config.Name.String())
	} else {
		b.WriteString(current.Name.String())
	}
	b.WriteString(comma + "bridge" + equal)
	if config.Bridge != nil {
		b.WriteString(*config.Bridge)
	} else {
		b.WriteString(*current.Bridge)
	}
	if config.Connected != nil {
		if !(*config.Connected) {
			b.WriteString(comma + "link_down" + equal + "1")
		}
	} else if !(*current.Connected) {
		b.WriteString(comma + "link_down" + equal + "1")
	}
	if config.Firewall != nil {
		if *config.Firewall {
			b.WriteString(comma + "firewall" + equal + "1")
		}
	} else if *current.Firewall {
		b.WriteString(comma + "firewall" + equal + "1")
	}
	if config.HostManaged != nil {
		if *config.HostManaged {
			b.WriteString(comma + "host-managed" + equal + "1")
		}
	} else if current.HostManaged != nil && *current.HostManaged {
		b.WriteString(comma + "host-managed" + equal + "1")
	}
	if config.IPv4 != nil {
		if current.IPv4 != nil {
			config.IPv4.mapToApiUpdate(*current.IPv4, bPtr)
		} else {
			config.IPv4.mapToApiCreate(bPtr)
		}
	} else if current.IPv4 != nil {
		current.IPv4.mapToApiCreate(bPtr)
	}
	if config.IPv6 != nil {
		if current.IPv6 != nil {
			config.IPv6.mapToApiUpdate(*current.IPv6, bPtr)
		} else {
			config.IPv6.mapToApiCreate(bPtr)
		}
	} else if current.IPv6 != nil {
		current.IPv6.mapToApiCreate(bPtr)
	}
	if config.MAC != nil {
		mac := config.MAC.String() // Returns a lowercase MAC address
		if mac != "" {
			if strings.EqualFold(mac, current.mac) { // Preserve the original case, changing causes network interface reconnect
				mac = current.mac
			} else {
				mac = strings.ToUpper(mac)
			}
			b.WriteString(comma + "hwaddr" + equal)
			b.WriteString(body.Escape(mac))
		}
	} else {
		b.WriteString(comma + "hwaddr" + equal)
		b.WriteString(body.Escape(current.mac))
	}
	if config.Mtu != nil {
		if *config.Mtu != 0 {
			b.WriteString(comma + "mtu" + equal)
			b.WriteString(config.Mtu.String())
		}
	} else if current.Mtu != nil && *current.Mtu != 0 {
		b.WriteString(comma + "mtu" + equal)
		b.WriteString(current.Mtu.String())
	}
	if config.NativeVlan != nil {
		if *config.NativeVlan != 0 {
			b.WriteString(comma + "tag" + equal)
			b.WriteString(config.NativeVlan.String())
		}
	} else if current.NativeVlan != nil && *current.NativeVlan != 0 {
		b.WriteString(comma + "tag" + equal)
		b.WriteString(current.NativeVlan.String())
	}
	if config.RateLimitKBps != nil {
		config.RateLimitKBps.mapToAPI(bPtr)
	} else if current.RateLimitKBps != nil {
		current.RateLimitKBps.mapToAPI(bPtr)
	}
	if config.TaggedVlans != nil {
		if v := config.TaggedVlans.string(); v != "" {
			b.WriteString(comma + "trunks" + equal)
			b.WriteString(v)
		}
	} else if current.TaggedVlans != nil {
		if v := current.TaggedVlans.string(); v != "" {
			b.WriteString(comma + "trunks" + equal)
			b.WriteString(v)
		}
	}
	return b.String()
}

func (config LxcNetwork) Validate(current *LxcNetwork, version Version) error {
	if current == nil {
		return config.validateCreate(version.Encode())
	}
	return config.validate(version.Encode())
}

func (config LxcNetwork) validate(version EncodedVersion) error {
	if config.Delete {
		return nil
	}
	if config.HostManaged != nil && version < version_9_1_0 {
		return errors.New(LxcNetwork_Error_HostManaged)
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

func (config LxcNetwork) validateCreate(version EncodedVersion) error {
	if config.Delete {
		return nil // nothing to validate
	}
	if config.Bridge == nil || *config.Bridge == "" {
		return errors.New(LxcNetwork_Error_BridgeRequired)
	}
	if config.Name == nil {
		return errors.New(LxcNetwork_Error_NameRequired)
	}
	return config.validate(version)
}

type LxcNetworks map[LxcNetworkID]LxcNetwork

const (
	LxcNetworksAmount               = 16
	LxcNetworks_Error_Amount        = ""
	LxcNetworks_Error_DuplicateName = "lxc network name must be unique across all networks"
)

func (config LxcNetworks) mapToApiCreate(b *strings.Builder) {
	for id, network := range config {
		if network.Delete {
			continue // nothing to delete
		}
		b.WriteString("&" + lxcPrefixApiKeyNetwork)
		b.WriteString(id.String())
		b.WriteByte('=')
		network.mapToApiCreate(b)
	}
}

func (config LxcNetworks) mapToApiUpdate(current LxcNetworks, b, delete *strings.Builder) {
	for id, network := range config {
		if v, isSet := current[id]; isSet { // Update
			if network.Delete {
				delete.WriteString("," + lxcPrefixApiKeyNetwork)
				delete.WriteString(id.String())
				continue
			}
			currentBody := v.mapToApiUpdate(&v)
			updatedBody := network.mapToApiUpdate(&v)
			if currentBody != updatedBody {
				b.WriteString("&" + lxcPrefixApiKeyNetwork)
				b.WriteString(id.String())
				b.WriteByte('=')
				b.WriteString(updatedBody)
			}
		} else { // Create
			if network.Delete {
				continue // nothing to delete
			}
			b.WriteString("&" + lxcPrefixApiKeyNetwork)
			b.WriteString(id.String())
			b.WriteByte('=')
			network.mapToApiCreate(b)
		}
	}
}

func (config LxcNetworks) Validate(current LxcNetworks, version Version) error {
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
			if err = network.validate(version.Encode()); err != nil {
				return err
			}
		} else {
			if err = network.validateCreate(version.Encode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// LxcNetworkID
//
//	const (
//		LxcNetworkID0
//		...
//		LxcNetworkID15
//
// )
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

func (raw *rawConfigLXC) GetNetworks() LxcNetworks {
	nets := LxcNetworks{}
	for i := 0; i <= LxcNetworkIdMaximum; i++ {
		if v, isSet := raw.a[lxcPrefixApiKeyNetwork+strconv.Itoa(i)]; isSet {
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
			if v, isSet := settings["link_down"]; isSet {
				connected = v != "1"
			}
			if v, isSet := settings["firewall"]; isSet {
				firewall = v == "1"
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
			if raw.version >= version_9_1_0 {
				var hostManaged bool = false
				if v, isSet := settings["host-managed"]; isSet {
					hostManaged = v == "1"
				}
				network.HostManaged = &hostManaged
			}
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
					ipv4.Address = new(IPv4CIDR(v))
				}
			}
			if v, isSet := settings["gw"]; isSet {
				ipSet = true
				ipv4.Gateway = new(IPv4Address(v))
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
					ipv6.Address = new(IPv6CIDR(v))
				}
			}
			if v, isSet := settings["gw6"]; isSet {
				ipSet = true
				ipv6.Gateway = new(IPv6Address(v))
			}
			if ipSet {
				network.IPv6 = &ipv6
			}
			if v, isSet := settings["mtu"]; isSet {
				mtu, _ := strconv.Atoi(v)
				network.Mtu = new(MTU(mtu))
			}
			if v, isSet := settings["tag"]; isSet {
				tag, _ := strconv.Atoi(v)
				network.NativeVlan = new(Vlan(tag))
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
				network.TaggedVlans = new(taggedVlans)
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

func (config LxcIPv4) mapToApiCreate(b *strings.Builder) {
	if config.DHCP {
		b.WriteString(comma + "ip" + equal + "dhcp")
		return
	}
	if config.Manual {
		b.WriteString(comma + "ip" + equal + "manual")
		return
	}
	if config.Address != nil {
		if v := config.Address.String(); v != "" {
			b.WriteString(comma + "ip" + equal)
			b.WriteString(body.Escape(v))
		}
	}
	if config.Gateway != nil {
		if v := config.Gateway.String(); v != "" {
			b.WriteString(comma + "gw" + equal)
			b.WriteString(v)
		}
	}
}

func (config LxcIPv4) mapToApiUpdate(current LxcIPv4, b *strings.Builder) {
	combined := config.combine(current) // Combine the current and new config to preserve settings not being updated
	if combined.DHCP {
		b.WriteString(comma + "ip" + equal + "dhcp")
		return
	}
	if combined.Manual {
		b.WriteString(comma + "ip" + equal + "manual")
		return
	}
	if combined.Address != nil {
		if v := combined.Address.String(); v != "" {
			b.WriteString(comma + "ip" + equal)
			b.WriteString(strings.ReplaceAll(v, "/", body.Slash))
		}
	}
	if combined.Gateway != nil {
		if v := combined.Gateway.String(); v != "" {
			b.WriteString(comma + "gw" + equal)
			b.WriteString(v)
		}
	}
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

func (config LxcIPv6) mapToApiCreate(b *strings.Builder) {
	if config.DHCP {
		b.WriteString(comma + "ip6" + equal + "dhcp")
		return
	}
	if config.SLAAC {
		b.WriteString(comma + "ip6" + equal + "auto")
		return
	}
	if config.Manual {
		b.WriteString(comma + "ip6" + equal + "manual")
		return
	}
	if config.Address != nil {
		if v := config.Address.String(); v != "" {
			b.WriteString(comma + "ip6" + equal)
			b.WriteString(body.Escape(v))
		}
	}
	if config.Gateway != nil {
		if v := config.Gateway.String(); v != "" {
			b.WriteString(comma + "gw6" + equal)
			b.WriteString(body.Escape(v))
		}
	}
}

func (config LxcIPv6) mapToApiUpdate(current LxcIPv6, b *strings.Builder) {
	combined := config.combine(current)
	if combined.DHCP {
		b.WriteString(comma + "ip6" + equal + "dhcp")
		return
	}
	if combined.Manual {
		b.WriteString(comma + "ip6" + equal + "manual")
		return
	}
	if combined.SLAAC {
		b.WriteString(comma + "ip6" + equal + "auto")
		return
	}
	if combined.Address != nil {
		if v := combined.Address.String(); v != "" {
			b.WriteString(comma + "ip6" + equal)
			b.WriteString(body.Escape(v))
		}
	}
	if combined.Gateway != nil {
		if v := combined.Gateway.String(); v != "" {
			b.WriteString(comma + "gw6" + equal)
			b.WriteString(body.Escape(v))
		}
	}
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
