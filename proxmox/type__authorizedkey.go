package proxmox

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"golang.org/x/crypto/ssh"
)

type AuthorizedKey struct {
	Options   []string
	PublicKey ssh.PublicKey
	Comment   string
}

const (
	AuthorizedKey_Error_NilPointer = "authorizedKey pointer is nil"
	AuthorizedKey_Error_Invalid    = "invalid value for AuthorizedKey"
)

// Parse parses a public key from an authorized_keys file used in OpenSSH according to the sshd(8) manual page.
func (key *AuthorizedKey) Parse(rawKey []byte) error {
	if key == nil {
		return errors.New(AuthorizedKey_Error_NilPointer)
	}
	return key.parse_unsafe(rawKey)
}

// Parse the raw key into the AuthorizedKey struct
func (key *AuthorizedKey) parse_unsafe(rawKey []byte) (err error) {
	key.PublicKey, key.Comment, key.Options, _, err = ssh.ParseAuthorizedKey(rawKey)
	return err
}

// Custom MarshalJSON for AuthorizedKey
func (key AuthorizedKey) MarshalJSON() ([]byte, error) {
	// Convert the AuthorizedKey to the OpenSSH format and return it as a JSON string
	return json.Marshal(key.String())
}

func (key AuthorizedKey) String() string { // String is for fmt.Stringer.
	if key.PublicKey == nil {
		return ""
	}
	var options string
	if len(key.Options) > 0 {
		options = strings.Join(key.Options, ",") + " "
	}
	tmpKey := string(ssh.MarshalAuthorizedKey(key.PublicKey))
	if key.Comment == "" {
		return options + tmpKey[:len(tmpKey)-1]
	}
	comment := regexMultipleSpaces.ReplaceAllString(key.Comment, " ")
	if comment == " " {
		return options + tmpKey[:len(tmpKey)-1]
	}
	return options + tmpKey[:len(tmpKey)-1] + " " + comment
}

// Custom UnmarshalJSON for AuthorizedKey
func (key *AuthorizedKey) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return errors.New(AuthorizedKey_Error_Invalid)
	}
	// Decode JSON string and handle unescaping
	rawString := string(data[1 : len(data)-1]) // Strip surrounding quotes
	unescapedString, _ := strconv.Unquote(`"` + rawString + `"`)
	return key.parse_unsafe([]byte(unescapedString))
}

const newlineEncoded = "%0A"
const spaceEncoded = "%20"

var regexMultipleNewlineEncoded = regexp.MustCompile(`(%0A)+`)
var regexMultipleSpaces = regexp.MustCompile(`( )+`)
var regexMultipleSpacesEncoded = regexp.MustCompile(`(%20)+`)

// URL encodes the ssh keys
func sshKeyUrlDecode(encodedKeys string) (keys []AuthorizedKey) {
	encodedKeys = regexMultipleSpacesEncoded.ReplaceAllString(encodedKeys, spaceEncoded)
	encodedKeys = strings.TrimSuffix(encodedKeys, newlineEncoded)
	encodedKeys = regexMultipleNewlineEncoded.ReplaceAllString(encodedKeys, newlineEncoded)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%2B`, `+`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%40`, `@`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%3D`, `=`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%3A`, `:`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%20`, ` `)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%2F`, `/`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%2C`, `,`)
	encodedKeys = strings.ReplaceAll(encodedKeys, `%22`, `"`)
	rawKeys := strings.Split(encodedKeys, newlineEncoded)
	keys = make([]AuthorizedKey, len(rawKeys))
	for i := range rawKeys {
		keys[i].PublicKey, keys[i].Comment, keys[i].Options, _, _ = ssh.ParseAuthorizedKey([]byte(rawKeys[i]))
	}
	return
}

// URL encodes the ssh keys
func sshKeyUrlEncode(keys []AuthorizedKey) string {
	encodedKeys := make([]string, len(keys))
	for i := range keys {
		tmpKey := keys[i].String()
		if tmpKey == "" {
			continue
		}
		encodedKeys[i] = body.QemuSshKeyEscape(tmpKey) + newlineEncoded
	}
	return strings.Join(encodedKeys, "")
}
