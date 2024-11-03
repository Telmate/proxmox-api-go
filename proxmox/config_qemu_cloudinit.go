package proxmox

import (
	"crypto"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

var regexMultipleNewlineEncoded = regexp.MustCompile(`(%0A)+`)
var regexMultipleSpaces = regexp.MustCompile(`\s+`)
var regexMultipleSpacesEncoded = regexp.MustCompile(`(%20)+`)

// URL encodes the ssh keys
func sshKeyUrlDecode(encodedKeys string) (keys []crypto.PublicKey) {
	encodedKeys = regexMultipleSpacesEncoded.ReplaceAllString(encodedKeys, "%20")
	encodedKeys = strings.TrimSuffix(encodedKeys, "%0A")
	encodedKeys = regexMultipleNewlineEncoded.ReplaceAllString(encodedKeys, "%0A")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%2B", "+")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%40", "@")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%3D", "=")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%3A", ":")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%20", " ")
	encodedKeys = strings.ReplaceAll(encodedKeys, "%2F", "/")
	for _, key := range strings.Split(encodedKeys, "%0A") {
		keys = append(keys, key)
	}
	return
}

// URL encodes the ssh keys
func sshKeyUrlEncode(keys []crypto.PublicKey) (encodedKeys string) {
	for _, key := range keys {
		tmpKey := regexMultipleSpaces.ReplaceAllString(key.(string), " ")
		tmpKey = url.PathEscape(tmpKey + "\n")
		tmpKey = strings.ReplaceAll(tmpKey, "+", "%2B")
		tmpKey = strings.ReplaceAll(tmpKey, "@", "%40")
		tmpKey = strings.ReplaceAll(tmpKey, "=", "%3D")
		encodedKeys += strings.ReplaceAll(tmpKey, ":", "%3A")
	}
	return
}

type CloudInit struct {
	Custom            *CloudInitCustom           `json:"cicustom,omitempty"`
	DNS               *GuestDNS                  `json:"dns,omitempty"`
	NetworkInterfaces CloudInitNetworkInterfaces `json:"ipconfig,omitempty"`
	PublicSSHkeys     *[]crypto.PublicKey        `json:"sshkeys,omitempty"`
	UpgradePackages   *bool                      `json:"ciupgrade,omitempty"`
	UserPassword      *string                    `json:"userpassword,omitempty"` // TODO custom type
	Username          *string                    `json:"username,omitempty"`     // TODO custom type
}

const CloudInit_Error_UpgradePackagesPre8 = "upgradePackages is only available in version 8 and above"

func (config CloudInit) mapToAPI(current *CloudInit, params map[string]interface{}, version Version) (delete string) {
	if current != nil { // Update
		if config.Custom != nil {
			params["cicustom"] = config.Custom.mapToAPI(current.Custom)
		}
		if config.Username != nil {
			tmp := *config.Username
			if tmp != "" {
				params["ciuser"] = *config.Username
			} else {
				delete += ",ciuser"
			}
		}
		if config.UserPassword != nil && *config.UserPassword == "" {
			delete += ",cipassword"
		}
		if config.DNS != nil {
			if config.DNS.SearchDomain != nil {
				if *config.DNS.SearchDomain != "" {
					params["searchdomain"] = *config.DNS.SearchDomain
				} else {
					delete += ",searchdomain"
				}
			}
			if config.DNS.NameServers != nil {
				if len(*config.DNS.NameServers) > 0 {
					var nameservers string
					for _, ns := range *config.DNS.NameServers {
						nameservers += " " + ns.String()
					}
					params["nameserver"] = nameservers[1:]
				} else {
					delete += ",nameserver"
				}
			}
		}
		delete += config.NetworkInterfaces.mapToAPI(current.NetworkInterfaces, params)
		if config.PublicSSHkeys != nil {
			if len(*config.PublicSSHkeys) > 0 {
				params["sshkeys"] = sshKeyUrlEncode(*config.PublicSSHkeys)
			} else {
				delete += ",sshkeys"
			}
		}
	} else { // Create
		if config.Custom != nil {
			params["cicustom"] = config.Custom.mapToAPI(nil)
		}
		if config.Username != nil && *config.Username != "" {
			params["ciuser"] = *config.Username
		}
		if config.DNS != nil {
			if config.DNS.SearchDomain != nil && *config.DNS.SearchDomain != "" {
				params["searchdomain"] = *config.DNS.SearchDomain
			}
			if config.DNS.NameServers != nil && len(*config.DNS.NameServers) > 0 {
				var nameservers string
				for _, ns := range *config.DNS.NameServers {
					nameservers += " " + ns.String()
				}
				params["nameserver"] = nameservers[1:]
			}
		}
		config.NetworkInterfaces.mapToAPI(nil, params)
		if config.PublicSSHkeys != nil && len(*config.PublicSSHkeys) > 0 {
			params["sshkeys"] = sshKeyUrlEncode(*config.PublicSSHkeys)
		}
	}
	// Shared
	if config.UpgradePackages != nil && !version.Smaller(Version{Major: 8}) {
		params["ciupgrade"] = Btoi(*config.UpgradePackages)
	}
	if config.UserPassword != nil && *config.UserPassword != "" {
		params["cipassword"] = *config.UserPassword
	}
	return
}

