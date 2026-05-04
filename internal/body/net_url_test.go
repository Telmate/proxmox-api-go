package body

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_escape(b *testing.B) {
	tests := []encoding{
		encodePathSegment,
		encodePveApiToken,
		encodePveQemuSshKey,
	}
	input := "FB$sVO>K^PzAx/5ktCp?DXuhI!5QV(`[02q]BR]c}+uEU$FF'\\7sO6m2{KmRzn>zMOL=^=9CQT%9!rh8uf=dJ7|`pg6Yi*u7(JI\\2)<Z8^6DW?a73L;,-*C[clyfUgmhFl1IPHaZx/1*p88.zB8\\SG!OyDfoU_K:;e-[WYn%,&%0q6e'+5{t8k/iK+o,7/GHMP|7&l~<aCj<I;3)iP._&Od5#]~mLVEysg>GC<%bw[Aa\\7]9I_X4s5qngEouRc<Q6)<9VG'|&23K&gzZA:swpKe-s7w\\[m1R,3p~!zLineo(Pg||WHtDP/dA6%v>5x(1.#WMumZ[0{BYaHK#|AQ/kb{S~&Agm2\"!0c8=q6yf2d=lqlDGJX]H9D#'3%6:^|Q[LU4Hf_%JIUFdQ8Jp2*0$\"pib?L;~?j97Dv?L<3F*B7YBX96<8jmXHc8V)8=x*#l=giM,eDPYW6=eun>m\v8f@r~IK=\\.;nnC~TA*\"90&e\"A&K$g[CjN|K2C94WHW$cp[<hdQ{NWfedH>)8KQ>{7E,r!RAwu.2Dhyl*,cK+!;fc]{\\ciR6\\z5cJt4ixUGTg7[x~QNT3BT|\">cGQk1SH:17yFAhPfzP=]lb-i6t`'N>}u#m<Dv_?oqDr19&UAT^pAT|Bld<ZA2)0$Eu'mI\\UB!6[oR>:R|Phn7q0d#3x((DYhmb`XJ?v6$v{if?:nU)R5o5;hmFU?6)cGW9\"-MTfGt&Fn`)VAII~b042W%qL^QjfL|W`:BTGSCqwypHh>ZfNJ{Qq}FT07YzvPuRc}?]SJ0x3q+Bn#s&n&]q'wj~iHZ_v)rGlTK}|~!=\\62DGbj',R^*3MTbv]f*u}&\")BhRN-]n.sRE!}[:Y&1#ugHJ\"LwnT5e:WFTE-5Id;KVgdeqiR2N\"OMIPLcqm\\Sn%c%=\"<x}Ehb)WNbGmsk4Kq@p1\"_I\"T#cJE]CWWw)%Q4r4@Z=VZ\\CGVT3's>ur:DpS]tqwch!Jx^$V|iCXE!sZ~FUUP,]9J%gZhsr.x.,yWJ'zploNXz[%"
	b.ResetTimer()
	for b.Loop() {
		for i := range tests {
			escape(input, tests[i])
		}
	}
}

func Test_escape(t *testing.T) {
	t.Parallel()
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
