package proxmox

import (
	"crypto"
	"errors"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var regexMultipleSpaces = regexp.MustCompile(`\s+`)
var regexMultipleSpacesEncoded = regexp.MustCompile(`(%20)+`)
var regexMultipleNewlineEncoded = regexp.MustCompile(`(%0A)+`)

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
	Custom            *CloudInitCustom           `json:"cicustom"`
	DNS               *GuestDNS                  `json:"dns"`
	NetworkInterfaces CloudInitNetworkInterfaces `json:"ipconfig"`
	PublicSSHkeys     *[]crypto.PublicKey        `json:"sshkeys"`
	UserPassword      *string                    `json:"userpassword"` // TODO custom type
	Username          *string                    `json:"username"`     // TODO custom type
}

func (config CloudInit) mapToAPI(current *CloudInit, params map[string]interface{}) (delete string) {
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
		if config.PublicSSHkeys != nil && len(*config.PublicSSHkeys) > 0 {
			params["sshkeys"] = sshKeyUrlEncode(*config.PublicSSHkeys)
		}
	}
	// Shared
	config.NetworkInterfaces.mapToAPI(params)
	if config.UserPassword != nil {
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
		tmp := v.(string)
		ci.UserPassword = &tmp
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
		tmp := sshKeyUrlDecode(v.(string))
		ci.PublicSSHkeys = &tmp
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

func (ci CloudInit) Validate() error {
	if ci.Custom != nil {
		if err := ci.Custom.Validate(); err != nil {
			return err
		}
	}
	return ci.NetworkInterfaces.Validate()
}

type CloudInitCustom struct {
	Meta    *CloudInitSnippet `json:"meta"`
	Network *CloudInitSnippet `json:"network"`
	User    *CloudInitSnippet `json:"user"`
	Vendor  *CloudInitSnippet `json:"vendor"`
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
		config.Meta = CloudInitSnippet{}.mapToSDK(v.(string))
		set = true
	}
	if v, isSet := params["network"]; isSet {
		config.Network = CloudInitSnippet{}.mapToSDK(v.(string))
		set = true
	}
	if v, isSet := params["user"]; isSet {
		config.User = CloudInitSnippet{}.mapToSDK(v.(string))
		set = true
	}
	if v, isSet := params["vendor"]; isSet {
		config.Vendor = CloudInitSnippet{}.mapToSDK(v.(string))
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

type CloudInitNetworkInterfaces map[QemuNetworkInterfaceID]string // TODO string should be a custom type

func (interfaces CloudInitNetworkInterfaces) mapToAPI(params map[string]interface{}) {
	for i, e := range interfaces {
		params["ipconfig"+strconv.FormatInt(int64(i), 10)] = e
	}
}

func (CloudInitNetworkInterfaces) mapToSDK(params map[string]interface{}) CloudInitNetworkInterfaces {
	ci := make(CloudInitNetworkInterfaces)
	for i := QemuNetworkInterfaceID(0); i < 32; i++ {
		if v, isSet := params["ipconfig"+strconv.FormatInt(int64(i), 10)]; isSet {
			tmp := v.(string)
			if len(tmp) > 1 { // can be "" or " "
				ci[i] = v.(string)
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
	}
	return
}

// If either Storage or FilePath is empty, the snippet will be removed
type CloudInitSnippet struct {
	Storage  string               `json:"storage"` // TODO custom type (storage)
	FilePath CloudInitSnippetPath `json:"path"`
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
