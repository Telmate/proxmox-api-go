package body

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_escape(t *testing.T) {
	tests := map[encoding]map[byte]string{
		encodePathSegment: {
			' ':         "%20",
			'!':         "%21",
			'"':         "%22",
			'#':         "%23",
			'$':         "$",
			'%':         "%25",
			'&':         "&",
			'(':         "%28",
			')':         "%29",
			'*':         "%2A",
			'+':         "+",
			',':         "%2C",
			'-':         "-",
			'.':         ".",
			'/':         "%2F",
			':':         ":",
			';':         "%3B",
			'<':         "%3C",
			'=':         "=",
			'>':         "%3E",
			'?':         "%3F",
			'@':         "@",
			'[':         "%5B",
			']':         "%5D",
			'^':         "%5E",
			'_':         "_",
			'`':         "%60",
			'{':         "%7B",
			'|':         "%7C",
			'}':         "%7D",
			'~':         "~",
			backSlash:   "%5C",
			singleQuote: "%27",
		},
		encodePveApiToken: {
			' ':         "%20",
			'!':         "!",
			'"':         "%22",
			'#':         "%23",
			'$':         "%24",
			'%':         "%25",
			'&':         "%26",
			'(':         "(",
			')':         ")",
			'*':         "*",
			'+':         "%2B",
			',':         "%2C",
			'-':         "-",
			'.':         ".",
			'/':         "%2F",
			':':         "%3A",
			';':         "%3B",
			'<':         "%3C",
			'=':         "%3D",
			'>':         "%3E",
			'?':         "%3F",
			'@':         "%40",
			'[':         "%5B",
			']':         "%5D",
			'^':         "%5E",
			'_':         "_",
			'`':         "%60",
			'{':         "%7B",
			'|':         "%7C",
			'}':         "%7D",
			'~':         "~",
			backSlash:   "%5C",
			singleQuote: "'",
		},
		encodePveQemuSshKey: {
			' ':         "%20",
			'!':         "%21",
			'"':         "%22",
			'#':         "%23",
			'$':         "$",
			'%':         "%25",
			'&':         "&",
			'(':         "%28",
			')':         "%29",
			'*':         "%2A",
			'+':         "%2B",
			',':         "%2C",
			'-':         "-",
			'.':         ".",
			'/':         "%2F",
			':':         "%3A",
			';':         "%3B",
			'<':         "%3C",
			'=':         "%3D",
			'>':         "%3E",
			'?':         "%3F",
			'@':         "%40",
			'[':         "%5B",
			']':         "%5D",
			'^':         "%5E",
			'_':         "_",
			'`':         "%60",
			'{':         "%7B",
			'|':         "%7C",
			'}':         "%7D",
			'~':         "~",
			backSlash:   "%5C",
			singleQuote: "%27",
		},
	}
	for mode, test := range tests {
		name := mode.String() + "/"
		for input, output := range test {
			var runName string
			switch input {
			case ' ':
				runName = name + "space"
			case '/':
				runName = name + "slash"
			default:
				runName = name + string(input)
			}
			t.Run(runName, func(*testing.T) {
				require.Equal(t, output, escape(string(input), mode))
			})
		}
	}
}