func (CloudInit) mapToSDK(params map[string]interface{}) *CloudInit {
	ci := CloudInit{}
	var set bool
	if v, isSet := params["cicustom"]; isSet {
		ci.Custom = CloudInitCustom{}.mapToSDK(v.(string))
		set = true
	}
	if v, isSet := params["cipassword"]; isSet {
		ci.UserPassword = util.Pointer(v.(string))
		set = true
	}
	if v, isSet := params["ciupgrade"]; isSet {
		ci.UpgradePackages = util.Pointer(Itob(int(v.(float64))))
		set = true
	}
	if v, isSet := params["ciuser"]; isSet {
		tmp := v.(string)
		if tmp != "" && tmp != " " {
			ci.Username = &tmp
			set = true
		}
	}
	if v, isSet := params["sshkeys"]; isSet {
		ci.PublicSSHkeys = util.Pointer(sshKeyUrlDecode(v.(string)))
		set = true
	}
	var dnsSet bool
	var nameservers []netip.Addr
	if v, isSet := params["nameserver"]; isSet {
		tmp := strings.Split(v.(string), " ")
		nameservers = make([]netip.Addr, len(tmp))
		for i, e := range tmp {
			nameservers[i], _ = netip.ParseAddr(e)
		}
		dnsSet = true
	}
	var domain string
	if v, isSet := params["searchdomain"]; isSet {
		if len(v.(string)) > 1 {
			domain = v.(string)
			dnsSet = true
		}
	}
	if dnsSet {
		ci.DNS = &GuestDNS{
			SearchDomain: &domain,
			NameServers:  &nameservers,
		}
		set = true
	}
	ci.NetworkInterfaces = CloudInitNetworkInterfaces{}.mapToSDK(params)
	if set || len(ci.NetworkInterfaces) > 0 {
		return &ci
	}
	return nil
}

func (ci CloudInit) Validate(version Version) error {
	if ci.Custom != nil {
		if err := ci.Custom.Validate(); err != nil {
			return err
		}
	}
	if ci.UpgradePackages != nil && *ci.UpgradePackages && version.Smaller(Version{Major: 8}) {
		return errors.New(CloudInit_Error_UpgradePackagesPre8)
	}
	return ci.NetworkInterfaces.Validate()
}

type CloudInitCustom struct {
	Meta    *CloudInitSnippet `json:"meta,omitempty"`
	Network *CloudInitSnippet `json:"network,omitempty"`
	User    *CloudInitSnippet `json:"user,omitempty"`
	Vendor  *CloudInitSnippet `json:"vendor,omitempty"`
}

