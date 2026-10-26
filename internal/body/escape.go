package body

// Same as net/url.PathEscape()
func PathEscape(s string) string { return escape(s, encodePathSegment) }

func Escape(s string) string { return escape(s, encodePveApiToken) }

func QemuSshKeyEscape(s string) string { return escape(s, encodePveQemuSshKey) }

const Symbols = " !\"#$%&'()*+,-./:;<=>?@[\\]^`_{|}~"
const Alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRESTUVXYZ1234567890"

const (
	Colon     = "%3A"
	Comma     = "%2C"
	Equal     = "%3D"
	Newline   = "%0A"
	Plus      = "%2B"
	Semicolon = "%3B"
	Slash     = "%2F"
)

type encoding int

const (
	encodePathSegment   encoding = 1 << iota // "$&+-.:=@_~" as normal characters
	encodePveApiToken                        // "!'()*-._~" as normal characters
	encodePveQemuSshKey                      // "$&-._~" as normal characters
)

func (e encoding) String() string {
	switch e {
	case encodePathSegment:
		return "encodePathSegment"
	case encodePveApiToken:
		return "encodePveApiToken"
	case encodePveQemuSshKey:
		return "encodePveQemuSshKey"
	default:
		return "unknown encoding"
	}
}
