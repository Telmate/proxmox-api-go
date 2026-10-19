package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type lxcTestCaseGet struct {
	name   string
	input  rawConfigLXC
	pool   *PoolName
	state  PowerState
	output *ConfigLXC
}

type lxcTestGetFunc func() []lxcTestCaseGet

func (tests lxcTestGetFunc) Test(t *testing.T) {
	t.Helper()
	test := tests()
	reference := tests()
	for i := range test {
		t.Run(test[i].name, func(*testing.T) {
			require.Equal(t, test[i].output, test[i].input.Get(test[i].pool, test[i].state))
			require.Equal(t, reference[i].input, test[i].input, "mutated input")
		})
	}
}

func (tests lxcTestGetFunc) Inject(t *testing.T, testFunc func(*testing.T, RawConfigLXC, *ConfigLXC)) {
	t.Helper()
	test := tests()
	reference := tests()
	for i := range tests() {
		t.Run(test[i].name, func(*testing.T) {
			testFunc(t, &test[i].input, test[i].output)
			require.Equal(t, reference[i].input, test[i].input, "mutated input")
		})
	}
}
