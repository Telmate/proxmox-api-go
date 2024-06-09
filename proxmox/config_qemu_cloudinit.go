package proxmox

import (
	"crypto"
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
	PublicSSHkeys *[]crypto.PublicKey `json:"sshkeys"`
}

func (config CloudInit) mapToAPI(current *CloudInit, params map[string]interface{}) (delete string) {
	if current != nil { // Update
		if config.PublicSSHkeys != nil {
			if len(*config.PublicSSHkeys) > 0 {
				params["sshkeys"] = sshKeyUrlEncode(*config.PublicSSHkeys)
			} else {
				delete += ",sshkeys"
			}
		}
	} else { // Create
		if config.PublicSSHkeys != nil && len(*config.PublicSSHkeys) > 0 {
			params["sshkeys"] = sshKeyUrlEncode(*config.PublicSSHkeys)
		}
	}
	return
}

func (CloudInit) mapToSDK(params map[string]interface{}) *CloudInit {
	ci := CloudInit{}
	var set bool
	if v, isSet := params["sshkeys"]; isSet {
		tmp := sshKeyUrlDecode(v.(string))
		ci.PublicSSHkeys = &tmp
		set = true
	}
	if set {
		return &ci
	}
	return nil
}