func (config CloudInitCustom) mapToAPI(current *CloudInitCustom) string {
	var param string
	if current != nil { // update
		if config.Meta != nil {
			param += config.Meta.mapToAPI("meta")
		} else {
			param += current.Meta.mapToAPI("meta")
		}
		if config.Network != nil {
			param += config.Network.mapToAPI("network")
		} else {
			param += current.Network.mapToAPI("network")
		}
		if config.User != nil {
			param += config.User.mapToAPI("user")
		} else {
			param += current.User.mapToAPI("user")
		}
		if config.Vendor != nil {
			param += config.Vendor.mapToAPI("vendor")
		} else {
			param += current.Vendor.mapToAPI("vendor")
		}
	} else { // create
		if config.Meta != nil {
			param += config.Meta.mapToAPI("meta")
		}
		if config.Network != nil {
			param += config.Network.mapToAPI("network")
		}
		if config.User != nil {
			param += config.User.mapToAPI("user")
		}
		if config.Vendor != nil {
			param += config.Vendor.mapToAPI("vendor")
		}
	}
	if param != "" {
		return param[1:]
	}
	return ""
}

func (CloudInitCustom) mapToSDK(raw string) *CloudInitCustom {
	var set bool
	var config CloudInitCustom
	params := splitStringOfSettings(raw)
	if v, isSet := params["meta"]; isSet {
		config.Meta = CloudInitSnippet{}.mapToSDK(v)
		set = true
	}
	if v, isSet := params["network"]; isSet {
		config.Network = CloudInitSnippet{}.mapToSDK(v)
		set = true
	}
	if v, isSet := params["user"]; isSet {
		config.User = CloudInitSnippet{}.mapToSDK(v)
		set = true
	}
	if v, isSet := params["vendor"]; isSet {
		config.Vendor = CloudInitSnippet{}.mapToSDK(v)
		set = true
	}
	if set {
		return &config
	}
	return nil
}

func (ci CloudInitCustom) Validate() (err error) {
	if ci.Meta != nil {
		if err = ci.Meta.Validate(); err != nil {
			return
		}
	}
	if ci.Network != nil {
		if err = ci.Network.Validate(); err != nil {
			return
		}
	}
	if ci.User != nil {
		if err = ci.User.Validate(); err != nil {
			return err
		}
	}
	if ci.Vendor != nil {
		err = ci.Vendor.Validate()
	}
	return
}

func (ci CloudInitCustom) String() string {
	return ci.mapToAPI(nil)
}

type CloudInitIPv4Config struct {
	Address *IPv4CIDR    `json:"address,omitempty"`
	DHCP    bool         `json:"dhcp,omitempty"`
	Gateway *IPv4Address `json:"gateway,omitempty"`
}

const CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive string = "ipv4 dhcp is mutually exclusive with address"
const CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive string = "ipv4 dhcp is mutually exclusive with gateway"

func (config CloudInitIPv4Config) mapToAPI(current *CloudInitIPv4Config) string {
	// config can only be nil during update
	if config.DHCP {
		return ",ip=dhcp"
	}
	if current != nil { // Update phase, Update value
		var param string
		if config.Address != nil {
			if *config.Address != "" {
				param = ",ip=" + string(*config.Address)
			}
		} else if current.Address != nil {
			param = ",ip=" + string(*current.Address)
		}
		if config.Gateway != nil {
			if *config.Gateway != "" {
				param += ",gw=" + string(*config.Gateway)
			}
		} else if current.Gateway != nil {
			param += ",gw=" + string(*current.Gateway)
		}
		return param
	}
	// Create phase
	var param string
	if config.Address != nil && *config.Address != "" {
		param = ",ip=" + string(*config.Address)
	}
	if config.Gateway != nil && *config.Gateway != "" {
		param += ",gw=" + string(*config.Gateway)
	}
	return param
}

