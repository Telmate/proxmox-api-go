package proxmox

import (
	"crypto"
	"net/netip"
	"net/url"
	"regexp"
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
	DNS           *GuestDNS           `json:"dns"`
	PublicSSHkeys *[]crypto.PublicKey `json:"sshkeys"`
	UserPassword  *string             `json:"userpassword"` // TODO custom type
	Username      *string             `json:"username"`     // TODO custom type
}

func (config CloudInit) mapToAPI(current *CloudInit, params map[string]interface{}) (delete string) {
	if current != nil { // Update
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
	if config.UserPassword != nil {
		params["cipassword"] = *config.UserPassword
	}
	return
}

func (CloudInit) mapToSDK(params map[string]interface{}) *CloudInit {
	ci := CloudInit{}
	var set bool
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
	if set {
		return &ci
	}
	return nil
}
