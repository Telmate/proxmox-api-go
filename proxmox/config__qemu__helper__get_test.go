package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type qemuTestCaseGet struct {
	name   string
	input  map[string]any
	vmr    VmRef
	output *ConfigQemu
	err    error
}

type qemuTestGetFunc func() []qemuTestCaseGet

func (tests qemuTestGetFunc) Test(t *testing.T) {
	t.Helper()
	test := tests()
	reference := tests()
	for i := range test {
		t.Run(test[i].name, func(*testing.T) {
			output, err := (&rawConfigQemu{a: test[i].input}).get(test[i].vmr)
			if err != nil {
				require.Equal(t, test[i].err, err)
			} else {
				require.Equal(t, test[i].output, output)
			}
			require.Equal(t, reference[i].input, test[i].input, "mutated input")
		})
	}
}

func (tests qemuTestGetFunc) Inject(t *testing.T, testFunc func(*testing.T, RawConfigQemu, *ConfigQemu, error)) {
	t.Helper()
	test := tests()
	reference := tests()
	for i := range tests() {
		t.Run(test[i].name, func(*testing.T) {
			testFunc(t, &rawConfigQemu{a: test[i].input}, test[i].output, test[i].err)
			require.Equal(t, reference[i].input, test[i].input, "mutated input")
		})
	}
}
