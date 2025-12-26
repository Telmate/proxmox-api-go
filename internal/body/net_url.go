package body

const (
	backSlash   byte = '\\'
	singleQuote byte = '\''
)

// All Code below is copied from net/url
// See https://golang.org/src/net/url/url.go
// Changes have been made to make it compatible with Proxmox API requirements.

const upperHex = "0123456789ABCDEF"

func escape(s string, mode encoding) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			hexCount++
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	var buf [64]byte
	var t []byte

	required := len(s) + 2*hexCount
	if required <= len(buf) {
		t = buf[:required]
	} else {
		t = make([]byte, required)
	}

	if hexCount == 0 {
		copy(t, s)
		for i := 0; i < len(s); i++ {
			if s[i] == ' ' {
				t[i] = '+'
			}
		}
		return string(t)
	}

	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case shouldEscape(c, mode):
			t[j] = '%'
			t[j+1] = upperHex[c>>4]
			t[j+2] = upperHex[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// Return true if the specified character should be escaped
func shouldEscape(c byte, mode encoding) bool {
	// Unreserved characters (alphanumeric)
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '-', '_', '.', '~': // Unreserved characters (mark)
		return false
	case '&', '$':
		return mode == encodePveApiToken
	case '=', ':', '@', '+':
		return mode == encodePveApiToken || mode == encodePveQemuSshKey
	case singleQuote, '(', ')', '*', '!':
		return mode != encodePveApiToken
	}

	// Everything else must be escaped.
	return true
}
