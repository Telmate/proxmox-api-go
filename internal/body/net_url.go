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
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c, mode) {
			hexCount++
		}
	}

	if hexCount == 0 {
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
func shouldEscape(c byte, mode encoding) bool { return table[c]&mode == 0 }

var table = [256]encoding{
	'!':  encodePveApiToken,
	'$':  encodePathSegment | encodePveQemuSshKey,
	'&':  encodePathSegment | encodePveQemuSshKey,
	'\'': encodePveApiToken,
	'(':  encodePveApiToken,
	')':  encodePveApiToken,
	'*':  encodePveApiToken,
	'+':  encodePathSegment,
	'-':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'.':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'0':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'1':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'2':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'3':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'4':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'5':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'6':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'7':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'8':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'9':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	':':  encodePathSegment,
	'=':  encodePathSegment,
	'@':  encodePathSegment,
	'A':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'B':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'C':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'D':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'E':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'F':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'G':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'H':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'I':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'J':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'K':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'L':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'M':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'N':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'O':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'P':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'Q':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'R':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'S':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'T':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'U':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'V':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'W':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'X':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'Y':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'Z':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'_':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'a':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'b':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'c':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'd':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'e':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'f':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'g':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'h':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'i':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'j':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'k':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'l':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'm':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'n':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'o':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'p':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'q':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'r':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	's':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	't':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'u':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'v':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'w':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'x':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'y':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'z':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
	'~':  encodePathSegment | encodePveApiToken | encodePveQemuSshKey,
}