func (config CloudInitIPv4Config) String() string {
	param := config.mapToAPI(nil)
	if param != "" {
		return param[1:]
	}
	return ""
}

func (config CloudInitIPv4Config) Validate() error {
	if config.Address != nil && *config.Address != "" {
		if config.DHCP {
			return errors.New(CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive)
		}
		if err := config.Address.Validate(); err != nil {
			return err
		}
	}
	if config.Gateway != nil && *config.Gateway != "" {
		if config.DHCP {
			return errors.New(CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive)
		}
		if err := config.Gateway.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CloudInitIPv6Config struct {
	Address *IPv6CIDR    `json:"address,omitempty"`
	DHCP    bool         `json:"dhcp,omitempty"`
	Gateway *IPv6Address `json:"gateway,omitempty"`
	SLAAC   bool         `json:"slaac,omitempty"`
}

func (config CloudInitIPv6Config) mapToAPI(current *CloudInitIPv6Config) string {
	if config.DHCP {
		return ",ip6=dhcp"
	}
	if config.SLAAC {
		return ",ip6=auto"
	}
	if current != nil { // Update
		var param string
		if config.Address != nil {
			if *config.Address != "" {
				param = ",ip6=" + string(*config.Address)
			}
		} else if current.Address != nil {
			param = ",ip6=" + string(*current.Address)
		}
		if config.Gateway != nil {
			if *config.Gateway != "" {
				param += ",gw6=" + string(*config.Gateway)
			}
		} else if current.Gateway != nil {
			param += ",gw6=" + string(*current.Gateway)
		}
		return param
	}
	// create
	var param string
	if config.Address != nil && *config.Address != "" {
		param = ",ip6=" + string(*config.Address)
	}
	if config.Gateway != nil && *config.Gateway != "" {
		param += ",gw6=" + string(*config.Gateway)
	}
	return param
}

func (config CloudInitIPv6Config) String() string {
	param := config.mapToAPI(nil)
	if param != "" {
		return param[1:]
	}
	return ""
}

const CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive string = "ipv6 dhcp is mutually exclusive with address"
const CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive string = "ipv6 dhcp is mutually exclusive with gateway"
const CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive string = "ipv6 dhcp is mutually exclusive with slaac"
const CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive string = "ipv6 slaac is mutually exclusive with address"
const CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive string = "ipv6 slaac is mutually exclusive with gateway"

func (config CloudInitIPv6Config) Validate() error {
	if config.DHCP && config.SLAAC {
		return errors.New(CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive)
	}
	if config.Address != nil && *config.Address != "" {
		if config.DHCP {
			return errors.New(CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive)
		}
		if config.SLAAC {
			return errors.New(CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive)
		}
		if err := config.Address.Validate(); err != nil {
			return err
		}
	}
	if config.Gateway != nil && *config.Gateway != "" {
		if config.DHCP {
			return errors.New(CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive)
		}
		if config.SLAAC {
			return errors.New(CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive)
		}
		if err := config.Gateway.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CloudInitNetworkConfig struct {
	IPv4 *CloudInitIPv4Config `json:"ip4,omitempty"`
	IPv6 *CloudInitIPv6Config `json:"ip6,omitempty"`
}

func (config CloudInitNetworkConfig) mapToAPI(current *CloudInitNetworkConfig) (param string) {
	if current != nil { // Update
		if config.IPv4 != nil {
			param += config.IPv4.mapToAPI(current.IPv4)
		} else {
			if current.IPv4 != nil {
				param += current.IPv4.mapToAPI(nil)
			}
		}
		if config.IPv6 != nil {
			param += config.IPv6.mapToAPI(current.IPv6)
		} else {
			if current.IPv6 != nil {
				param += current.IPv6.mapToAPI(nil)
			}
		}
	} else { // Create
		if config.IPv4 != nil {
			param += config.IPv4.mapToAPI(nil)
		}
		if config.IPv6 != nil {
			param += config.IPv6.mapToAPI(nil)
		}
	}
	return
}

func (CloudInitNetworkConfig) mapToSDK(param string) (config CloudInitNetworkConfig) {
	params := splitStringOfSettings(param)
	var ipv4Set bool
	var ipv6Set bool
	var ipv4 CloudInitIPv4Config
	var ipv6 CloudInitIPv6Config
	if v, isSet := params["ip"]; isSet {
		ipv4Set = true
		if v == "dhcp" {
			ipv4.DHCP = true
		} else {
			tmp := IPv4CIDR(v)
			ipv4.Address = &tmp
		}
	}
	if v, isSet := params["gw"]; isSet {
		ipv4Set = true
		tmp := IPv4Address(v)
		ipv4.Gateway = &tmp
	}
	if v, isSet := params["ip6"]; isSet {
		ipv6Set = true
		switch v {
		case "dhcp":
			ipv6.DHCP = true
		case "auto":
			ipv6.SLAAC = true
		default:
			ipv6.Address = util.Pointer(IPv6CIDR(v))
		}
	}
	if v, isSet := params["gw6"]; isSet {
		ipv6Set = true
		tmp := IPv6Address(v)
		ipv6.Gateway = &tmp
	}
	if ipv4Set {
		config.IPv4 = &ipv4
	}
	if ipv6Set {
		config.IPv6 = &ipv6
	}
	return
}

func (config CloudInitNetworkConfig) Validate() (err error) {
	if config.IPv4 != nil {
		if err = config.IPv4.Validate(); err != nil {
			return
		}
	}
	if config.IPv6 != nil {
		err = config.IPv6.Validate()
	}
	return
}

type CloudInitNetworkInterfaces map[QemuNetworkInterfaceID]CloudInitNetworkConfig

func (interfaces CloudInitNetworkInterfaces) mapToAPI(current CloudInitNetworkInterfaces, params map[string]interface{}) (delete string) {
	for i, e := range interfaces {
		var tmpCurrent *CloudInitNetworkConfig
		if current != nil {
			if _, isSet := current[i]; isSet {
				tmp := current[i]
				tmpCurrent = &tmp
			}
		}
		param := e.mapToAPI(tmpCurrent)
		if param != "" {
			params["ipconfig"+strconv.FormatInt(int64(i), 10)] = param[1:]
		} else if tmpCurrent != nil {
			delete += ",ipconfig" + strconv.FormatInt(int64(i), 10)
		}
	}
	return
}

func (CloudInitNetworkInterfaces) mapToSDK(params map[string]interface{}) CloudInitNetworkInterfaces {
	ci := make(CloudInitNetworkInterfaces)
	for i := QemuNetworkInterfaceID(0); i < 32; i++ {
		if v, isSet := params["ipconfig"+strconv.FormatInt(int64(i), 10)]; isSet {
			tmp := v.(string)
			if len(tmp) > 1 { // can be "" or " "
				ci[i] = CloudInitNetworkConfig{}.mapToSDK(tmp)
			}
		}
	}
	return ci
}

func (interfaces CloudInitNetworkInterfaces) Validate() (err error) {
	for i := range interfaces {
		if err = i.Validate(); err != nil {
			return
		}
		if err = interfaces[i].Validate(); err != nil {
			return
		}
	}
	return
}

// If either Storage or FilePath is empty, the snippet will be removed
type CloudInitSnippet struct {
	FilePath CloudInitSnippetPath `json:"path,omitempty"`
	Storage  string               `json:"storage,omitempty"` // TODO custom type (storage)
}

func (ci CloudInitSnippet) mapToAPI(kind string) string {
	tmp := ci.String()
	if tmp != ":" {
		return "," + kind + "=" + tmp
	}
	return ""
}

func (CloudInitSnippet) mapToSDK(param string) *CloudInitSnippet {
	file := strings.SplitN(param, ":", 2)
	if len(file) == 2 {
		return &CloudInitSnippet{
			Storage:  file[0],
			FilePath: CloudInitSnippetPath(file[1])}
	}
	return nil
}

func (config CloudInitSnippet) String() string {
	return config.Storage + ":" + string(config.FilePath)
}

func (ci CloudInitSnippet) Validate() error {
	if ci.FilePath != "" {
		return ci.FilePath.Validate()
	}
	return nil
}

type CloudInitSnippetPath string

var (
	regexCloudInitSnippetPath_Charters = regexp.MustCompile(`^[a-zA-Z0-9- _\/.]+$`)
	regexCloudInitSnippetPath_Path     = regexp.MustCompile(`^[^,=/]+(\/[^,=/]+)*$`)
)

const (
	CloudInitSnippetPath_Error_Empty             = "cloudInitSnippetPath may not be empty"
	CloudInitSnippetPath_Error_InvalidCharacters = "cloudInitSnippetPath may ony contain the following characters: [a-zA-Z0-9_ -/.]"
	CloudInitSnippetPath_Error_InvalidPath       = "cloudInitSnippetPath must be a valid unix path"
	CloudInitSnippetPath_Error_MaxLength         = "cloudInitSnippetPath may not be longer than 256 characters"
	CloudInitSnippetPath_Error_Relative          = "cloudInitSnippetPath must be an relative path"
)

func (path CloudInitSnippetPath) Validate() error {
	if path == "" {
		return errors.New(CloudInitSnippetPath_Error_Empty)
	}
	if path[:1] == "/" {
		return errors.New(CloudInitSnippetPath_Error_Relative)
	}
	if len(path) > 256 {
		return errors.New(CloudInitSnippetPath_Error_MaxLength)
	}
	if !regexCloudInitSnippetPath_Charters.MatchString(string(path)) {
		return errors.New(CloudInitSnippetPath_Error_InvalidCharacters)
	}
	if !regexCloudInitSnippetPath_Path.MatchString(string(path)) {
		return errors.New(CloudInitSnippetPath_Error_InvalidPath)
	}
	return nil
}

type IPv4Address string

const IPv4Address_Error_Invalid = "ipv4Address is not a valid ipv6 address"

func (ip IPv4Address) Validate() error {
	if ip == "" {
		return nil
	}
	if net.ParseIP(string(ip)) == nil {
		return errors.New(IPv4Address_Error_Invalid)
	}
	if !isIPv4(string(ip)) {
		return errors.New(IPv4Address_Error_Invalid)
	}
	return nil
}

type IPv4CIDR string

const IPv4CIDR_Error_Invalid = "ipv4CIDR is not a valid ipv4 address"

func (cidr IPv4CIDR) Validate() error {
	if cidr == "" {
		return nil
	}
	ip, _, err := net.ParseCIDR(string(cidr))
	if err != nil {
		return errors.New(IPv4CIDR_Error_Invalid)
	}
	if !isIPv4(ip.String()) {
		return errors.New(IPv4CIDR_Error_Invalid)
	}
	return err
}

type IPv6Address string

const IPv6Address_Error_Invalid = "ipv6Address is not a valid ipv6 address"

func (ip IPv6Address) Validate() error {
	if ip == "" {
		return nil
	}
	if net.ParseIP(string(ip)) == nil {
		return errors.New(IPv6Address_Error_Invalid)
	}
	if !isIPv6(string(ip)) {
		return errors.New(IPv6Address_Error_Invalid)
	}
	return nil
}

type IPv6CIDR string

const IPv6CIDR_Error_Invalid = "ipv6CIDR is not a valid ipv6 address"

func (cidr IPv6CIDR) Validate() error {
	if cidr == "" {
		return nil
	}
	ip, _, err := net.ParseCIDR(string(cidr))
	if err != nil {
		return errors.New(IPv6CIDR_Error_Invalid)
	}
	if !isIPv6(ip.String()) {
		return errors.New(IPv6CIDR_Error_Invalid)
	}
	return nil
}
