package body

// Same as net/url.PathEscape()
func PathEscape(s string) string { return escape(s, encodePathSegment) }

func Escape(s string) string { return escape(s, encodePveApiToken) }

func QemuSshKeyEscape(s string) string { return escape(s, encodePveQemuSshKey) }

const Symbols = " !\"#$%&'()*+,-./:;<=>?@[\\]^`_{|}~"

type encoding int

const (
	encodePathSegment   encoding = 1 + iota // "$&+-.:=@_~" as normal characters
	encodePveApiToken                       // "!'()*-._~" as normal characters
	encodePveQemuSshKey                     // "$&-._~" as normal characters
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
